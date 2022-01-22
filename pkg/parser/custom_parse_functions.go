package parser

import (
	"strconv"
	"strings"
)

func GTFRIReportDecode(payload string) ([]PacketDetails, error) {
	firstFields := []string{"protocol_version","unique_id","device_name","report_id","report_type","number"}
	repeatedFields := []string{"gps_accuracy","speed","azimuth","altitude","longitude","latitude","gps_utc_time","mcc","mnc","lac","cell_id","odo_mileage"}
	trailingFields := []string{"battery_percentage","send_time","count_number"}

	parts := strings.Split(payload, ",")

	// 21 is the minimum number of fields it can be
	if len(parts) < 21 {
		return []PacketDetails{}, InvalidContentErr
	}
	// 9 constant fields + number*12 fields
	entries, err := strconv.Atoi(parts[5]) // Number field
	if err != nil {
		return []PacketDetails{}, InvalidContentErr
	}
	if len(parts) != (entries * 12) + 9 {
		return []PacketDetails{}, InvalidContentErr
	}

	result := make([]PacketDetails, 0)
	trailOffset := (entries * 12) + 6
	for i := 0; i < entries; i++ {
		resultMap := make(map[string]string)

		// Read first fields
		for index, fieldName := range firstFields {
			resultMap[fieldName] = parts[index]
		}

		offset := 6 + (12 * i)
		for index, fieldName := range repeatedFields {
			resultMap[fieldName] = parts[offset + index]
		}

		// Read trail
		for index, fieldName := range trailingFields {
			resultMap[fieldName] = parts[trailOffset + index]
		}

		result = append(result, PacketDetails{
			"RESP",
			"GTFRI",
			"Report of AT+GTFRI",
			resultMap,
		})
	}

	return result, nil
}
