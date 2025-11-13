package mainconfig_test

import (
	"os"
	"testing"

	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/internal/mainconfig"
	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	wd, _ := os.Getwd()
	t.Chdir("../..")
	mainconfig.ConfigInit()
	t.Chdir(wd)

	t.Logf("Using config file: %s", viper.ConfigFileUsed())

	t.Logf("Debug: %v", viper.AllSettings())

	channels := mainconfig.GetChannels()
	t.Logf("Channels: %v", channels)

	for _, channel := range channels {
		t.Logf("Channel[%s]: %s", channel.Name, channel.Key)
	}
}
