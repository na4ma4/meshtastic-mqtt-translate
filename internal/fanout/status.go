package fanout

type Status struct {
	SourceBrokerConnected bool `json:"source_broker_connected"`
	DestBrokerConnected   bool `json:"dest_broker_connected"`
	Status                bool `json:"status"`
}

func (f *Fanout) GetStatus() Status {
	sourceConnected := f.sourceClient != nil && f.sourceClient.IsConnected()
	destConnected := f.destClient != nil && f.destClient.IsConnected()
	return Status{
		SourceBrokerConnected: sourceConnected,
		DestBrokerConnected:   destConnected,
		Status:                sourceConnected && destConnected,
	}
}
