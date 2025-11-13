package relay

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/na4ma4/go-slogtool"
	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/store"
	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/translator"
	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/pkg/meshtastic"
	"google.golang.org/protobuf/proto"
)

// defaultQuiesceInMilliseconds - Default quiesce time for MQTT disconnects.
const defaultQuiesceInMilliseconds = 250

// Config holds the relay configuration.
type Config struct {
	Broker   string
	ClientID string
	Topic    string
	Username string
	Password string
	Store    store.Store
	DryRun   bool
}

// Relay handles the MQTT relay logic.
type Relay struct {
	Config       Config
	Logger       *slog.Logger
	sourceClient mqtt.Client
	destClient   mqtt.Client
	wg           sync.WaitGroup
}

// NewRelay creates a new relay instance.
func NewRelay(config Config, logger *slog.Logger) (*Relay, error) {
	return &Relay{
		Config: config,
		Logger: logger,
	}, nil
}

// Start begins the relay operation.
func (r *Relay) Start(ctx context.Context) error {
	// Connect to destination broker first
	destOpts := mqtt.NewClientOptions().
		AddBroker(r.Config.Broker).
		SetClientID(r.Config.ClientID + "-dest")

	if r.Config.Username != "" {
		destOpts.SetUsername(r.Config.Username)
		destOpts.SetPassword(r.Config.Password)
	}

	r.destClient = mqtt.NewClient(destOpts)
	if token := r.destClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to destination broker: %w", token.Error())
	}
	r.Logger.InfoContext(ctx, "Connected to destination MQTT broker")

	// Connect to source broker
	sourceOpts := mqtt.NewClientOptions().
		AddBroker(r.Config.Broker).
		SetClientID(r.Config.ClientID + "-source")

	if r.Config.Username != "" {
		sourceOpts.SetUsername(r.Config.Username)
		sourceOpts.SetPassword(r.Config.Password)
	}

	sourceOpts.SetDefaultPublishHandler(r.messageHandler)

	r.sourceClient = mqtt.NewClient(sourceOpts)
	if token := r.sourceClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to source broker: %w", token.Error())
	}
	r.Logger.InfoContext(ctx, "Connected to source MQTT broker")

	// Subscribe to source topic
	if token := r.sourceClient.Subscribe(r.Config.Topic, 0, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", token.Error())
	}
	r.Logger.InfoContext(ctx, "Subscribed to topic", "topic", r.Config.Topic)

	return nil
}

// Stop stops the relay.
func (r *Relay) Stop(ctx context.Context) {
	if r.sourceClient != nil && r.sourceClient.IsConnected() {
		r.sourceClient.Disconnect(defaultQuiesceInMilliseconds)
		r.Logger.InfoContext(ctx, "Disconnected from source MQTT broker")
	}
	if r.destClient != nil && r.destClient.IsConnected() {
		r.destClient.Disconnect(defaultQuiesceInMilliseconds)
		r.Logger.InfoContext(ctx, "Disconnected from destination MQTT broker")
	}
	r.wg.Wait()
}

// messageHandler processes incoming MQTT messages.
func (r *Relay) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	r.wg.Add(1)
	defer r.wg.Done()

	payload, topic := r.HandleMessagePayload(msg.Payload(), msg.Topic())
	if payload != nil && topic != "" {
		// Publish to destination
		if r.Config.DryRun {
			r.Logger.Debug("Dry run enabled, not publishing message", "topic", topic)
			return
		}
		if token := r.destClient.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			r.Logger.Error("Failed to publish to destination", slogtool.ErrorAttr(token.Error()))
		} else {
			r.Logger.Info(">", slog.String("topic", topic))
		}
	}
}

// HandleMessagePayload processes the message payload and returns JSON data and new topic.
func (r *Relay) HandleMessagePayload(payload []byte, topic string) ([]byte, string) {
	// Attempt to decode as ServiceEnvelope (the standard Meshtastic MQTT format)
	var envelope meshtastic.ServiceEnvelope
	if err := proto.Unmarshal(payload, &envelope); err != nil {
		r.Logger.Error("Failed to unmarshal ServiceEnvelope", slogtool.ErrorAttr(err))
		return nil, ""
	}

	r.Logger.Debug("Received message",
		slog.String("topic", topic),
		slog.String("payload", base64.StdEncoding.EncodeToString(payload)),
	)
	if envelope.GetPacket() == nil {
		r.Logger.Info("<", slog.String("topic", topic))
	} else {
		if dc := envelope.GetPacket().GetDecoded(); dc != nil {
			r.Logger.Info("<", slog.String("topic", topic), slog.String("portnum", dc.GetPortnum().String()))
		} else {
			r.Logger.Info("< [encrypted]", slog.String("topic", topic))
		}
	}

	// Convert to JSON
	jsonData, err := r.ConvertToJSON(topic, &envelope)
	if err != nil {
		r.Logger.Error("Failed to convert to JSON", slogtool.ErrorAttr(err))
		return nil, ""
	}

	r.conditionalStore(&envelope, payload, jsonData)

	newTopic := strings.Replace(topic, "/e/", "/json/", 1)

	// log.Printf(">: %s", newTopic)
	r.Logger.Debug("Relaying message",
		slog.String("topic", newTopic),
		slog.String("payload", string(jsonData)),
	)

	return jsonData, newTopic
}

