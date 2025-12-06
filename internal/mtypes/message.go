package mtypes

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/translator"
)

type Message struct {
	Bitfield  *uint32                   `json:"bitfield,omitempty"`
	Channel   uint32                    `json:"channel"`
	From      uint32                    `json:"from"`
	HopStart  uint32                    `json:"hop_start"`
	HopsAway  uint32                    `json:"hops_away"`
	ID        uint32                    `json:"id"`
	Payload   any                       `json:"payload"`
	RSSI      int32                     `json:"rssi"`
	Sender    string                    `json:"sender"`
	SNR       translator.SpecialFloat64 `json:"snr"`
	Timestamp uint32                    `json:"timestamp"`
	To        uint32                    `json:"to"`
	Type      string                    `json:"type"`
}

func (m *Message) GetFrom() uint32 {
	return m.From
}

func (m *Message) GetTo() uint32 {
	return m.To
}

// Value Marshal.
func (m *Message) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan Unmarshal.
func (m *Message) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, m)
}
