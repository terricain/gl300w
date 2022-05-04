package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/rs/zerolog/log"
	kafka "github.com/segmentio/kafka-go"
	"github.com/terrycain/gl300w/pkg/parser"
	"io/ioutil"
	"net"
	"os"
)

var portFlag = flag.Int("port", 9090, "UDP Port to listen on")

func main() {
	log.Info().Msgf("Listening for UDP on %d", *portFlag)


	// listen to incoming udp packets
	pc, err := net.ListenPacket("udp", fmt.Sprintf("0.0.0.0:%d", *portFlag))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to bind to port")
	}
	defer pc.Close()

	kafkaURL := os.Getenv("KAFKA_URL")
	caCert := os.Getenv("CA_CERT")
	clientCert := os.Getenv("CLIENT_CERT")
	clientKey := os.Getenv("CLIENT_KEY")

	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load client cert")
	}

	// Load CA cert
	caCertData, err := ioutil.ReadFile(caCert)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load CA cert")
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertData)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	client := kafka.Client{
		Addr: kafka.TCP(kafkaURL),
		Transport: &kafka.Transport{TLS: tlsConfig},
	}

	if resp, err := client.CreateTopics(context.Background(), &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{
			{
				Topic: "gl300.raw",
				NumPartitions: 1,
				ReplicationFactor: 1,
			},
		},
	}); err != nil {
		log.Fatal().Interface("error_info", resp).Err(err).Msg("Failed to create topics")
	} else {
		for topicName, err := range resp.Errors {
			if errors.Is(err, kafka.TopicAlreadyExists) {
				continue
			}
			if err != nil {
				log.Fatal().Str("topic", topicName).Err(err).Msg("Got topic specific error")
			}
		}
	}

	rawWriter := &kafka.Writer{
		Addr: kafka.TCP(kafkaURL),
		Topic:   "gl300.raw",
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{TLS: tlsConfig},
	}

	defer rawWriter.Close()

	for {
		buf := make([]byte, 4096)  // This is larger than any UDP packet it will ever send.
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			log.Error().Err(err).Msg("Caught error when reading from buffer")
			continue
		}
		go serve(addr, rawWriter, buf[:n])
	}
}

func serve(addr net.Addr, rawWriter *kafka.Writer, payload []byte) {
	data, err := parser.Decode(payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse packet")
		return
	}

	for _, item := range data {
		imsi, found := item.Params["unique_id"]
		if !found {
			continue
		}

		jsonData, err := json.Marshal(item)
		if err != nil {
			log.Error().Err(err).Msg("failed to convert packet to json")
			return
		}

		msg := kafka.Message{
			Key:   []byte(imsi),
			Value: jsonData,
		}
		if err = rawWriter.WriteMessages(context.Background(), msg); err != nil {
			log.Error().Err(err).Msg("Failed to send to kafka")
		} else {
			log.Info().Str("imsi", imsi).Msg("written event to kafka")
		}
	}
}