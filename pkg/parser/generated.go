package parser

import (
	"errors"
	"strings"
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

func GTPNAReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{"protocol_version","unique_id","device_name","send_time","count_number",}

	parts := strings.Split(payload, ",")
	if len(parts) != 5 {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"GTPNA",
			"Power on report",
			resultMap,
		},
    }, nil
}

func GTPFAReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{"protocol_version","unique_id","device_name","send_time","count_number",}

	parts := strings.Split(payload, ",")
	if len(parts) != 5 {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"GTPFA",
			"Power off report",
			resultMap,
		},
    }, nil
}

func GTSTTReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{"protocol_version","unique_id","device_name","state","gps_accuracy","speed","azimuth","altitude","last_longitude","last_latitude","gps_utc_time","mcc","mnc","lac","cell_id","odo_mileage","send_time","count_number",}

	parts := strings.Split(payload, ",")
	if len(parts) != 18 {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"GTSTT",
			"Device motion state indication report",
			resultMap,
		},
    }, nil
}

func GTPDPReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{"protocol_version","unique_id","device_name","send_time","count_number",}

	parts := strings.Split(payload, ",")
	if len(parts) != 5 {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"GTPDP",
			"GPRS PDP connection report",
			resultMap,
		},
    }, nil
}

func GTINFReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{"protocol_version","unique_id","device_name","state","iccid","csq_rssi","csq_ber","ext_power_supply","mileage","reserved","battery_voltage","charging","led_on","gps_on_need","gps_antenna_type","gps_antenna_state","last_gps_fix_utc_time","battery_percentage","flash_type","temperature","reserved2","reserved3","send_time","count_number",}

	parts := strings.Split(payload, ",")
	if len(parts) != 24 {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"GTINF",
			"Device information report",
			resultMap,
		},
    }, nil
}

func GTPNLReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{"protocol_version","unique_id","device_name","report_id","report_type","number","gps_accuracy","speed","azimuth","altitude","longitude","latitude","gps_utc_time","mcc","mnc","lac","cell_id","odo_mileage","battery_percentage","send_time","count_number",}

	parts := strings.Split(payload, ",")
	if len(parts) != 21 {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"GTPNL",
			"First location after power on report",
			resultMap,
		},
    }, nil
}

func GTSOSReportDecode(payload string) ([]PacketDetails, error) {
	resultMap := make(map[string]string)
	fields := []string{"protocol_version","unique_id","device_name","report_id","report_type","number","gps_accuracy","speed","azimuth","altitude","longitude","latitude","gps_utc_time","mcc","mnc","lac","cell_id","odo_mileage","battery_percentage","send_time","count_number",}

	parts := strings.Split(payload, ",")
	if len(parts) != 21 {
		return []PacketDetails{}, InvalidContentErr
	}

	for index, fieldName := range fields {
		resultMap[fieldName] = parts[index]
	}

	return []PacketDetails{
		{
			"RESP",
			"GTSOS",
			"SOS function key report",
			resultMap,
		},
    }, nil
}

func DecodePacket(payload string) ([]PacketDetails, error) {
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
		case "GTPNA":
			return GTPNAReportDecode(payload)
        case "GTPFA":
			return GTPFAReportDecode(payload)
        case "GTSTT":
			return GTSTTReportDecode(payload)
        case "GTPDP":
			return GTPDPReportDecode(payload)
        case "GTINF":
			return GTINFReportDecode(payload)
        case "GTPNL":
			return GTPNLReportDecode(payload)
        case "GTSOS":
			return GTSOSReportDecode(payload)
        case "GTFRI":
			return GTFRIReportDecode(payload)
        
		default:
			return []PacketDetails{}, NotImplementedErr
		}

	default:
		return []PacketDetails{}, NotImplementedErr
	}
}

