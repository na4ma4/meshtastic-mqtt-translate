package fanout_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"math"
	"os"
	"path"
	"testing"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/fanout"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/parser"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/relay"

	"github.com/google/go-cmp/cmp"
	"github.com/na4ma4/go-contextual"
)

const testTopic = "msh/ANZ/2/json/MediumFast/!44be043f"

func TestNewRelay(t *testing.T) {
	config := fanout.Config{
		Broker:      "tcp://localhost:1883",
		ClientID:    "test-client",
		SourceTopic: "test/topic",
	}

	fanoutClient, err := fanout.NewFanout(contextual.New(context.TODO()), config, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("Failed to create relay: %v", err)
	}

	if fanoutClient == nil {
		t.Fatal("Expected relay to be non-nil")
	}

	if fanoutClient.Config.Broker != config.Broker {
		t.Errorf("Expected Broker to be %s, got %s", config.Broker, fanoutClient.Config.Broker)
	}
}

func TestConvertToJSONExamples(t *testing.T) {
	// const encodedMessage = `ClQNAqkdVhX/////IicIAxIhDQCAqO8VAIAzWxg7JSk1FGkoAljuAngCgAEAmAEFuAEQSAA1aUXeZT07NRRpRQAATMFIA2CM//////////8BeAWYAR4SCk1lZGl1bUZhc3QaCSE0NGJlMDQzZg==`
	// const encodedMessage = `Ck4NoBJToBX/////IiEIQxIdDbxCFGkSFghlFTeJhUAdTxtYQCVhxI08KLb50gE1VKDu4z22QRRpRQAA0EBIAmDC//////////8BeAeYATgSCk1lZGl1bUZhc3QaCSE0NGJlMDQzZg==`
	// const encodedMessage = `Co8BDVBWMEoV/////yJiCAQSXAoJITRhMzA1NjUwEhxUZXN0IFNlbWktUGVybWFuZW50IFJlcGVhdGVyGgNsbzAiBiTsSjBWUCgQQiCcQ4D2/JjAwtC31HoIzbngJRmLWcsdxcO043jvDPRdf0gASAA1243IIz2lQhRpRQAAMEFIB2DO//////////8BeAeYAVASCk1lZGl1bUZhc3QaCSE0NGJlMDQzZg==`
	const encodedMessage = `CooBDSDPWnwVSASN9yJdCAQSVAoJITdjNWFjZjIwEhVBbm5lcmxleSBKdW5jdGlvbiBXU0waBGNmMjAiBiRYfFrPICgsQiBy9dXwrv1cheaZadLd6mkQc8qaIOyVAbhtziuwmcQTfjVCWQJ6NYe3B1A9YEMUaUUAADRBSAFg0P//////////AXgFmAFQEgpNZWRpdW1GYXN0GgkhNDRiZTA0M2Y=`
	fanoutClient := &fanout.Fanout{
		Logger: slog.New(slog.DiscardHandler),
		Parser: parser.NewParser(slog.New(slog.DiscardHandler)),
	}

	data, err := base64.StdEncoding.DecodeString(encodedMessage)
	if err != nil {
		t.Fatalf("Failed to decode base64 message: %v", err)
	}

	fanoutClient.HandleMessagePayload(t.Context(), data, "msh/ANZ/2/e/MediumFast/!44be043f")
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
		loadTestCase(t, "message-10"),
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
			logger := slog.New(slog.NewTextHandler(t.Output(), &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			fanoutClient := &relay.Relay{
				// Logger: slog.New(slog.DiscardHandler),
				Logger: logger,
				Parser: parser.NewParser(logger),
			}

			data, err := base64.StdEncoding.DecodeString(tt.encodedMessage)
			if err != nil {
				t.Fatalf("Failed to decode base64 message: %v", err)
			}

			payload, _ := fanoutClient.HandleMessagePayload(t.Context(), data, testTopic)

			// log.Printf("Converted JSON: %s", payload)
			// log.Printf("Expected JSON: %s", tt.expectedJSON)

			if diff := cmp.Diff(payload, tt.expectedJSON, transformJSON...); diff != "" {
				t.Errorf("Converted JSON does not match expected (-got +want):\n%s", diff)
			}
		})
	}
}
