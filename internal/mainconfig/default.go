package mainconfig

import (
	"fmt"

	"github.com/spf13/viper"
)

// ConfigInit is the common config initialisation for the commands.
func ConfigInit() {
	viper.SetConfigName("options")
	viper.SetConfigType("json")
	viper.AddConfigPath("./artifacts")
	viper.AddConfigPath("./test")
	viper.AddConfigPath("$HOME/.meshtastic-mqtt-relay")
	viper.AddConfigPath("/etc/meshtastic-mqtt-relay")
	viper.AddConfigPath("/usr/local/meshtastic-mqtt-relay/etc")
	viper.AddConfigPath("/data")
	viper.AddConfigPath(".")

	_ = configReadIn()
}

func configReadIn() error {
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("unable to read config: %w", err)
	}

	return nil
}
