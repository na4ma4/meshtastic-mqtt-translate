package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/dosquad/go-cliversion"
	"github.com/na4ma4/go-contextual"
	"github.com/na4ma4/go-slogtool"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/fanout"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/health"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/mainconfig"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/relay"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "meshtastic-mqtt-relay",
	Short:        "Meshtastic MQTT Relay",
	Long:         `Meshtastic MQTT Relay is a tool for converting Meshtastic messages over MQTT from protobuf to JSON.`,
	RunE:         mainCmd,
	Args:         cobra.NoArgs,
	Version:      cliversion.Get().VersionString(),
	SilenceUsage: true,
}

func init() {
	cobra.OnInitialize(mainconfig.ConfigInit)

	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Debug output")
	_ = viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	_ = viper.BindEnv("debug", "DEBUG")

	rootCmd.PersistentFlags().StringP("broker", "b", "tcp://localhost:1883", "Source MQTT broker URL")
	_ = viper.BindPFlag("broker.address", rootCmd.PersistentFlags().Lookup("broker"))
	_ = viper.BindEnv("broker.address", "MQTT_BROKER")

	rootCmd.PersistentFlags().StringP("clientid", "c", "meshtastic-mqtt-relay", "MQTT client ID")
	_ = viper.BindPFlag("broker.clientid", rootCmd.PersistentFlags().Lookup("clientid"))
	_ = viper.BindEnv("broker.clientid", "MQTT_CLIENTID")

	rootCmd.PersistentFlags().StringP("username", "u", "", "MQTT username (optional)")
	_ = viper.BindPFlag("broker.username", rootCmd.PersistentFlags().Lookup("username"))
	_ = viper.BindEnv("broker.username", "MQTT_USERNAME")

	rootCmd.PersistentFlags().StringP("password", "p", "", "MQTT password (optional)")
	_ = viper.BindPFlag("broker.password", rootCmd.PersistentFlags().Lookup("password"))
	_ = viper.BindEnv("broker.password", "MQTT_PASSWORD")

	rootCmd.PersistentFlags().StringP("topic", "t", "msh/ANZ/2/e/#", "MQTT topic to subscribe to")
	_ = viper.BindPFlag("broker.topic", rootCmd.PersistentFlags().Lookup("topic"))
	_ = viper.BindEnv("broker.topic", "MQTT_TOPIC")

	rootCmd.PersistentFlags().StringP("fanout-topic", "f", "msh/ANZ/fanout/", "Fanout MQTT topic parent (optional)")
	_ = viper.BindPFlag("fanout.topic", rootCmd.PersistentFlags().Lookup("fanout-topic"))
	_ = viper.BindEnv("fanout.topic", "FANOUT_TOPIC")

	rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "Dry run mode (optional)")
	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindEnv("dry-run", "MQTT_DRY_RUN")

	rootCmd.PersistentFlags().StringP("dsn", "o", "", "Data store DSN (optional)")
	_ = viper.BindPFlag("store.dsn", rootCmd.PersistentFlags().Lookup("dsn"))
	_ = viper.BindEnv("store.dsn", "STORE_DSN")
}

var ErrNoUsage = errors.New("")

func main() {
	if err := rootCmd.Execute(); errors.Is(err, ErrNoUsage) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	} else if err != nil {
		_ = rootCmd.Usage()
		os.Exit(1)
	}
}

