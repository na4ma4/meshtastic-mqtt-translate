package translator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"
)

type User struct {
	ID             string `json:"id,omitempty"`
	LongName       string `json:"long_name,omitempty"`
	ShortName      string `json:"short_name,omitempty"`
	Macaddr        string `json:"macaddr,omitempty"`
	HwModel        string `json:"hw_model,omitempty"`
	IsLicensed     bool   `json:"is_licensed,omitempty"`
	Role           string `json:"role,omitempty"`
	PublicKey      []byte `json:"public_key,omitempty"`
	IsUnmessagable *bool  `json:"is_unmessagable,omitempty"`
}

func NewUser(in *meshtastic.User) *User {
	if in == nil {
		return nil
	}
	return &User{
		ID:             in.GetId(),
		LongName:       in.GetLongName(),
		ShortName:      in.GetShortName(),
		Macaddr:        formatMACAddr(in.GetMacaddr()), //nolint:staticcheck // deprecated field
		HwModel:        in.GetHwModel().String(),
		IsLicensed:     in.GetIsLicensed(),
		Role:           in.GetRole().String(),
		PublicKey:      in.GetPublicKey(),
		IsUnmessagable: in.IsUnmessagable,
	}
}

func (p *User) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

func formatMACAddr(mac []byte) string {
	if len(mac) == 0 {
		return ""
	}

	// Format the bytes into a MAC address (e.g., AA:BB:CC:DD:EE:FF)
	parts := make([]string, len(mac))
	for i, b := range mac {
		parts[i] = fmt.Sprintf("%02X", b)
	}
	return strings.Join(parts, ":")
}
