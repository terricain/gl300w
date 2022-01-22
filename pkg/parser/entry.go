package parser

//go:generate go run gen.go

type Packet struct {
	Raw string // Raw string
	PacketType string `json:"type"` // Either command, acknowledgement or report
	ActionID string `json:"id"` // GTXXX
	ActionDescription string // Friendly name of description

	Params map[string]string `json:"data"`
	Valid bool
}

func Decode(packet []byte) ([]Packet, error) {
	payload := string(packet)

	data, err := DecodePacket(payload)
	result := make([]Packet, len(data))
	for i, pktInfo := range data {
		result[i].PacketType = pktInfo.Type
		result[i].ActionID = pktInfo.ID
		result[i].ActionDescription = pktInfo.Desc
		result[i].Params = pktInfo.Parts
		result[i].Valid = err == nil
	}

	return result, err
}