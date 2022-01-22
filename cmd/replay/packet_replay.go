package main

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/rs/zerolog/log"
	"github.com/terrycain/gl300w/pkg/parser"
)

func main() {
	log.Info().Msg("Packet reader")

	handle, err := pcap.OpenOffline("gl300w.pcapng")
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		payload := packet.ApplicationLayer().Payload()
		if len(payload) < 3 || payload[0] != '+' || payload[len(payload)-1] != '$' {
			continue
		}

		data, err := parser.Decode(payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse packet")
			continue
		}

		log.Info().Interface("pkt", data).Msg("")
	}
}
