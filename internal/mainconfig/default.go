package mainconfig

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	defaultHealthCheckPort = 8099
)

// ConfigInit is the common config initialisation for the commands.
func ConfigInit() {
	viper.SetConfigName("options")
	viper.SetConfigType("json")
	viper.AddConfigPath("./artifacts")
	viper.AddConfigPath("./testdata")
	viper.AddConfigPath("$HOME/.meshtastic-mqtt-relay")
	viper.AddConfigPath("/etc/meshtastic-mqtt-relay")
	viper.AddConfigPath("/usr/local/meshtastic-mqtt-relay/etc")
	viper.AddConfigPath("/data")
	viper.AddConfigPath(".")

	viper.SetDefault("store.slow-threshold", "1s")
	_ = viper.BindEnv("store.slow-threshold", "STORE_SLOW_THRESHOLD")

	viper.SetDefault("broker.keepalive", "1m")
	_ = viper.BindEnv("broker.keepalive", "BROKER_KEEPALIVE")

	viper.SetDefault("healthcheck.port", defaultHealthCheckPort)
	_ = viper.BindEnv("healthcheck.port", "HEALTHCHECK_PORT")

	viper.SetDefault("features.fanout-relay", false)
	_ = viper.BindEnv("features.fanout-relay", "FEATURE_FANOUT_RELAY")

	viper.SetDefault("features.message-store", false)
	_ = viper.BindEnv("features.message-store", "FEATURE_MESSAGE_STORE")

	// check if /data/optons.json or ./testdata/options.json exists.
	if _, err := os.Stat("./testdata/options.json"); err == nil {
		viper.SetConfigFile("./testdata/options.json")
	}
	if _, err := os.Stat("/data/options.json"); err == nil {
		viper.SetConfigFile("/data/options.json")
	}

	_ = viper.BindEnv("debug", "DEBUG")

	if err := configReadIn(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
	}
}

func configReadIn() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("unable to read config: %w", err)
	}

	return nil
}

type Channel struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

func GetChannels() []Channel {
	var channels []Channel

	channelsRaw := viper.Get("channels")
	if channelsRaw == nil {
		return channels
	}

	switch ch := channelsRaw.(type) {
	case []any:
		for _, channel := range ch {
			if c := parseChannelFromMap(channel, ""); c != nil {
				channels = append(channels, *c)
			}
		}
	case map[string]any:
		for name, value := range ch {
			if c := parseChannelFromMap(value, name); c != nil {
				channels = append(channels, *c)
			}
		}
	}

	return channels
}

// parseChannelFromMap extracts a Channel from a map[string]any.
// If nameOverride is provided, it's used instead of the "name" field in the map.
func parseChannelFromMap(raw any, nameOverride string) *Channel {
	channelMap, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	key, ok := channelMap["key"].(string)
	if !ok {
		return nil
	}

	name := nameOverride
	if name == "" {
		name, _ = channelMap["name"].(string)
	}

	if name == "" {
		return nil
	}

	return &Channel{Name: name, Key: key}
}
