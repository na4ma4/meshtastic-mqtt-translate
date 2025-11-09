package relay

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/pkg/meshtastic"
	"google.golang.org/protobuf/proto"
)

// Config holds the relay configuration
type Config struct {
	SourceBroker   string
	SourceClientID string
	SourceTopic    string
	DestBroker     string
	DestClientID   string
	DestTopic      string
	Username       string
	Password       string
}

// Relay handles the MQTT relay logic
type Relay struct {
	config       Config
	sourceClient mqtt.Client
	destClient   mqtt.Client
	wg           sync.WaitGroup
}

// NewRelay creates a new relay instance
func NewRelay(config Config) (*Relay, error) {
	return &Relay{
		config: config,
	}, nil
}

// Start begins the relay operation
func (r *Relay) Start() error {
	// Connect to destination broker first
	destOpts := mqtt.NewClientOptions().
		AddBroker(r.config.DestBroker).
		SetClientID(r.config.DestClientID)

	if r.config.Username != "" {
		destOpts.SetUsername(r.config.Username)
		destOpts.SetPassword(r.config.Password)
	}

	r.destClient = mqtt.NewClient(destOpts)
	if token := r.destClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to destination broker: %w", token.Error())
	}
	log.Println("Connected to destination MQTT broker")

	// Connect to source broker
	sourceOpts := mqtt.NewClientOptions().
		AddBroker(r.config.SourceBroker).
		SetClientID(r.config.SourceClientID)

	if r.config.Username != "" {
		sourceOpts.SetUsername(r.config.Username)
		sourceOpts.SetPassword(r.config.Password)
	}

	sourceOpts.SetDefaultPublishHandler(r.messageHandler)

	r.sourceClient = mqtt.NewClient(sourceOpts)
	if token := r.sourceClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to source broker: %w", token.Error())
	}
	log.Println("Connected to source MQTT broker")

	// Subscribe to source topic
	if token := r.sourceClient.Subscribe(r.config.SourceTopic, 0, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", token.Error())
	}
	log.Printf("Subscribed to topic: %s", r.config.SourceTopic)

	return nil
}

// Stop stops the relay
func (r *Relay) Stop() {
	if r.sourceClient != nil && r.sourceClient.IsConnected() {
		r.sourceClient.Disconnect(250)
		log.Println("Disconnected from source MQTT broker")
	}
	if r.destClient != nil && r.destClient.IsConnected() {
		r.destClient.Disconnect(250)
		log.Println("Disconnected from destination MQTT broker")
	}
	r.wg.Wait()
}

// messageHandler processes incoming MQTT messages
func (r *Relay) messageHandler(client mqtt.Client, msg mqtt.Message) {
	r.wg.Add(1)
	defer r.wg.Done()

	log.Printf("Received message on topic: %s", msg.Topic())

	// Attempt to decode as ServiceEnvelope (the standard Meshtastic MQTT format)
	var envelope meshtastic.ServiceEnvelope
	if err := proto.Unmarshal(msg.Payload(), &envelope); err != nil {
		log.Printf("Failed to unmarshal ServiceEnvelope: %v", err)
		return
	}

	// Convert to JSON
	jsonData, err := r.convertToJSON(&envelope)
	if err != nil {
		log.Printf("Failed to convert to JSON: %v", err)
		return
	}

	// Publish to destination
	topic := fmt.Sprintf("%s/%s", r.config.DestTopic, msg.Topic())
	if token := r.destClient.Publish(topic, 0, false, jsonData); token.Wait() && token.Error() != nil {
		log.Printf("Failed to publish to destination: %v", token.Error())
		return
	}

	log.Printf("Successfully relayed message to topic: %s", topic)
}

// convertToJSON converts a ServiceEnvelope to JSON
func (r *Relay) convertToJSON(envelope *meshtastic.ServiceEnvelope) ([]byte, error) {
	// Create a map representation for better JSON output
	data := make(map[string]interface{})

	if envelope.Packet != nil {
		data["packet"] = map[string]interface{}{
			"from":         envelope.Packet.From,
			"to":           envelope.Packet.To,
			"channel":      envelope.Packet.Channel,
			"id":           envelope.Packet.Id,
			"rxTime":       envelope.Packet.RxTime,
			"rxSnr":        envelope.Packet.RxSnr,
			"rxRssi":       envelope.Packet.RxRssi,
			"hopLimit":     envelope.Packet.HopLimit,
			"wantAck":      envelope.Packet.WantAck,
			"viaMqtt":      envelope.Packet.ViaMqtt,
			"hopStart":     envelope.Packet.HopStart,
			"publicKey":    envelope.Packet.PublicKey,
			"pkiEncrypted": envelope.Packet.PkiEncrypted,
		}

		// Decode the payload if present
		if decoded := envelope.Packet.GetDecoded(); decoded != nil {
			decodedData := map[string]interface{}{
				"portnum": decoded.Portnum.String(),
				"payload": decoded.Payload,
			}

			if decoded.WantResponse {
				decodedData["wantResponse"] = decoded.WantResponse
			}
			if decoded.Dest != 0 {
				decodedData["dest"] = decoded.Dest
			}
			if decoded.Source != 0 {
				decodedData["source"] = decoded.Source
			}
			if decoded.RequestId != 0 {
				decodedData["requestId"] = decoded.RequestId
			}
			if decoded.ReplyId != 0 {
				decodedData["replyId"] = decoded.ReplyId
			}
			if decoded.Emoji != 0 {
				decodedData["emoji"] = decoded.Emoji
			}

			data["packet"].(map[string]interface{})["decoded"] = decodedData
		}

		if encrypted := envelope.Packet.GetEncrypted(); len(encrypted) > 0 {
			data["packet"].(map[string]interface{})["encrypted"] = encrypted
		}
	}

	if envelope.ChannelId != "" {
		data["channelId"] = envelope.ChannelId
	}
	if envelope.GatewayId != "" {
		data["gatewayId"] = envelope.GatewayId
	}

	return json.Marshal(data)
}
