package translator

import (
	"encoding/json"

	"github.com/na4ma4/meshtastic-mqtt-bin-to-json/pkg/meshtastic"
)

// type TracerouteApp struct {
// 	// The list of nodenums this packet has visited so far to the destination.
// 	Route []uint32 `json:"route,omitempty"`
// 	// The list of SNRs (in dB, scaled by 4) in the route towards the destination.
// 	SnrTowards []int32 `json:"snr_towards,omitempty"`
// 	// The list of nodenums the packet has visited on the way back from the destination.
// 	RouteBack []uint32 `json:"route_back,omitempty"`
// 	// The list of SNRs (in dB, scaled by 4) in the route back from the destination.
// 	SnrBack []int32 `json:"snr_back,omitempty"`
// }

type RouteHop struct {
	// Nodenum this packet has visited.
	Route uint32 `json:"route,omitempty"`
	// SNR (in dB, scaled by 4) in the route.
	Snr int32 `json:"snr,omitempty"`
}

type TracerouteApp struct {
	// Message in case of error or info
	Message string `json:"message,omitempty"`
	// The list of hops towards the destination.
	Towards []RouteHop `json:"towards,omitempty"`
	// The list of hops back from the destination.
	Back []RouteHop `json:"back,omitempty"`
}

// type RouteHops struct {
// 	// Nodenum this packet has visited so far to the destination.
// 	Route uint32 `json:"route,omitempty"`
// 	// SNR (in dB, scaled by 4) in the route towards the destination.
// 	SnrTowards int32 `json:"snr_towards,omitempty"`
// 	// Nodenum the packet has visited on the way back from the destination.
// 	RouteBack uint32 `json:"route_back,omitempty"`
// 	// SNR (in dB, scaled by 4) in the route back from the destination.
// 	SnrBack int32 `json:"snr_back,omitempty"`
// }

func NewTracerouteApp(in *meshtastic.RouteDiscovery) *TracerouteApp {
	if in == nil {
		return nil
	}

	// if len(in.GetRoute()) != len(in.GetSnrTowards()) || len(in.GetRouteBack()) != len(in.GetSnrBack()) {
	// 	return &TracerouteApp{
	// 		Message: fmt.Sprintf("inconsistent route and snr lengths: route=%d snr_towards=%d route_back=%d snr_back=%d", len(in.GetRoute()), len(in.GetSnrTowards()), len(in.GetRouteBack()), len(in.GetSnrBack())),
	// 	}
	// }

	// log.Printf("TracerouteApp: route=%v snr_towards=%v route_back=%v snr_back=%v", in.GetRoute(), in.GetSnrTowards(), in.GetRouteBack(), in.GetSnrBack())
	// log.Printf("TracerouteApp lengths: route=%d snr_towards=%d route_back=%d snr_back=%d", len(in.GetRoute()), len(in.GetSnrTowards()), len(in.GetRouteBack()), len(in.GetSnrBack()))

	out := &TracerouteApp{
		Towards: make([]RouteHop, max(len(in.GetRoute()), len(in.GetSnrTowards()))),
		Back:    make([]RouteHop, max(len(in.GetRouteBack()), len(in.GetSnrBack()))),
	}

	for i := range len(in.GetRoute()) {
		// log.Printf("Adding towards hop %d: route=%d", i, in.GetRoute()[i])
		out.Towards[i].Route = in.GetRoute()[i]
	}

	for i := range len(in.GetSnrTowards()) {
		// log.Printf("Adding towards hop %d: snr=%d", i, in.GetSnrTowards()[i])
		out.Towards[i].Snr = in.GetSnrTowards()[i]
	}

	for i := range len(in.GetRouteBack()) {
		// log.Printf("Adding back hop %d: route=%d", i, in.GetRouteBack()[i])
		out.Back[i].Route = in.GetRouteBack()[i]
	}

	for i := range len(in.GetSnrBack()) {
		// log.Printf("Adding back hop %d: snr=%d", i, in.GetSnrBack()[i])
		out.Back[i].Snr = in.GetSnrBack()[i]
	}

	return out
}

func (p *TracerouteApp) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