//nolint:funlen,gocognit // TODO refactor for simplicity
func mainCmd(_ *cobra.Command, _ []string) error {
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

	logger.InfoContext(ctx, "Starting Meshtastic MQTT Relay",
		slog.String("broker.address", viper.GetString("broker.address")),
		slog.String("broker.topic", viper.GetString("broker.topic")),
		slog.String("fanout.topic", viper.GetString("fanout.topic")),
		slog.Any("features", getFeatures()),
		slog.String("version", cliversion.Get().VersionString()),
		// slog.String("config.file", viper.ConfigFileUsed()),
	)

	config := relay.Config{
		Broker:    viper.GetString("broker.address"),
		ClientID:  viper.GetString("broker.clientid"),
		Username:  viper.GetString("broker.username"),
		Password:  viper.GetString("broker.password"),
		Topic:     viper.GetString("broker.topic"),
		DryRun:    viper.GetBool("dry-run"),
		Keepalive: viper.GetDuration("broker.keepalive"),
	}

	//nolint:nestif // TODO refactor for simplicity
	if viper.GetBool("features.message-store") {
		if st, err := getStore(viper.GetString("store.dsn"), storeCfg); err != nil && !errors.Is(err, ErrEmptyDSN) {
			logger.ErrorContext(ctx, "Failed to create store", slogtool.ErrorAttr(err))
			return fmt.Errorf("%w%w", ErrNoUsage, err)
		} else if errors.Is(err, ErrEmptyDSN) {
			logger.DebugContext(ctx, "No Store DSN set, not archiving messages")
		} else if st != nil {
			config.Store = st
			if logger.Enabled(ctx, slog.LevelInfo) {
				sanitizedDSN := store.SanitizeURL(store.MustURL(viper.GetString("store.dsn")))
				logger.InfoContext(ctx, "Store DSN set",
					slog.String("store.dsn", sanitizedDSN.String()),
				)
			}
		}
	} else {
		logger.InfoContext(ctx, "Message store feature disabled, not archiving messages")
	}

	var foClient *fanout.Fanout
	if viper.GetBool("features.fanout-relay") {
		foConfig := fanout.Config{
			TargetBaseTopic: viper.GetString("fanout.topic"),
		}
		foConfig.CopyFromRelayConfig(config)
		foConfig.ClientID += "-fanout"

		var err error
		foClient, err = fanout.NewFanout(ctx, foConfig, logger)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create fanout relay", slogtool.ErrorAttr(err))
			return fmt.Errorf("%w%w", ErrNoUsage, err)
		}
	}

	var client *relay.Relay
	{
		var err error
		client, err = relay.NewRelay(ctx, config, logger)
		if err != nil {
			logger.ErrorContext(ctx, "Failed to create relay", slogtool.ErrorAttr(err))
			return fmt.Errorf("%w%w", ErrNoUsage, err)
		}
	}

	healthErrChan, stopHealthServer := getHealthServer(ctx, logger, client, foClient)
	defer stopHealthServer()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	clientErrChan := client.Run(ctx)
	var foClientErrChan <-chan error
	if foClient != nil {
		foClientErrChan = foClient.Run(ctx)
	}
	defer client.Stop(ctx)
	if foClient != nil {
		defer foClient.Stop(ctx)
	}
	defer logger.InfoContext(ctx, "Shutting down Meshtastic MQTT Relay")

	for {
		select {
		case <-sigChan:
			return nil
		case err := <-clientErrChan:
			if err != nil {
				logger.ErrorContext(ctx, "Relay error", slogtool.ErrorAttr(err))
				return fmt.Errorf("%w%w", ErrNoUsage, err)
			}

			return nil
		case err := <-foClientErrChan:
			if err != nil {
				logger.ErrorContext(ctx, "Fanout client error", slogtool.ErrorAttr(err))
				return fmt.Errorf("%w%w", ErrNoUsage, err)
			}

			return nil
		case err := <-healthErrChan:
			if err != nil {
				logger.ErrorContext(ctx, "Health server error", slogtool.ErrorAttr(err))
				return fmt.Errorf("%w%w", ErrNoUsage, err)
			}

			return nil
		}
	}
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

func getHealthServer(
	ctx context.Context,
	logger *slog.Logger,
	client *relay.Relay,
	fanout *fanout.Fanout,
) (<-chan error, func()) {
	if viper.GetInt("healthcheck.port") > 0 {
		healthServer := health.NewServer(viper.GetInt("healthcheck.port"), logger, client, fanout)
		return healthServer.Start(), func() {
			if err := healthServer.Stop(ctx); err != nil {
				logger.ErrorContext(ctx, "Failed to stop health server", slogtool.ErrorAttr(err))
			}
		}
	}

	healthErrChan := make(chan error)
	return healthErrChan, func() {
		close(healthErrChan)
	}
}

func getFeatures() map[string]bool {
	features := make(map[string]bool)
	features["fanout-relay"] = viper.GetBool("features.fanout-relay")
	features["message-store"] = viper.GetBool("features.message-store")
	return features
}
