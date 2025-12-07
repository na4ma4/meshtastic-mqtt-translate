package fanout

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/cmdconst"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/mtypes"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/parser"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/translator"
	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/na4ma4/go-contextual"
	"github.com/na4ma4/go-slogtool"
	"google.golang.org/protobuf/proto"
)

// Fanout handles the MQTT fanout logic.
type Fanout struct {
	Context      contextual.Context
	Config       Config
	Logger       *slog.Logger
	Parser       *parser.Parser
	sourceClient mqtt.Client
	destClient   mqtt.Client
	wg           sync.WaitGroup
	errChan      chan error
}

// NewFanout creates a new fanout instance.
func NewFanout(ctx contextual.Context, config Config, logger *slog.Logger) (*Fanout, error) {
	logger = logger.With(slog.String("type", "fanout"))
	f := &Fanout{
		Context: ctx,
		Config:  config,
		Logger:  logger,
		Parser:  parser.NewParser(logger),
		errChan: make(chan error, cmdconst.DefaultErrorChannelBufferSize),
	}
	return f, nil
}

func (f *Fanout) connectDest(ctx context.Context) {
	defer f.Logger.DebugContext(ctx, "connectDest(): finished")
	f.Logger.DebugContext(ctx, "connectDest(): starting")

	destOpts := mqtt.NewClientOptions().
		AddBroker(f.Config.Broker).
		SetClientID(f.Config.ClientID + "-dest").
		SetAutoReconnect(true).
		SetOnConnectHandler(f.destOnConnectHandler).
		SetOrderMatters(false).
		SetKeepAlive(f.Config.Keepalive)
		// .
		// SetWriteTimeout(defaultWriteTimeout)

	if f.Config.Username != "" {
		destOpts.SetUsername(f.Config.Username)
		destOpts.SetPassword(f.Config.Password)
		destOpts.SetAutoReconnect(true)
	}

	f.destClient = mqtt.NewClient(destOpts)
	token := f.destClient.Connect()

	select {
	case <-ctx.Done():
		f.Logger.InfoContext(ctx, "Context done before source broker connected")
		return
	case <-token.Done():
		// continue
	}

	if token.Error() != nil {
		f.errChan <- fmt.Errorf("failed to connect to destination broker: %w", token.Error())
		return
	}

	if f.Config.DryRun {
		f.Logger.InfoContext(ctx, "Dry run enabled, not publishing to destination broker")
		f.destClient.Disconnect(cmdconst.DefaultQuiesceInMilliseconds)
	}
}

func (f *Fanout) connectSrc(ctx context.Context) {
	defer f.Logger.DebugContext(ctx, "connectSrc(): finished")
	f.Logger.DebugContext(ctx, "connectSrc(): starting")

	sourceOpts := mqtt.NewClientOptions().
		AddBroker(f.Config.Broker).
		SetClientID(f.Config.ClientID + "-source").
		SetAutoReconnect(true).
		SetOnConnectHandler(f.srcOnConnectHandler).
		SetOrderMatters(false).
		SetKeepAlive(f.Config.Keepalive)
		// .
		// SetWriteTimeout(defaultWriteTimeout)

	if f.Config.Username != "" {
		sourceOpts.SetUsername(f.Config.Username)
		sourceOpts.SetPassword(f.Config.Password)
	}

	sourceOpts.SetDefaultPublishHandler(f.messageHandler)

	f.sourceClient = mqtt.NewClient(sourceOpts)
	token := f.sourceClient.Connect()

	select {
	case <-ctx.Done():
		f.Logger.InfoContext(ctx, "Context done before source broker connected")
		return
	case <-token.Done():
		// continue
	}

	if token.Error() != nil {
		f.errChan <- fmt.Errorf("failed to connect to source broker: %w", token.Error())
	}
}

// Start begins the relay operation.
func (f *Fanout) Start(ctx context.Context) {
	// Connect to destination broker first
	go f.connectDest(ctx)

	// Connect to source broker
	go f.connectSrc(ctx)
}

func (f *Fanout) destOnConnectHandler(client mqtt.Client) {
	f.Logger.Info("Connected to destination MQTT broker", slog.Bool("dest.connected", client.IsConnected()))
}

func (f *Fanout) srcOnConnectHandler(client mqtt.Client) {
	f.Logger.Info("Connected to source MQTT broker", slog.Bool("src.connected", client.IsConnected()))
	// Subscribe to source topic
	token := client.Subscribe(f.Config.SourceTopic, 0, nil)

	select {
	case <-f.Context.Done():
		f.Logger.InfoContext(f.Context, "Context done before subscription completed")
		return
	case <-token.Done():
		// continue
	}

	if token.Error() != nil {
		f.errChan <- fmt.Errorf("failed to subscribe to topic: %w", token.Error())
		return
	}
	f.Logger.Info("Subscribed to topic", slog.String("topic", f.Config.SourceTopic))
}

func (f *Fanout) Run(ctx context.Context) <-chan error {
	defer f.Logger.DebugContext(ctx, "Run(): finished")
	f.Logger.DebugContext(ctx, "Run(): starting")

	outErrChan := make(chan error, 1)

	go func() {
		defer f.Logger.DebugContext(ctx, "Run().go func(): finished")
		f.Logger.DebugContext(ctx, "Run().go func(): starting")

		defer close(outErrChan)
		defer f.Stop(ctx)
		f.Start(ctx)

		for {
			select {
			case <-ctx.Done():
				outErrChan <- ctx.Err()
				return
			case err := <-f.errChan:
				outErrChan <- err
				return
			}
		}
	}()

	return outErrChan
}

