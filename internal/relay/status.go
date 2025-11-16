package relay

type Status struct {
	SourceBrokerConnected bool `json:"source_broker_connected"`
	DestBrokerConnected   bool `json:"dest_broker_connected"`
	Status                bool `json:"status"`
}

func (r *Relay) GetStatus() Status {
	return Status{
		SourceBrokerConnected: r.sourceClient.IsConnected(),
		DestBrokerConnected:   r.destClient.IsConnected(),
		Status:                r.sourceClient.IsConnected() && r.destClient.IsConnected(),
	}
}
