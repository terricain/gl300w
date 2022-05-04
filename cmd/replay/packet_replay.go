package main

import (
	"flag"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

var portFlag = flag.Int("port", 9090, "UDP Port to send to")

func main() {
	log.Info().Msg("Packet reader")

	handle, err := pcap.OpenOffline("gl300w.pcapng")
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("udp", fmt.Sprintf("127.0.0.1:%d", *portFlag))
	if err != nil {
		fmt.Printf("Some error %v", err)
		return
	}
	defer conn.Close()

	var lastTimeStamp = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		if lastTimeStamp.Year() == 1970 {
			lastTimeStamp = packet.Metadata().Timestamp
		} else if packet.Metadata().Timestamp.Sub(lastTimeStamp) > 0 {
			time.Sleep(packet.Metadata().Timestamp.Sub(lastTimeStamp))
			lastTimeStamp = packet.Metadata().Timestamp
		}

		payload := packet.ApplicationLayer().Payload()
		if len(payload) < 3 || payload[0] != '+' || payload[len(payload)-1] != '$' {
			continue
		}

		fmt.Println(string(payload))
		_, _ = conn.Write(payload)
	}


}
