package store

import (
	"fmt"
	"os"
	"path"
)

type JSONDirStoreConfig struct {
	Directory string
}

type JSONDirStore struct {
	config JSONDirStoreConfig
}

func NewJSONDirStore(storeDir string) *JSONDirStore {
	return &JSONDirStore{config: JSONDirStoreConfig{Directory: storeDir}}
}

func (s *JSONDirStore) Save(messageID, portNum string, payload, jsonData []byte) error {
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
