package relay

import (
	"encoding/json"
	"testing"

	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/pkg/meshtastic"
)

func TestConvertToJSON(t *testing.T) {
	relay := &Relay{}

	// Test basic ServiceEnvelope with packet
	envelope := &meshtastic.ServiceEnvelope{
		Packet: &meshtastic.MeshPacket{
			From:     123456789,
			To:       987654321,
			Channel:  0,
			Id:       12345,
			RxTime:   1699435200,
			RxSnr:    8.5,
			RxRssi:   -85,
			HopLimit: 3,
			WantAck:  false,
			ViaMqtt:  false,
			HopStart: 3,
		},
		ChannelId: "LongFast",
		GatewayId: "!12345678",
	}

	jsonData, err := relay.convertToJSON(envelope)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Validate the JSON is valid
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check that key fields are present
	if result["channelId"] != "LongFast" {
		t.Errorf("Expected channelId to be 'LongFast', got %v", result["channelId"])
	}

	if result["gatewayId"] != "!12345678" {
		t.Errorf("Expected gatewayId to be '!12345678', got %v", result["gatewayId"])
	}

	// Check packet data
	packet, ok := result["packet"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected packet to be a map")
	}

	// Note: JSON numbers are float64
	if from := packet["from"].(float64); from != 123456789 {
		t.Errorf("Expected from to be 123456789, got %v", from)
	}

	if to := packet["to"].(float64); to != 987654321 {
		t.Errorf("Expected to to be 987654321, got %v", to)
	}
}

func TestConvertToJSONWithDecoded(t *testing.T) {
	relay := &Relay{}

	// Test ServiceEnvelope with decoded payload
	envelope := &meshtastic.ServiceEnvelope{
		Packet: &meshtastic.MeshPacket{
			From:    123456789,
			To:      987654321,
			Channel: 0,
			PayloadVariant: &meshtastic.MeshPacket_Decoded{
				Decoded: &meshtastic.Data{
					Portnum:      meshtastic.PortNum_TEXT_MESSAGE_APP,
					Payload:      []byte("Hello World"),
					WantResponse: false,
					Dest:         987654321,
					Source:       123456789,
				},
			},
		},
	}

	jsonData, err := relay.convertToJSON(envelope)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Validate the JSON is valid
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check packet data
	packet, ok := result["packet"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected packet to be a map")
	}

	// Check decoded data
	decoded, ok := packet["decoded"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected decoded to be a map")
	}

	if decoded["portnum"] != "TEXT_MESSAGE_APP" {
		t.Errorf("Expected portnum to be 'TEXT_MESSAGE_APP', got %v", decoded["portnum"])
	}

	// Payload is base64 encoded in JSON
	if decoded["payload"] == nil {
		t.Error("Expected payload to be present")
	}
}

func TestConvertToJSONWithEncrypted(t *testing.T) {
	relay := &Relay{}

	// Test ServiceEnvelope with encrypted payload
	envelope := &meshtastic.ServiceEnvelope{
		Packet: &meshtastic.MeshPacket{
			From:    123456789,
			To:      987654321,
			Channel: 0,
			PayloadVariant: &meshtastic.MeshPacket_Encrypted{
				Encrypted: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			},
		},
	}

	jsonData, err := relay.convertToJSON(envelope)
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	// Validate the JSON is valid
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Check packet data
	packet, ok := result["packet"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected packet to be a map")
	}

	// Check encrypted data is present
	if packet["encrypted"] == nil {
		t.Error("Expected encrypted payload to be present")
	}
}

func TestNewRelay(t *testing.T) {
	config := Config{
		SourceBroker:   "tcp://localhost:1883",
		SourceClientID: "test-source",
		SourceTopic:    "test/topic",
		DestBroker:     "tcp://localhost:1884",
		DestClientID:   "test-dest",
		DestTopic:      "test/dest",
	}

	relay, err := NewRelay(config)
	if err != nil {
		t.Fatalf("Failed to create relay: %v", err)
	}

	if relay == nil {
		t.Fatal("Expected relay to be non-nil")
	}

	if relay.config.SourceBroker != config.SourceBroker {
		t.Errorf("Expected SourceBroker to be %s, got %s", config.SourceBroker, relay.config.SourceBroker)
	}
}
