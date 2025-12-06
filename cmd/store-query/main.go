package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dosquad/go-cliversion"
	"github.com/na4ma4/meshtastic-mqtt-translate/cmd/store-query/repeat"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/cmdconst"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/mainconfig"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:          "store-query <message-ID>",
	Short:        "Datastore Query Tool",
	Long:         `Datastore Query Tool is a tool for querying data stores.`,
	Args:         cobra.MinimumNArgs(0),
	Version:      cliversion.Get().VersionString(),
	SilenceUsage: true,
}

func init() {
	cobra.OnInitialize(mainconfig.ConfigInit)

	rootCmd.AddCommand(repeat.CmdRepeat)
}

func main() {
	if err := rootCmd.Execute(); errors.Is(err, cmdconst.ErrNoUsage) {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	} else if err != nil {
		_ = rootCmd.Usage()
		os.Exit(1)
	}
}
