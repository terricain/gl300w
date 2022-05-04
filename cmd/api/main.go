package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/terrycain/gl300w/pkg/parser"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

var portFlag = flag.Int("port", 9091, "HTTP Port to listen on")

type Details struct {
	BatteryPercentage string
	GPSTime string
	Lat string
	Lon string

}
var clients = make(map[string]Details)
var clientsMutex = sync.RWMutex{}

func main() {
	log.Info().Msgf("Running API on %d", *portFlag)
	gin.SetMode(gin.ReleaseMode)

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

	dialer := kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
		TLS: tlsConfig,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaURL},
		Topic:     "gl300.raw",
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
		Dialer: &dialer,
	})
	defer reader.Close()

	go func() {
		log.Info().Msg("Starting kafka consumer")
		for {
			m, err := reader.ReadMessage(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("Failed to read message from kafka")
				time.Sleep(5 * time.Second)
				continue
			}
			handleMessage(m)

		}
	}()


	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.File("cmd/api/index.html")
	})
	r.GET("/api", func(c *gin.Context) {
		clientsMutex.RLock()
		defer clientsMutex.RUnlock()
		c.JSON(200, clients)
	})

	r.Run(fmt.Sprintf("0.0.0.0:%d", *portFlag))
}

func handleMessage(msg kafka.Message) {
	imsi := string(msg.Key)
	var result parser.Packet
	if err := json.Unmarshal(msg.Value, &result); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal data")
		return
	}
	log.Info().Int("partition", msg.Partition).Int64("offset", msg.Offset).Str("imsi", imsi).Str("action_id", result.ActionID).Msg("Read value")

	if result.ActionID != "GTFRI" {
		return
	}

	d := Details{
		BatteryPercentage: result.Params["battery_percentage"],
		GPSTime:           result.Params["gps_utc_time"],
		Lat:               result.Params["latitude"],
		Lon:               result.Params["longitude"],
	}
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	clients[imsi] = d
}