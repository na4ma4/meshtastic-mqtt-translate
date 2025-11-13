//nolint:protogetter // copying structures
package translator

import (
	"encoding/json"

	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/pkg/meshtastic"
)

type NodeInfoApp struct {
	Num                   uint32                    `json:"num,omitempty"`
	User                  *User                     `json:"user,omitempty"`
	Position              *PositionApp              `json:"position,omitempty"`
	Snr                   float64                   `json:"snr,omitempty"`
	LastHeard             uint32                    `json:"last_heard,omitempty"`
	DeviceMetrics         *meshtastic.DeviceMetrics `json:"device_metrics,omitempty"`
	Channel               uint32                    `json:"channel,omitempty"`
	ViaMqtt               bool                      `json:"via_mqtt,omitempty"`
	HopsAway              *uint32                   `json:"hops_away,omitempty"`
	IsFavorite            bool                      `json:"is_favorite,omitempty"`
	IsIgnored             bool                      `json:"is_ignored,omitempty"`
	IsKeyManuallyVerified bool                      `json:"is_key_manually_verified,omitempty"`
}

func NewNodeInfoApp(in *meshtastic.NodeInfo) *NodeInfoApp {
	if in == nil {
		return nil
	}
	return &NodeInfoApp{
		Num:                   in.Num,
		User:                  NewUser(in.User),
		Position:              NewPositionApp(in.Position),
		Snr:                   float64(in.Snr),
		LastHeard:             in.LastHeard,
		DeviceMetrics:         in.DeviceMetrics,
		Channel:               in.Channel,
		ViaMqtt:               in.ViaMqtt,
		HopsAway:              in.HopsAway,
		IsFavorite:            in.IsFavorite,
		IsIgnored:             in.IsIgnored,
		IsKeyManuallyVerified: in.IsKeyManuallyVerified,
	}
}

func (p *NodeInfoApp) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
