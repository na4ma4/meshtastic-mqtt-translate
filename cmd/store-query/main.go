package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"github.com/dosquad/go-cliversion"
	"github.com/na4ma4/go-contextual"
	"github.com/na4ma4/go-slogtool"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/mainconfig"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "store-query",
	Short:        "Datastore Query Tool",
	Long:         `Datastore Query Tool is a tool for querying data stores.`,
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

	rootCmd.PersistentFlags().StringP("dsn", "o", "", "Datastore DSN")
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

	logger.InfoContext(ctx, "Starting Datastore Query Tool")

	var st store.Store
	{
		var err error
		if st, err = getStore(viper.GetString("store.dsn"), storeCfg); err != nil && !errors.Is(err, ErrEmptyDSN) {
			logger.ErrorContext(ctx, "Failed to create store", slogtool.ErrorAttr(err))
			return fmt.Errorf("%w%w", ErrNoUsage, err)
		} else if errors.Is(err, ErrEmptyDSN) {
			logger.DebugContext(ctx, "No Store DSN set, exiting")
			return err
		} else if st == nil {
			logger.ErrorContext(ctx, "Store is nil")
			return fmt.Errorf("%wstore is nil", ErrNoUsage)
		}
	}

	if err := st.Iterate(ctx, func(mt store.MessageType) error {
		jsonData, err := mt.Value()
		if err != nil {
			return fmt.Errorf("failed to get JSON data: %w", err)
		}
		fmt.Fprintf(os.Stdout, "%s\n", jsonData)
		return nil
	}); err != nil {
		logger.ErrorContext(ctx, "Failed to iterate store", slogtool.ErrorAttr(err))
		return fmt.Errorf("%w%w", ErrNoUsage, err)
	}

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
