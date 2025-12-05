package relay

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/parser"
	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/na4ma4/go-contextual"
	"github.com/na4ma4/go-slogtool"
	"google.golang.org/protobuf/proto"
)

const (
	// defaultQuiesceInMilliseconds - Default quiesce time for MQTT disconnects.
	defaultQuiesceInMilliseconds = 250

	// defaultErrorChannelBufferSize - Default buffer size for error channel.
	defaultErrorChannelBufferSize = 10
)

// Relay handles the MQTT relay logic.
type Relay struct {
	Context      contextual.Context
	Config       Config
	Logger       *slog.Logger
	Parser       *parser.Parser
	sourceClient mqtt.Client
	destClient   mqtt.Client
	wg           sync.WaitGroup
	errChan      chan error
}

// NewRelay creates a new relay instance.
func NewRelay(ctx contextual.Context, config Config, logger *slog.Logger) (*Relay, error) {
	logger = logger.With(slog.String("type", "relay"))
	r := &Relay{
		Context: ctx,
		Config:  config,
		Logger:  logger,
		errChan: make(chan error, defaultErrorChannelBufferSize),
	}
	r.Parser = parser.NewParser(logger, parser.WithOnParseHandler(r.conditionalStore))
	return r, nil
}

func (r *Relay) connectDest(ctx context.Context) {
	defer r.Logger.DebugContext(ctx, "connectDest(): finished")
	r.Logger.DebugContext(ctx, "connectDest(): starting")

	destOpts := mqtt.NewClientOptions().
		AddBroker(r.Config.Broker).
		SetClientID(r.Config.ClientID + "-dest").
		SetAutoReconnect(true).
		SetOnConnectHandler(r.destOnConnectHandler).
		SetOrderMatters(false).
		SetKeepAlive(r.Config.Keepalive)
		// .
		// SetWriteTimeout(defaultWriteTimeout)

	if r.Config.Username != "" {
		destOpts.SetUsername(r.Config.Username)
		destOpts.SetPassword(r.Config.Password)
		destOpts.SetAutoReconnect(true)
	}

	r.destClient = mqtt.NewClient(destOpts)
	token := r.destClient.Connect()

	select {
	case <-ctx.Done():
		r.Logger.InfoContext(ctx, "Context done before source broker connected")
		return
	case <-token.Done():
		// continue
	}

	if token.Error() != nil {
		r.errChan <- fmt.Errorf("failed to connect to destination broker: %w", token.Error())
		return
	}

	if r.Config.DryRun {
		r.Logger.InfoContext(ctx, "Dry run enabled, not publishing to destination broker")
		r.destClient.Disconnect(defaultQuiesceInMilliseconds)
	}
}

func (r *Relay) connectSrc(ctx context.Context) {
	defer r.Logger.DebugContext(ctx, "connectSrc(): finished")
	r.Logger.DebugContext(ctx, "connectSrc(): starting")

	sourceOpts := mqtt.NewClientOptions().
		AddBroker(r.Config.Broker).
		SetClientID(r.Config.ClientID + "-source").
		SetAutoReconnect(true).
		SetOnConnectHandler(r.srcOnConnectHandler).
		SetOrderMatters(false).
		SetKeepAlive(r.Config.Keepalive)
		// .
		// SetWriteTimeout(defaultWriteTimeout)

	if r.Config.Username != "" {
		sourceOpts.SetUsername(r.Config.Username)
		sourceOpts.SetPassword(r.Config.Password)
	}

	sourceOpts.SetDefaultPublishHandler(r.messageHandler)

	r.sourceClient = mqtt.NewClient(sourceOpts)
	token := r.sourceClient.Connect()

	select {
	case <-ctx.Done():
		r.Logger.InfoContext(ctx, "Context done before source broker connected")
		return
	case <-token.Done():
		// continue
	}

	if token.Error() != nil {
		r.errChan <- fmt.Errorf("failed to connect to source broker: %w", token.Error())
	}
}

// Start begins the relay operation.
func (r *Relay) Start(ctx context.Context) {
	// Connect to destination broker first
	go r.connectDest(ctx)

	// Connect to source broker
	go r.connectSrc(ctx)
}

func (r *Relay) destOnConnectHandler(client mqtt.Client) {
	r.Logger.Info("Connected to destination MQTT broker", slog.Bool("dest.connected", client.IsConnected()))
}

