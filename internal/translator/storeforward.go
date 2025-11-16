package translator

import (
	"encoding/json"

	"github.com/na4ma4/meshtastic-mqtt-translate/pkg/meshtastic"
)

type StoreForwardApp struct {
	// TODO: REPLACE
	Rr string `json:"rr,omitempty"`
	// TODO: REPLACE
	//
	// Types that are valid to be assigned to Variant:
	//
	//	*StoreAndForward_Stats
	//	*StoreAndForward_History_
	//	*StoreAndForward_Heartbeat_
	//	*StoreAndForward_Text
	Stats     *StoreForwardStatistics `json:"stats,omitempty"`
	History   *StoreForwardHistory    `json:"history,omitempty"`
	Heartbeat *StoreForwardHeartbeat  `json:"heartbeat,omitempty"`
	Text      []byte                  `json:"text,omitempty"`
}

func NewStoreForwardApp(in *meshtastic.StoreAndForward) *StoreForwardApp {
	if in == nil {
		return nil
	}
	return &StoreForwardApp{
		Rr:        in.GetRr().String(),
		Stats:     NewStoreForwardStatistics(in.GetStats()),
		History:   NewStoreForwardHistory(in.GetHistory()),
		Heartbeat: NewStoreForwardHeartbeat(in.GetHeartbeat()),
		Text:      in.GetText(),
	}
}

func (p *StoreForwardApp) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

type StoreForwardHeartbeat struct {
	// Period in seconds that the heartbeat is sent out that will be sent to the client
	Period uint32 `json:"period,omitempty"`
	// If set, this is not the primary Store & Forward router on the mesh
	Secondary uint32 `json:"secondary,omitempty"`
}

func NewStoreForwardHeartbeat(in *meshtastic.StoreAndForward_Heartbeat) *StoreForwardHeartbeat {
	if in == nil {
		return nil
	}
	return &StoreForwardHeartbeat{
		Period:    in.GetPeriod(),
		Secondary: in.GetSecondary(),
	}
}

type StoreForwardStatistics struct {
	// Number of messages we have ever seen
	MessagesTotal uint32 `json:"messages_total,omitempty"`
	// Number of messages we have currently saved our history.
	MessagesSaved uint32 `json:"messages_saved,omitempty"`
	// Maximum number of messages we will save
	MessagesMax uint32 `json:"messages_max,omitempty"`
	// Router uptime in seconds
	UpTime uint32 `json:"up_time,omitempty"`
	// Number of times any client sent a request to the S&F.
	Requests uint32 `json:"requests,omitempty"`
	// Number of times the history was requested.
	RequestsHistory uint32 `json:"requests_history,omitempty"`
	// Is the heartbeat enabled on the server?
	Heartbeat bool `json:"heartbeat,omitempty"`
	// Maximum number of messages the server will return.
	ReturnMax uint32 `json:"return_max,omitempty"`
	// Maximum history window in minutes the server will return messages from.
	ReturnWindow uint32 `json:"return_window,omitempty"`
}

func NewStoreForwardStatistics(in *meshtastic.StoreAndForward_Statistics) *StoreForwardStatistics {
	if in == nil {
		return nil
	}
	return &StoreForwardStatistics{
		MessagesTotal:   in.GetMessagesTotal(),
		MessagesSaved:   in.GetMessagesSaved(),
		MessagesMax:     in.GetMessagesMax(),
		UpTime:          in.GetUpTime(),
		Requests:        in.GetRequests(),
		RequestsHistory: in.GetRequestsHistory(),
		Heartbeat:       in.GetHeartbeat(),
		ReturnMax:       in.GetReturnMax(),
		ReturnWindow:    in.GetReturnWindow(),
	}
}

type StoreForwardHistory struct {
	// Number of that will be sent to the client
	HistoryMessages uint32 `json:"history_messages,omitempty"`
	// The window of messages that was used to filter the history client requested
	Window uint32 `json:"window,omitempty"`
	// Index in the packet history of the last message sent in a previous request to the server.
	// Will be sent to the client before sending the history and can be set in a subsequent request to avoid getting packets the server already sent to the client.
	LastRequest uint32 `json:"last_request,omitempty"`
}

func NewStoreForwardHistory(in *meshtastic.StoreAndForward_History) *StoreForwardHistory {
	if in == nil {
		return nil
	}
	return &StoreForwardHistory{
		HistoryMessages: in.GetHistoryMessages(),
		Window:          in.GetWindow(),
		LastRequest:     in.GetLastRequest(),
	}
}
