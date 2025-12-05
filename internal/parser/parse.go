package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/na4ma4/go-slogtool"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/translator"
	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"
	"google.golang.org/protobuf/proto"
)

type OnParseHandler func(ctx context.Context, envelope *meshtastic.ServiceEnvelope, payload []byte, data *Message) error

type Config struct {
	OnParseHandler OnParseHandler
}

type Parser struct {
	Config *Config
	Logger *slog.Logger
}

func NewParser(logger *slog.Logger, opts ...OptionFunc) *Parser {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}
	return &Parser{
		Config: config,
		Logger: logger,
	}
}

// ConvertToJSON converts a ServiceEnvelope to JSON.
func (p *Parser) ConvertToJSON(
	ctx context.Context,
	topic string,
	payload []byte,
	envelope *meshtastic.ServiceEnvelope,
) ([]byte, error) {
	// Create a map representation for better JSON output
	// data := make(map[string]interface{})

	// return protojson.Marshal(envelope)

	var sender string
	{
		sp := strings.Split(topic, "/")
		if len(sp) > 0 {
			sender = sp[len(sp)-1]
		}
	}

	data := &Message{
		Channel:   envelope.GetPacket().GetChannel(),
		From:      envelope.GetPacket().GetFrom(),
		HopStart:  envelope.GetPacket().GetHopStart(),
		HopsAway:  envelope.GetPacket().GetHopStart() - envelope.GetPacket().GetHopLimit(),
		ID:        envelope.GetPacket().GetId(),
		RSSI:      envelope.GetPacket().GetRxRssi(),
		Sender:    sender,
		SNR:       translator.SpecialFloat64(envelope.GetPacket().GetRxSnr()),
		Timestamp: envelope.GetPacket().GetRxTime(),
		To:        envelope.GetPacket().GetTo(),
	}

	if decoded := envelope.GetPacket().GetDecoded(); decoded != nil {
		data.Bitfield = decoded.Bitfield
		data.Type = decoded.GetPortnum().String()

		payloadData, payloadErr := p.decodePayload(ctx, decoded)

		if payloadData != nil {
			if payloadErr == nil {
				data.Payload = payloadData
			} else {
				p.Logger.ErrorContext(ctx, "Error converting payload", slogtool.ErrorAttr(payloadErr))
				data.Payload = decoded.GetPayload()
			}
		}
	}

	// if decoded := envelope.Packet.GetDecoded(); decoded != nil {
	// 	data.Payload = string(decoded.Payload)
	// 	data["type"] = decoded.Portnum.String()
	// }

	// p.conditionalStore(ctx, envelope, payload, data)
	if p.Config.OnParseHandler != nil {
		if err := p.Config.OnParseHandler(ctx, envelope, payload, data); err != nil {
			return nil, fmt.Errorf("OnParseHandler error: %w", err)
		}
	}

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	// log.Printf("Message: %+v", data)
	// enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return nil, fmt.Errorf("failed to encode JSON: %w", err)
	}

	return buf.Bytes(), nil
}

