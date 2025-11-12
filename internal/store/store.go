package store

type Store interface {
	Save(messageID, portNum string, payload, jsonData []byte) error
}
