package relay

import (
	"time"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/store"
)

// Config holds the relay configuration.
type Config struct {
	Broker    string
	ClientID  string
	Topic     string
	Username  string
	Password  string
	Store     store.Store
	DryRun    bool
	Keepalive time.Duration
}
