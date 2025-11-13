package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/mainconfig"
	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/relay"
	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "meshtastic-mqtt-relay",
	Short: "Meshtastic MQTT Relay",
	Long:  `Meshtastic MQTT Relay is a tool for converting Meshtastic messages over MQTT from protobuf to JSON.`,
	RunE:  mainCmd,
	Args:  cobra.NoArgs,
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

	rootCmd.PersistentFlags().BoolP("dry-run", "n", false, "Dry run mode (optional)")
	_ = viper.BindPFlag("dry-run", rootCmd.PersistentFlags().Lookup("dry-run"))
	_ = viper.BindEnv("dry-run", "MQTT_DRY_RUN")

	rootCmd.PersistentFlags().StringP("output", "o", "", "Output store directory (optional)")
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindEnv("output", "OUTPUT_DIRECTORY")
}

func main() {
	_ = rootCmd.Execute()
}

func mainCmd(_ *cobra.Command, _ []string) error {
	log.Printf("Starting Meshtastic MQTT Relay")
	log.Printf("Source: %s (topic: %s)", viper.GetString("broker"), viper.GetString("topic"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if viper.GetBool("debug") {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
		logger.Debug("Debug logging enabled")
	}

	config := relay.Config{
		Broker:   viper.GetString("broker.address"),
		ClientID: viper.GetString("broker.clientid"),
		Username: viper.GetString("broker.username"),
		Password: viper.GetString("broker.password"),
		Topic:    viper.GetString("broker.topic"),
	}

	if outputDir := viper.GetString("output"); outputDir != "" {
		config.Store = store.NewJSONDirStore(outputDir)
		log.Printf("Output directory set to: %s", outputDir)
	}

	var client *relay.Relay
	{
		var err error
		client, err = relay.NewRelay(config, logger)
		if err != nil {
			log.Printf("Failed to create relay: %v", err)
			return err
		}
	}

	if err := client.Start(ctx); err != nil {
		log.Printf("Failed to start relay: %v", err)
		return err
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	client.Stop(ctx)

	return nil
}