func (r *Relay) srcOnConnectHandler(client mqtt.Client) {
	r.Logger.Info("Connected to source MQTT broker", slog.Bool("src.connected", client.IsConnected()))
	// Subscribe to source topic
	token := client.Subscribe(r.Config.Topic, 0, nil)

	select {
	case <-r.Context.Done():
		r.Logger.InfoContext(r.Context, "Context done before subscription completed")
		return
	case <-token.Done():
		// continue
	}

	if token.Error() != nil {
		r.errChan <- fmt.Errorf("failed to subscribe to topic: %w", token.Error())
		return
	}
	r.Logger.Info("Subscribed to topic", slog.String("topic", r.Config.Topic))
}

func (r *Relay) Run(ctx context.Context) <-chan error {
	defer r.Logger.DebugContext(ctx, "Run(): finished")
	r.Logger.DebugContext(ctx, "Run(): starting")

	outErrChan := make(chan error, 1)

	go func() {
		defer r.Logger.DebugContext(ctx, "Run().go func(): finished")
		r.Logger.DebugContext(ctx, "Run().go func(): starting")

		defer close(outErrChan)
		defer r.Stop(ctx)
		r.Start(ctx)

		for {
			select {
			case <-ctx.Done():
				outErrChan <- ctx.Err()
				return
			case err := <-r.errChan:
				outErrChan <- err
				return
			}
		}
	}()

	return outErrChan
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
	defer msg.Ack()

	ctx, cancel := contextual.WithTimeout(r.Context, time.Minute)
	defer cancel()

	payload, topic := r.HandleMessagePayload(ctx, msg.Payload(), msg.Topic())
	if payload != nil && topic != "" {
		// Publish to destination
		if r.Config.DryRun {
			r.Logger.Debug("Dry run enabled, not publishing message", "topic", topic)
			return
		}
		if token := r.destClient.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			r.Logger.Error("Failed to publish to destination", slogtool.ErrorAttr(token.Error()))
			r.errChan <- token.Error()
		} else {
			r.Logger.Info(">", slog.String("topic", topic))
		}
	}
}

// HandleMessagePayload processes the message payload and returns JSON data and new topic.
func (r *Relay) HandleMessagePayload(ctx context.Context, payload []byte, topic string) ([]byte, string) {
	// Attempt to decode as ServiceEnvelope (the standard Meshtastic MQTT format)
	var envelope meshtastic.ServiceEnvelope
	if err := proto.Unmarshal(payload, &envelope); err != nil {
		r.Logger.ErrorContext(ctx, "Failed to unmarshal ServiceEnvelope", slogtool.ErrorAttr(err))
		return nil, ""
	}

	if r.Logger.Enabled(ctx, slog.LevelDebug) {
		r.Logger.DebugContext(ctx, "Received message",
			slog.String("topic", topic),
			slog.String("payload", base64.StdEncoding.EncodeToString(payload)),
		)
	}
	if envelope.GetPacket() == nil {
		r.Logger.InfoContext(ctx, "<", slog.String("topic", topic))
	} else {
		if dc := envelope.GetPacket().GetDecoded(); dc != nil {
			r.Logger.InfoContext(ctx, "<", slog.String("topic", topic), slog.String("portnum", dc.GetPortnum().String()))
		} else {
			r.Logger.InfoContext(ctx, "< [encrypted]", slog.String("topic", topic))
		}
	}

	// Convert to JSON
	jsonData, err := r.Parser.ConvertToJSON(ctx, topic, payload, &envelope)
	if err != nil {
		r.Logger.ErrorContext(ctx, "Failed to convert to JSON", slogtool.ErrorAttr(err))
		return nil, ""
	}

	newTopic := strings.Replace(topic, "/e/", "/json/", 1)

	// log.Printf(">: %s", newTopic)
	r.Logger.DebugContext(ctx, "Relaying message",
		slog.String("topic", newTopic),
		slog.String("payload", string(jsonData)),
	)

	return jsonData, newTopic
}

func (r *Relay) conditionalStore(
	ctx context.Context,
	envelope *meshtastic.ServiceEnvelope,
	payload []byte,
	jsonData *parser.Message,
) error {
	if r.Config.Store != nil {
		var portNum string
		messageID := strconv.FormatInt(int64(envelope.GetPacket().GetId()), 10)
		if envelope.GetPacket().GetDecoded() != nil {
			portNum = envelope.GetPacket().GetDecoded().GetPortnum().String()
		} else {
			portNum = "ENCRYPTED"
		}
		if saveErr := r.Config.Store.Save(ctx, messageID, portNum, payload, jsonData); saveErr != nil {
			r.Logger.ErrorContext(ctx, "Failed to save JSON data", slogtool.ErrorAttr(saveErr))
		} else {
			r.Logger.DebugContext(ctx, "Saved JSON data",
				slog.String("messageID", messageID),
				slog.String("portNum", portNum),
			)
		}
	}

	return nil
}
