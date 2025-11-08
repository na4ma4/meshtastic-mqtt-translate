package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/relay"
)

func main() {
	// Configuration flags
	sourceBroker := flag.String("source-broker", "tcp://localhost:1883", "Source MQTT broker URL")
	sourceClientID := flag.String("source-client-id", "meshtastic-relay-source", "Source MQTT client ID")
	sourceTopic := flag.String("source-topic", "msh/+/json/+/+", "Source MQTT topic to subscribe to")

	destBroker := flag.String("dest-broker", "tcp://localhost:1884", "Destination MQTT broker URL")
	destClientID := flag.String("dest-client-id", "meshtastic-relay-dest", "Destination MQTT client ID")
	destTopic := flag.String("dest-topic", "meshtastic/json", "Destination MQTT topic prefix")

	username := flag.String("username", "", "MQTT username (optional)")
	password := flag.String("password", "", "MQTT password (optional)")

	flag.Parse()

	config := relay.Config{
		SourceBroker:   *sourceBroker,
		SourceClientID: *sourceClientID,
		SourceTopic:    *sourceTopic,
		DestBroker:     *destBroker,
		DestClientID:   *destClientID,
		DestTopic:      *destTopic,
		Username:       *username,
		Password:       *password,
	}

	log.Printf("Starting Meshtastic MQTT Relay")
	log.Printf("Source: %s (topic: %s)", config.SourceBroker, config.SourceTopic)
	log.Printf("Destination: %s (topic: %s)", config.DestBroker, config.DestTopic)

	relay, err := relay.NewRelay(config)
	if err != nil {
		log.Fatalf("Failed to create relay: %v", err)
	}

	if err := relay.Start(); err != nil {
		log.Fatalf("Failed to start relay: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	relay.Stop()
}
