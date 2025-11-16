//nolint:protogetter // copying structures
package translator

import (
	"encoding/json"

	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"
)

type PositionApp struct {
	LatitudeI      *int32  `json:"latitude_i"`
	LongitudeI     *int32  `json:"longitude_i"`
	Altitude       *int32  `json:"altitude"`
	Time           uint32  `json:"time"`
	LocationSource string  `json:"location_source"`
	PDOP           uint32  `json:"PDOP"`
	GroundSpeed    *uint32 `json:"ground_speed"`
	GroundTrack    *uint32 `json:"ground_track"`
	SatsInView     uint32  `json:"sats_in_view"`
	PrecisionBits  uint32  `json:"precision_bits"`
}

func NewPositionApp(in *meshtastic.Position) *PositionApp {
	if in == nil {
		return nil
	}
	return &PositionApp{
		LatitudeI:      in.LatitudeI,
		LongitudeI:     in.LongitudeI,
		Altitude:       in.Altitude,
		Time:           in.Time,
		LocationSource: in.LocationSource.String(),
		PDOP:           in.PDOP,
		GroundSpeed:    in.GroundSpeed,
		GroundTrack:    in.GroundTrack,
		SatsInView:     in.SatsInView,
		PrecisionBits:  in.PrecisionBits,
	}
}

func (p *PositionApp) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
