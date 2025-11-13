package relay_test

import (
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"math"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/relay"
)

const testTopic = "msh/ANZ/2/json/MediumFast/!44be043f"

// func TestConvertToJSON(t *testing.T) {
// 	relayClient := &relay.Relay{}

// 	// Test basic ServiceEnvelope with packet
// 	envelope := &meshtastic.ServiceEnvelope{
// 		Packet: &meshtastic.MeshPacket{
// 			From:     123456789,
// 			To:       987654321,
// 			Channel:  0,
// 			Id:       12345,
// 			RxTime:   1699435200,
// 			RxSnr:    8.5,
// 			RxRssi:   -85,
// 			HopLimit: 3,
// 			WantAck:  false,
// 			ViaMqtt:  false,
// 			HopStart: 3,
// 		},
// 		ChannelId: "LongFast",
// 		GatewayId: "!12345678",
// 	}

// 	jsonData, err := relayClient.ConvertToJSON(testTopic, envelope)
// 	if err != nil {
// 		t.Fatalf("Failed to convert to JSON: %v", err)
// 	}

// 	// Validate the JSON is valid
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(jsonData, &result); err != nil {
// 		t.Fatalf("Failed to unmarshal JSON: %v", err)
// 	}

// 	// Check that key fields are present
// 	if result["channelId"] != "LongFast" {
// 		t.Errorf("Expected channelId to be 'LongFast', got %v", result["channelId"])
// 	}

// 	if result["gatewayId"] != "!12345678" {
// 		t.Errorf("Expected gatewayId to be '!12345678', got %v", result["gatewayId"])
// 	}

// 	// Check packet data
// 	packet, ok := result["packet"].(map[string]interface{})
// 	if !ok {
// 		t.Fatal("Expected packet to be a map")
// 	}

// 	// Note: JSON numbers are float64
// 	if from := packet["from"].(float64); from != 123456789 {
// 		t.Errorf("Expected from to be 123456789, got %v", from)
// 	}

// 	if to := packet["to"].(float64); to != 987654321 {
// 		t.Errorf("Expected to to be 987654321, got %v", to)
// 	}
// }

// func TestConvertToJSONWithDecoded(t *testing.T) {
// 	relay := &relay.Relay{}

// 	// Test ServiceEnvelope with decoded payload
// 	envelope := &meshtastic.ServiceEnvelope{
// 		Packet: &meshtastic.MeshPacket{
// 			From:    123456789,
// 			To:      987654321,
// 			Channel: 0,
// 			PayloadVariant: &meshtastic.MeshPacket_Decoded{
// 				Decoded: &meshtastic.Data{
// 					Portnum:      meshtastic.PortNum_TEXT_MESSAGE_APP,
// 					Payload:      []byte("Hello World"),
// 					WantResponse: false,
// 					Dest:         987654321,
// 					Source:       123456789,
// 				},
// 			},
// 		},
// 	}

// 	jsonData, err := relay.ConvertToJSON(testTopic, envelope)
// 	if err != nil {
// 		t.Fatalf("Failed to convert to JSON: %v", err)
// 	}

// 	// Validate the JSON is valid
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(jsonData, &result); err != nil {
// 		t.Fatalf("Failed to unmarshal JSON: %v", err)
// 	}

// 	// Check packet data
// 	packet, ok := result["packet"].(map[string]interface{})
// 	if !ok {
// 		t.Fatal("Expected packet to be a map")
// 	}

// 	// Check decoded data
// 	decoded, ok := packet["decoded"].(map[string]interface{})
// 	if !ok {
// 		t.Fatal("Expected decoded to be a map")
// 	}

// 	if decoded["portnum"] != "TEXT_MESSAGE_APP" {
// 		t.Errorf("Expected portnum to be 'TEXT_MESSAGE_APP', got %v", decoded["portnum"])
// 	}

// 	// Payload is base64 encoded in JSON
// 	if decoded["payload"] == nil {
// 		t.Error("Expected payload to be present")
// 	}
// }

// func TestConvertToJSONWithEncrypted(t *testing.T) {
// 	relay := &relay.Relay{}

// 	// Test ServiceEnvelope with encrypted payload
// 	envelope := &meshtastic.ServiceEnvelope{
// 		Packet: &meshtastic.MeshPacket{
// 			From:    123456789,
// 			To:      987654321,
// 			Channel: 0,
// 			PayloadVariant: &meshtastic.MeshPacket_Encrypted{
// 				Encrypted: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
// 			},
// 		},
// 	}

// 	jsonData, err := relay.ConvertToJSON(testTopic, envelope)
// 	if err != nil {
// 		t.Fatalf("Failed to convert to JSON: %v", err)
// 	}

// 	// Validate the JSON is valid
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(jsonData, &result); err != nil {
// 		t.Fatalf("Failed to unmarshal JSON: %v", err)
// 	}

// 	// Check packet data
// 	packet, ok := result["packet"].(map[string]interface{})
// 	if !ok {
// 		t.Fatal("Expected packet to be a map")
// 	}

// 	// Check encrypted data is present
// 	if packet["encrypted"] == nil {
// 		t.Error("Expected encrypted payload to be present")
// 	}
// }

func TestNewRelay(t *testing.T) {
	config := relay.Config{
		Broker:   "tcp://localhost:1883",
		ClientID: "test-client",
		Topic:    "test/topic",
	}

	relayClient, err := relay.NewRelay(config, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("Failed to create relay: %v", err)
	}

	if relayClient == nil {
		t.Fatal("Expected relay to be non-nil")
	}

	if relayClient.Config.Broker != config.Broker {
		t.Errorf("Expected Broker to be %s, got %s", config.Broker, relayClient.Config.Broker)
	}
}

func TestConvertToJSONExamples(t *testing.T) {
	// const encodedMessage = `ClQNAqkdVhX/////IicIAxIhDQCAqO8VAIAzWxg7JSk1FGkoAljuAngCgAEAmAEFuAEQSAA1aUXeZT07NRRpRQAATMFIA2CM//////////8BeAWYAR4SCk1lZGl1bUZhc3QaCSE0NGJlMDQzZg==`
	// const encodedMessage = `Ck4NoBJToBX/////IiEIQxIdDbxCFGkSFghlFTeJhUAdTxtYQCVhxI08KLb50gE1VKDu4z22QRRpRQAA0EBIAmDC//////////8BeAeYATgSCk1lZGl1bUZhc3QaCSE0NGJlMDQzZg==`
	// const encodedMessage = `Co8BDVBWMEoV/////yJiCAQSXAoJITRhMzA1NjUwEhxUZXN0IFNlbWktUGVybWFuZW50IFJlcGVhdGVyGgNsbzAiBiTsSjBWUCgQQiCcQ4D2/JjAwtC31HoIzbngJRmLWcsdxcO043jvDPRdf0gASAA1243IIz2lQhRpRQAAMEFIB2DO//////////8BeAeYAVASCk1lZGl1bUZhc3QaCSE0NGJlMDQzZg==`
	const encodedMessage = `CooBDSDPWnwVSASN9yJdCAQSVAoJITdjNWFjZjIwEhVBbm5lcmxleSBKdW5jdGlvbiBXU0waBGNmMjAiBiRYfFrPICgsQiBy9dXwrv1cheaZadLd6mkQc8qaIOyVAbhtziuwmcQTfjVCWQJ6NYe3B1A9YEMUaUUAADRBSAFg0P//////////AXgFmAFQEgpNZWRpdW1GYXN0GgkhNDRiZTA0M2Y=`
	relayClient := &relay.Relay{
		Logger: slog.New(slog.DiscardHandler),
	}

	data, err := base64.StdEncoding.DecodeString(encodedMessage)
	if err != nil {
		t.Fatalf("Failed to decode base64 message: %v", err)
	}

	relayClient.HandleMessagePayload(data, "msh/ANZ/2/e/MediumFast/!44be043f")
	// t.Fail()
}

type convertJSONTestCase struct {
	name           string
	encodedMessage string
	expectedJSON   []byte
}

func loadTestCase(t *testing.T, filename string) convertJSONTestCase {
	t.Helper()
	testCase := convertJSONTestCase{
		name: filename,
	}

	{
		data, err := os.ReadFile(path.Join("..", "..", "testdata", "msgs", filename+".json"))
		if err != nil {
			t.Fatalf("Failed to read test case file %s: %v", filename, err)
		}
		testCase.expectedJSON = data
	}
	{
		data, err := os.ReadFile(path.Join("..", "..", "testdata", "msgs", filename+".enc"))
		if err != nil {
			t.Fatalf("Failed to read test case file %s: %v", filename, err)
		}
		testCase.encodedMessage = string(data)
	}

	return testCase
}

func TestConvertToJSONTable(t *testing.T) {
	tests := []convertJSONTestCase{
		loadTestCase(t, "message-01"),
		loadTestCase(t, "message-02"),
		loadTestCase(t, "message-03"),
		loadTestCase(t, "message-04"),
		loadTestCase(t, "message-05"),
		loadTestCase(t, "message-06"),
		loadTestCase(t, "message-07"),
		loadTestCase(t, "message-08"),
		loadTestCase(t, "message-09"),
	}

	transformJSON := []cmp.Option{
		cmp.Comparer(func(x, y float64) bool {
			if x == y {
				return true
			}
			delta := math.Abs(x - y)
			mean := math.Abs(x+y) / 2.0
			return delta/mean < 0.001
		}),
		cmp.FilterValues(func(x, y []byte) bool {
			return json.Valid(x) && json.Valid(y)
		}, cmp.Transformer("ParseJSON", func(in []byte) interface{} {
			var out any
			if err := json.Unmarshal(in, &out); err != nil {
				panic(err) // should never occur given previous filter to ensure valid JSON
			}
			return out
		})),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			relayClient := &relay.Relay{
				Logger: slog.New(slog.DiscardHandler),
			}

			data, err := base64.StdEncoding.DecodeString(tt.encodedMessage)
			if err != nil {
				t.Fatalf("Failed to decode base64 message: %v", err)
			}

			payload, _ := relayClient.HandleMessagePayload(data, testTopic)

			// log.Printf("Converted JSON: %s", payload)
			// log.Printf("Expected JSON: %s", tt.expectedJSON)

			if diff := cmp.Diff(payload, tt.expectedJSON, transformJSON...); diff != "" {
				t.Errorf("Converted JSON does not match expected (-got +want):\n%s", diff)
			}
		})
	}
}
