//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"os"
	"text/template"
)

type reportDesc struct {
	ID string
	Desc string
	Parts []string
}

const (
	P = "protocol_version"
	UID = "unique_id"
	DN = "device_name"
	ST = "send_time"
	CN = "count_number"
)

var Reports = []reportDesc{
	{"GTPNA", "Power on report", []string{P, UID, DN, ST, CN}},
	{"GTPFA", "Power off report", []string{P, UID, DN, ST, CN}},
	// EPN
	// EPF
	// BPL
	// BTC
	// STC
	{"GTSTT", "Device motion state indication report", []string{P, UID, DN, "state", "gps_accuracy",
		"speed", "azimuth", "altitude", "last_longitude", "last_latitude", "gps_utc_time", "mcc", "mnc", "lac",
		"cell_id", "odo_mileage", ST, CN}},
	{"GTPDP", "GPRS PDP connection report", []string{P, UID, DN, ST, CN}},
	// SWG
	// IGN
	// IGF
	// GSM
	// TSM
	// UPC
	// JDR
	// JDS
	{"GTINF", "Device information report", []string{P, UID, DN, "state", "iccid",
		"csq_rssi", "csq_ber", "ext_power_supply", "mileage", "reserved", "battery_voltage", "charging", "led_on",
		"gps_on_need", "gps_antenna_type", "gps_antenna_state", "last_gps_fix_utc_time", "battery_percentage",
		"flash_type", "temperature", "reserved2", "reserved3", ST, CN}},

	// These below are the same report sent for different reasons
	{"GTPNL", "First location after power on report", []string{P, UID, DN,
		"report_id", "report_type", "number", "gps_accuracy", "speed",
		"azimuth", "altitude", "longitude", "latitude", "gps_utc_time", "mcc", "mnc", "lac", "cell_id",
		"odo_mileage", "battery_percentage", ST, CN}},
	{"GTSOS", "SOS function key report", []string{P, UID, DN,
		"report_id", "report_type", "number", "gps_accuracy", "speed",
		"azimuth", "altitude", "longitude", "latitude", "gps_utc_time", "mcc", "mnc", "lac", "cell_id",
		"odo_mileage", "battery_percentage", ST, CN}},

	// This is a special case where more than 1 gps point can be returned
	{"GTFRI", "Report of AT+GTFRI", []string{P, UID, DN,
		"report_id", "report_type", "number", "gps_accuracy", "speed",
		"azimuth", "altitude", "longitude", "latitude", "gps_utc_time", "mcc", "mnc", "lac", "cell_id",
		"odo_mileage", "battery_percentage", ST, CN}},
}

const Header =`package parser

import (
    "strings"
    "errors"
)

type PacketDetails struct {
	Type string
	ID string
	Desc string
	Parts map[string]string
}

var InvalidContentErr = errors.New("payload does not contain enough parts")
var InvalidPreambleErr = errors.New("payload does not contain + and $")
var NotImplementedErr = errors.New("packet type not implemented")

`

const DecodeFuncBody = `func [[.ID]]ReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{[[range .Parts]]"[[.]]",[[end]]}

	parts := strings.Split(payload, ",")
	if len(parts) != [[len .Parts]] {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"[[.ID]]",
			"[[.Desc]]",
			resultMap,
		},
    }, nil
}

`

const DecodeFactory = `func DecodePacket(payload string) ([]PacketDetails, error) {
	if len(payload) < 2 || payload[0] != '+' || payload[len(payload) - 1] != '$' {
		return []PacketDetails{}, InvalidPreambleErr
	}

	payload = strings.Trim(payload, "+$")
	parts := strings.SplitN(payload, ",", 2)
	if len(parts) != 2 {
		return []PacketDetails{}, InvalidContentErr
	}

	payload = parts[1]
	typeIdParts := strings.Split(parts[0], ":")
	if len(typeIdParts) != 2 {
		return []PacketDetails{}, InvalidContentErr
	}

	packetType := typeIdParts[0]
	packetAction := typeIdParts[1]

	switch packetType {
	case "ACK":
		switch packetAction {

		default:
			return []PacketDetails{}, NotImplementedErr
		}

	case "RESP":
		switch packetAction {
		[[range .]]case "[[.ID]]":
			return [[.ID]]ReportDecode(payload)
        [[end]]
		default:
			return []PacketDetails{}, NotImplementedErr
		}

	default:
		return []PacketDetails{}, NotImplementedErr
	}
}

`

func main() {
	buf := bytes.NewBuffer(nil)
	_, err := buf.WriteString(Header)
	if err != nil {
		panic(err)
	}

	reportDecodeFuncTemplate, err := template.New("a").Delims("[[", "]]").Parse(DecodeFuncBody)
	if err != nil {
		panic(err)
	}
	decodeFuncTemplate, err := template.New("b").Delims("[[", "]]").Parse(DecodeFactory)
	if err != nil {
		panic(err)
	}

	for _, reportDescriptor := range Reports {
		// Skip this one as its defined manually
		if reportDescriptor.ID == "GTFRI" {
			continue
		}

		err = reportDecodeFuncTemplate.Execute(buf, reportDescriptor)
		if err != nil {
			panic(err)
		}
	}

	err = decodeFuncTemplate.Execute(buf, Reports)
	if err != nil {
		panic(err)
	}

	f, err := os.Create("generated.go")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf.WriteTo(f)
}