// Stop stops the relay.
func (f *Fanout) Stop(ctx context.Context) {
	if f.sourceClient != nil && f.sourceClient.IsConnected() {
		f.sourceClient.Disconnect(cmdconst.DefaultQuiesceInMilliseconds)
		f.Logger.InfoContext(ctx, "Disconnected from source MQTT broker")
	}
	if f.destClient != nil && f.destClient.IsConnected() {
		f.destClient.Disconnect(cmdconst.DefaultQuiesceInMilliseconds)
		f.Logger.InfoContext(ctx, "Disconnected from destination MQTT broker")
	}
	f.wg.Wait()
}

// messageHandler processes incoming MQTT messages.
func (f *Fanout) messageHandler(_ mqtt.Client, msg mqtt.Message) {
	f.wg.Add(1)
	defer f.wg.Done()
	defer msg.Ack()

	ctx, cancel := contextual.WithTimeout(f.Context, time.Minute)
	defer cancel()

	payload, topic := f.HandleMessagePayload(ctx, msg.Payload(), msg.Topic())
	if payload != nil && topic != "" {
		// Publish to destination
		if f.Config.DryRun {
			f.Logger.Debug("Dry run enabled, not publishing message", "topic", topic)
			return
		}
		if token := f.destClient.Publish(topic, 0, false, payload); token.Wait() && token.Error() != nil {
			f.Logger.Error("Failed to publish to destination", slogtool.ErrorAttr(token.Error()))
			f.errChan <- token.Error()
		} else {
			f.Logger.Info(">", slog.String("topic", topic))
		}
	}
}

func (f *Fanout) addCustomSuffixToTopic(topic string, msg *meshtastic.Data, mtMsg *mtypes.Message) string {
	switch msg.GetPortnum() { //nolint:exhaustive,gocritic // only needed for these specific so far.
	case meshtastic.PortNum_TELEMETRY_APP:
		switch mtMsg.Payload.(type) {
		case *translator.TelemetryEnvironmentMetrics:
			return path.Join(topic, "EnvironmentMetrics")
		case *translator.TelemetryDeviceMetrics:
			return path.Join(topic, "DeviceMetrics")
		case *translator.TelemetryAirQualityMetrics:
			return path.Join(topic, "AirQualityMetrics")
		case *translator.TelemetryHostMetrics:
			return path.Join(topic, "HostMetrics")
		case *meshtastic.Telemetry_LocalStats:
			return path.Join(topic, "LocalStats")
		case *meshtastic.Telemetry_PowerMetrics:
			return path.Join(topic, "PowerMetrics")
		}
	}

	return topic
}

// HandleMessagePayload processes the message payload and returns JSON data and new topic.
func (f *Fanout) HandleMessagePayload(ctx context.Context, payload []byte, topic string) ([]byte, string) {
	// Attempt to decode as ServiceEnvelope (the standard Meshtastic MQTT format)
	var envelope meshtastic.ServiceEnvelope
	if err := proto.Unmarshal(payload, &envelope); err != nil {
		f.Logger.ErrorContext(ctx, "Failed to unmarshal ServiceEnvelope", slogtool.ErrorAttr(err))
		return nil, ""
	}

	if f.Logger.Enabled(ctx, slog.LevelDebug) {
		f.Logger.DebugContext(ctx, "Received message",
			slog.String("topic", topic),
			slog.String("payload", base64.StdEncoding.EncodeToString(payload)),
		)
	}

	if envelope.GetPacket() == nil {
		f.Logger.InfoContext(ctx, "< [no packet]", slog.String("topic", topic))
		return nil, ""
	}

	dc := envelope.GetPacket().GetDecoded()
	if dc == nil {
		f.Logger.InfoContext(ctx, "< [encrypted]", slog.String("topic", topic))
		return nil, ""
	}

	// Convert to JSON
	var message *mtypes.Message
	{
		var err error
		message, err = f.Parser.ConvertToMessage(ctx, topic, payload, &envelope)
		if err != nil {
			f.Logger.ErrorContext(ctx, "Failed to convert to Message", slogtool.ErrorAttr(err))
			return nil, ""
		}
	}

	var jsonData []byte
	{
		var err error
		jsonData, err = message.ToJSON()
		if err != nil {
			f.Logger.ErrorContext(ctx, "Failed to convert to JSON", slogtool.ErrorAttr(err))
			return nil, ""
		}
	}

	newTopic := path.Join(
		f.Config.TargetBaseTopic,
		strconv.FormatUint(uint64(envelope.GetPacket().GetFrom()), 10),
		dc.GetPortnum().String(),
	)

	newTopic = f.addCustomSuffixToTopic(newTopic, dc, message)

	// log.Printf(">: %s", newTopic)
	f.Logger.DebugContext(ctx, "Relaying message",
		slog.String("topic", newTopic),
		slog.String("payload", string(jsonData)),
	)

	return jsonData, newTopic
}
