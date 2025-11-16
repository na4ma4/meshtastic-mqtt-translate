package translator

import (
	"encoding/json"

	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"
)

type RoutingApp struct {
	// Types that are valid to be assigned to Variant:
	//
	//	*Routing_RouteRequest
	//	*Routing_RouteReply
	//	*Routing_ErrorReason
	// Variant       isRouting_Variant `protobuf_oneof:"variant"`
	RouteRequest *TracerouteApp      `json:"routeRequest,omitempty"`
	RouteReply   *TracerouteApp      `json:"routeReply,omitempty"`
	ErrorReason  *RoutingErrorReason `json:"errorReason,omitempty"`
}

func NewRoutingApp(in *meshtastic.Routing) *RoutingApp {
	if in == nil {
		return nil
	}
	out := &RoutingApp{}
	switch variant := in.GetVariant().(type) {
	case *meshtastic.Routing_RouteRequest:
		out.RouteRequest = NewTracerouteApp(variant.RouteRequest)
	case *meshtastic.Routing_RouteReply:
		out.RouteReply = NewTracerouteApp(variant.RouteReply)
	case *meshtastic.Routing_ErrorReason:
		out.ErrorReason = NewRoutingErrorReason(variant)
	}
	return out
}

func (p *RoutingApp) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

type RoutingErrorReason struct {
	Reason string `json:"reason,omitempty"`
}

func NewRoutingErrorReason(in *meshtastic.Routing_ErrorReason) *RoutingErrorReason {
	return &RoutingErrorReason{
		Reason: in.ErrorReason.String(),
	}
}

func (p *RoutingErrorReason) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
