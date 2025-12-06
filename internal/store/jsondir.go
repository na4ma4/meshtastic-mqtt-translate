package store

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/mtypes"
)

// JSONDirFactory is a Factory for JSON Directory Store.
type JSONDirFactory struct{}

func (f JSONDirFactory) Match(in *url.URL) bool {
	return in.Scheme == "jsondir" ||
		in.Scheme == "dir" ||
		in.Scheme == "file"
}

func (f JSONDirFactory) NewStore(in *url.URL, cfg Config) (Store, error) {
	// Implementation for creating a JSON Directory Store
	return NewJSONDirStore(in.Path, cfg), nil
}

type JSONDirStoreConfig struct {
	Logger    *slog.Logger
	Directory string
}

type JSONDirStore struct {
	config JSONDirStoreConfig
}

func NewJSONDirStore(storeDir string, _ Config) *JSONDirStore {
	return &JSONDirStore{config: JSONDirStoreConfig{Directory: storeDir}}
}

func (s *JSONDirStore) Close() error {
	return nil
}

func (s *JSONDirStore) SaveOld(messageID, portNum string, payload, jsonData []byte) error {
	{
		fileName := path.Join(s.config.Directory, fmt.Sprintf("%s_%s.enc", messageID, portNum))
		if err := writeFileAtomic(fileName, payload); err != nil {
			return fmt.Errorf("failed to write Encoded file %s: %w", fileName, err)
		}
	}

	{
		fileName := path.Join(s.config.Directory, fmt.Sprintf("%s_%s.json", messageID, portNum))
		if err := writeFileAtomic(fileName, jsonData); err != nil {
			return fmt.Errorf("failed to write JSON file %s: %w", fileName, err)
		}
	}

	return nil
}

func (s *JSONDirStore) Save(_ context.Context, messageID, portNum string, payload []byte, msg *mtypes.Message) error {
	{
		fileName := path.Join(s.config.Directory, fmt.Sprintf("%s_%s.enc", messageID, portNum))
		if err := writeFileAtomic(fileName, payload); err != nil {
			return fmt.Errorf("failed to write Encoded file %s: %w", fileName, err)
		}
	}

	var jsonData []byte
	{
		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(msg); err != nil {
			return fmt.Errorf("failed to encode JSON data: %w", err)
		}
		jsonData = buf.Bytes()
	}

	{
		fileName := path.Join(s.config.Directory, fmt.Sprintf("%s_%s.json", messageID, portNum))
		if err := writeFileAtomic(fileName, jsonData); err != nil {
			return fmt.Errorf("failed to write JSON file %s: %w", fileName, err)
		}
	}

	return nil
}

func (s *JSONDirStore) Get(_ context.Context, _ string) (*mtypes.Message, error) {
	// Implementation for retrieving a message by ID
	return nil, errors.New("Get method not implemented")
}

func (s *JSONDirStore) GetPayload(_ context.Context, _ string) ([]byte, error) {
	// Implementation for retrieving a message payload by ID
	return nil, errors.New("GetPayload method not implemented")
}

func (s *JSONDirStore) Iterate(_ context.Context, _ func(*mtypes.Message) error) error {
	// Implementation for iterating over all messages
	return errors.New("Iterate method not implemented")
}

func writeFileAtomic(filename string, data []byte) error {
	var tmpFile *os.File
	{
		var err error
		tmpFile, err = os.CreateTemp(path.Dir(filename), "tmpfile")
		if err != nil {
			return err
		}
		defer os.Remove(tmpFile.Name())
	}

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpFile.Name(), filename)
}