func (r *Relay) conditionalStore(envelope *meshtastic.ServiceEnvelope, payload, jsonData []byte) {
	if r.Config.Store != nil {
		var portNum string
		messageID := strconv.FormatInt(int64(envelope.GetPacket().GetId()), 10)
		if envelope.GetPacket().GetDecoded() != nil {
			portNum = envelope.GetPacket().GetDecoded().GetPortnum().String()
		} else {
			portNum = "ENCRYPTED"
		}
		if saveErr := r.Config.Store.Save(messageID, portNum, payload, jsonData); saveErr != nil {
			r.Logger.Error("Failed to save JSON data", slogtool.ErrorAttr(saveErr))
		} else {
			r.Logger.Info("Saved JSON data",
				slog.String("messageID", messageID),
				slog.String("portNum", portNum),
			)
		}
	}
}

type Message struct {
	Bitfield  *uint32 `json:"bitfield,omitempty"`
	Channel   uint32  `json:"channel"`
	From      uint32  `json:"from"`
	HopStart  uint32  `json:"hop_start"`
	HopsAway  uint32  `json:"hops_away"`
	ID        uint32  `json:"id"`
	Payload   any     `json:"payload"`
	RSSI      int32   `json:"rssi"`
	Sender    string  `json:"sender"`
	SNR       float64 `json:"snr"`
	Timestamp uint32  `json:"timestamp"`
	To        uint32  `json:"to"`
	Type      string  `json:"type"`
}

// decodePayload decodes the payload based on the port number.
func (r *Relay) decodePayload(decoded *meshtastic.Data) (any, error) {
	switch decoded.GetPortnum() {
	case meshtastic.PortNum_TELEMETRY_APP:
		return r.decodeTelemetry(decoded.GetPayload())
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
	case meshtastic.PortNum_UNKNOWN_APP, //nolint:staticcheck // deprecated field
		meshtastic.PortNum_REMOTE_HARDWARE_APP,         // reserved for GPIO remote hardware
		meshtastic.PortNum_ROUTING_APP,                 // Protocol control packets for mesh protocol use.
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
func (r *Relay) decodeTelemetry(payload []byte) (any, error) {
	telemetry := &meshtastic.Telemetry{}
	if err := proto.Unmarshal(payload, telemetry); err != nil {
		return nil, err
	}

	switch variant := telemetry.GetVariant().(type) {
	case *meshtastic.Telemetry_DeviceMetrics:
		data, err := translator.New(translator.NewDeviceMetrics).Convert(variant.DeviceMetrics)
		if err == nil {
			data.Time = ptr(telemetry.GetTime())
		}
		return data, err
	case *meshtastic.Telemetry_LocalStats:
		return translator.New(translator.NewLocalStats).Convert(variant.LocalStats)
	case *meshtastic.Telemetry_PowerMetrics:
		return translator.New(translator.NewPowerMetrics).Convert(variant.PowerMetrics)
	case *meshtastic.Telemetry_HostMetrics:
		return translator.New(translator.NewHostMetrics).Convert(variant.HostMetrics)
	case *meshtastic.Telemetry_EnvironmentMetrics:
		return translator.New(translator.NewEnvironmentMetrics).Convert(variant.EnvironmentMetrics)
	case *meshtastic.Telemetry_AirQualityMetrics:
		return translator.New(translator.NewAirQualityMetrics).Convert(variant.AirQualityMetrics)
	default:
		return fmt.Sprintf("unknown telemetry variant: %T", variant), nil
	}
}

// ConvertToJSON converts a ServiceEnvelope to JSON.
func (r *Relay) ConvertToJSON(topic string, envelope *meshtastic.ServiceEnvelope) ([]byte, error) {
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
		SNR:       float64(envelope.GetPacket().GetRxSnr()),
		Timestamp: envelope.GetPacket().GetRxTime(),
		To:        envelope.GetPacket().GetTo(),
	}

	if decoded := envelope.GetPacket().GetDecoded(); decoded != nil {
		data.Bitfield = decoded.Bitfield
		data.Type = decoded.GetPortnum().String()

		payloadData, payloadErr := r.decodePayload(decoded)

		if payloadData != nil {
			if payloadErr == nil {
				data.Payload = payloadData
			} else {
				r.Logger.Error("Error converting payload", slogtool.ErrorAttr(payloadErr))
				data.Payload = decoded.GetPayload()
			}
		}
	}

	// if decoded := envelope.Packet.GetDecoded(); decoded != nil {
	// 	data.Payload = string(decoded.Payload)
	// 	data["type"] = decoded.Portnum.String()
	// }

	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