// decodePayload decodes the payload based on the port number.
func (p *Parser) decodePayload(ctx context.Context, decoded *meshtastic.Data) (any, error) {
	switch decoded.GetPortnum() {
	case meshtastic.PortNum_TELEMETRY_APP:
		return p.decodeTelemetry(ctx, decoded.GetPayload())
	case meshtastic.PortNum_NODEINFO_APP:
		return translator.New(translator.NewUser).Decode(decoded.GetPayload())
	case meshtastic.PortNum_POSITION_APP:
		return translator.New(translator.NewPositionApp).Decode(decoded.GetPayload())
	case meshtastic.PortNum_TEXT_MESSAGE_APP:
		return string(decoded.GetPayload()), nil
	case meshtastic.PortNum_STORE_FORWARD_APP: // STORE_FORWARD_APP (Work in Progress)
		return translator.New(translator.NewStoreForwardApp).Decode(decoded.GetPayload())
	case meshtastic.PortNum_TRACEROUTE_APP: // TODO: Provides a traceroute functionality to show the route a packet towards
		return translator.New(translator.NewTracerouteApp).Decode(decoded.GetPayload())
	case meshtastic.PortNum_ROUTING_APP: // Protocol control packets for mesh protocol use.
		return translator.New(translator.NewRoutingApp).Decode(decoded.GetPayload())
	case meshtastic.PortNum_UNKNOWN_APP, //nolint:staticcheck // deprecated field
		meshtastic.PortNum_REMOTE_HARDWARE_APP,         // reserved for GPIO remote hardware
		meshtastic.PortNum_ADMIN_APP,                   // Admin control packets.
		meshtastic.PortNum_TEXT_MESSAGE_COMPRESSED_APP, // Compressed TEXT_MESSAGE payloads. (handled in firmware)
		meshtastic.PortNum_WAYPOINT_APP,                // TODO: Waypoint payloads.
		meshtastic.PortNum_AUDIO_APP,                   // Audio payloads (2.4GHz only).
		meshtastic.PortNum_DETECTION_SENSOR_APP,        // TODO: Detection sensor payloads.
		meshtastic.PortNum_ALERT_APP,                   // TODO: Same as Text Message but used for critical alerts.
		meshtastic.PortNum_KEY_VERIFICATION_APP,        // TODO: Module/port for handling key verification requests.
		meshtastic.PortNum_REPLY_APP,                   // TODO: Provides a 'ping' service that replies to any packet it receives.
		meshtastic.PortNum_IP_TUNNEL_APP,               // TODO: Used for the python IP tunnel feature
		meshtastic.PortNum_PAXCOUNTER_APP,              // TODO: Paxcounter lib included in the firmware
		meshtastic.PortNum_SERIAL_APP,                  // TODO: Provides a hardware serial interface to send and receive from the Meshtastic network.
		meshtastic.PortNum_RANGE_TEST_APP,              // Optional port for messages for the range test module.
		meshtastic.PortNum_ZPS_APP,                     // TODO: Experimental tools for estimating node position without a GPS
		meshtastic.PortNum_SIMULATOR_APP,               // Used to let multiple instances of Linux native applications communicate
		meshtastic.PortNum_NEIGHBORINFO_APP,            // TODO: Aggregates edge info for the network by sending out a list of each node's neighbors
		meshtastic.PortNum_ATAK_PLUGIN,                 // TODO: ATAK Plugin
		meshtastic.PortNum_MAP_REPORT_APP,              // TODO: Provides unencrypted information about a node for consumption by a map via MQTT
		meshtastic.PortNum_POWERSTRESS_APP,             // PowerStress based monitoring support (for automated power consumption testing)
		meshtastic.PortNum_RETICULUM_TUNNEL_APP,        // TODO: Reticulum Network Stack Tunnel App
		meshtastic.PortNum_CAYENNE_APP,                 // App for transporting Cayenne Low Power Payload, popular for LoRaWAN sensor nodes.
		meshtastic.PortNum_PRIVATE_APP,                 // Private applications should use portnums >= 256.
		meshtastic.PortNum_ATAK_FORWARDER,              // ATAK Forwarder Module https://github.com/paulmandal/atak-forwarder
		meshtastic.PortNum_MAX:                         // Currently we limit port nums to no higher than this value
		fallthrough
	default:
		return decoded.GetPayload(), nil
	}
}

// decodeTelemetry decodes telemetry payloads.
func (p *Parser) decodeTelemetry(ctx context.Context, payload []byte) (any, error) {
	telemetry := &meshtastic.Telemetry{}
	if err := proto.Unmarshal(payload, telemetry); err != nil {
		return nil, err
	}

	switch variant := telemetry.GetVariant().(type) {
	case *meshtastic.Telemetry_DeviceMetrics:
		p.Logger.DebugContext(ctx, "Decoding DeviceMetrics telemetry")
		data, err := translator.New(translator.NewDeviceMetrics).Convert(variant.DeviceMetrics)
		if err == nil {
			data.Time = ptr(telemetry.GetTime())
		}
		return data, err
	case *meshtastic.Telemetry_LocalStats:
		p.Logger.DebugContext(ctx, "Decoding LocalStats telemetry")
		return translator.New(translator.NewLocalStats).Convert(variant.LocalStats)
	case *meshtastic.Telemetry_PowerMetrics:
		p.Logger.DebugContext(ctx, "Decoding PowerMetrics telemetry")
		return translator.New(translator.NewPowerMetrics).Convert(variant.PowerMetrics)
	case *meshtastic.Telemetry_HostMetrics:
		p.Logger.DebugContext(ctx, "Decoding HostMetrics telemetry")
		return translator.New(translator.NewHostMetrics).Convert(variant.HostMetrics)
	case *meshtastic.Telemetry_EnvironmentMetrics:
		p.Logger.DebugContext(ctx, "Decoding EnvironmentMetrics telemetry")
		return translator.New(translator.NewEnvironmentMetrics).Convert(variant.EnvironmentMetrics)
	case *meshtastic.Telemetry_AirQualityMetrics:
		p.Logger.DebugContext(ctx, "Decoding AirQualityMetrics telemetry")
		return translator.New(translator.NewAirQualityMetrics).Convert(variant.AirQualityMetrics)
	default:
		return fmt.Sprintf("unknown telemetry variant: %T", variant), nil
	}
}
