package fanout

import (
	"time"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/relay"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/store"
)

// Config holds the relay configuration.
type Config struct {
	Broker          string
	ClientID        string
	SourceTopic     string
	Username        string
	Password        string
	Store           store.Store
	DryRun          bool
	Keepalive       time.Duration
	TargetBaseTopic string
	RetainFlag      bool
}

func (c *Config) CopyFromRelayConfig(src relay.Config) {
	c.Broker = src.Broker
	c.ClientID = src.ClientID
	c.Username = src.Username
	c.Password = src.Password
	c.SourceTopic = src.Topic
	c.Store = src.Store
	c.DryRun = src.DryRun
	c.Keepalive = src.Keepalive
}
