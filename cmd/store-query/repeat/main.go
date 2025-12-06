package repeat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/dosquad/go-cliversion"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/na4ma4/go-contextual"
	"github.com/na4ma4/go-slogtool"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/cmdconst"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/mainconfig"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CmdRepeat represents the base command when called without any subcommands.
var CmdRepeat = &cobra.Command{
	Use:          "repeat <message-ID>",
	Short:        "Repeat message from Datastore to MQTT Broker",
	Long:         `Repeat message from Datastore to MQTT Broker.`,
	RunE:         repeatCmd,
	Args:         cobra.MinimumNArgs(1),
	Version:      cliversion.Get().VersionString(),
	SilenceUsage: true,
}

func init() {
	cobra.OnInitialize(mainconfig.ConfigInit)

	CmdRepeat.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", CmdRepeat.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")

	CmdRepeat.PersistentFlags().StringP("broker", "b", "tcp://localhost:1883", "Source MQTT broker URL")
	_ = viper.BindPFlag("broker.address", CmdRepeat.PersistentFlags().Lookup("broker"))
	_ = viper.BindEnv("broker.address", "MQTT_BROKER")

	CmdRepeat.PersistentFlags().StringP("clientid", "c", "meshtastic-mqtt-relay", "MQTT client ID")
	_ = viper.BindPFlag("broker.clientid", CmdRepeat.PersistentFlags().Lookup("clientid"))
	_ = viper.BindEnv("broker.clientid", "MQTT_CLIENTID")

	CmdRepeat.PersistentFlags().StringP("username", "u", "", "MQTT username (optional)")
	_ = viper.BindPFlag("broker.username", CmdRepeat.PersistentFlags().Lookup("username"))
	_ = viper.BindEnv("broker.username", "MQTT_USERNAME")

	CmdRepeat.PersistentFlags().StringP("password", "p", "", "MQTT password (optional)")
	_ = viper.BindPFlag("broker.password", CmdRepeat.PersistentFlags().Lookup("password"))
	_ = viper.BindEnv("broker.password", "MQTT_PASSWORD")

	CmdRepeat.PersistentFlags().StringP("topic", "t", "msh/ANZ/2/e/#", "MQTT topic to subscribe to")
	_ = viper.BindPFlag("broker.topic", CmdRepeat.PersistentFlags().Lookup("topic"))
	_ = viper.BindEnv("broker.topic", "MQTT_TOPIC")

	CmdRepeat.PersistentFlags().BoolP("dry-run", "n", false, "Dry run mode (optional)")
	_ = viper.BindPFlag("dry-run", CmdRepeat.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindEnv("dry-run", "MQTT_DRY_RUN")

	CmdRepeat.PersistentFlags().StringP("dsn", "o", "", "Datastore DSN")
	_ = viper.BindPFlag("store.dsn", CmdRepeat.PersistentFlags().Lookup("dsn"))
	_ = viper.BindEnv("store.dsn", "STORE_DSN")
}

func repeatCmd(_ *cobra.Command, args []string) error {
	ctx := contextual.NewCancellable(context.Background())
	defer ctx.Cancel()

	storeCfg := store.Config{
		SlowThreshold: viper.GetDuration("store.slow-threshold"),
		LogLevel:      slog.LevelInfo,
	}
	if viper.GetBool("debug") {
		storeCfg.LogLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: storeCfg.LogLevel}))
	storeCfg.Logger = logger
	logger.Debug("Debug logging enabled") // will only show if debug is enabled

	logger.InfoContext(ctx, "Starting Datastore Query Tool")

	var st store.Store
	{
		var err error
		if st, err = getStore(viper.GetString("store.dsn"), storeCfg); err != nil && !errors.Is(err, ErrEmptyDSN) {
			logger.ErrorContext(ctx, "Failed to create store", slogtool.ErrorAttr(err))
			return fmt.Errorf("%w%w", cmdconst.ErrNoUsage, err)
		} else if errors.Is(err, ErrEmptyDSN) {
			logger.DebugContext(ctx, "No Store DSN set, exiting")
			return err
		} else if st == nil {
			logger.ErrorContext(ctx, "Store is nil")
			return fmt.Errorf("%wstore is nil", cmdconst.ErrNoUsage)
		}
	}

	msg, err := st.GetPayload(ctx, args[0])
	if err != nil {
		logger.ErrorContext(
			ctx, "Failed to get message",
			slogtool.ErrorAttr(err), slog.String("message_id", args[0]),
		)
		return fmt.Errorf("%w%w", cmdconst.ErrNoUsage, err)
	}
	if msg == nil {
		logger.InfoContext(ctx, "Message not found", slog.String("message_id", args[0]))
		return nil
	}

	destClient, err := connectDest(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to connect to destination broker: %w", err)
	}
	defer func() {
		if destClient != nil && destClient.IsConnected() {
			destClient.Disconnect(cmdconst.DefaultQuiesceInMilliseconds)
			logger.InfoContext(ctx, "Disconnected from destination MQTT broker")
		}
	}()

	token := destClient.Publish(
		viper.GetString("broker.topic"),
		0, false, msg,
	)

	if token.Wait() && token.Error() != nil {
		logger.ErrorContext(ctx, "Failed to publish to destination", slogtool.ErrorAttr(token.Error()))
		return fmt.Errorf("failed to publish to destination: %w", token.Error())
	}

	logger.InfoContext(ctx, "Published message to destination broker",
		slog.String("topic", viper.GetString("broker.topic")),
		slog.String("message_id", args[0]),
	)

	return nil
}

var ErrEmptyDSN = errors.New("empty DSN")

func getStore(dsn string, cfg store.Config) (store.Store, error) {
	if dsn == "" {
		return nil, ErrEmptyDSN
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	return store.NewDetectStore(u, cfg)
}

func connectDest(
	ctx context.Context,
	logger *slog.Logger,
) (mqtt.Client, error) {
	defer logger.DebugContext(ctx, "connectDest(): finished")
	logger.DebugContext(ctx, "connectDest(): starting")

	destOpts := mqtt.NewClientOptions().
		AddBroker(viper.GetString("broker.address")).
		SetClientID(viper.GetString("broker.clientid") + "-dest").
		SetAutoReconnect(true).
		SetOrderMatters(false).
		SetKeepAlive(viper.GetDuration("broker.keepalive"))

	if viper.GetString("broker.username") != "" {
		destOpts.SetUsername(viper.GetString("broker.username"))
		destOpts.SetPassword(viper.GetString("broker.password"))
		destOpts.SetAutoReconnect(true)
	}

	destClient := mqtt.NewClient(destOpts)
	token := destClient.Connect()

	select {
	case <-ctx.Done():
		logger.InfoContext(ctx, "Context done before source broker connected")
		return nil, ctx.Err()
	case <-token.Done():
		// continue
	}

	if token.Error() != nil {
		logger.ErrorContext(ctx, "Failed to connect to destination broker", slogtool.ErrorAttr(token.Error()))
		return nil, token.Error()
	}

	return destClient, nil
}
