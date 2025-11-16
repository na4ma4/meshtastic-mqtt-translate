package store

import "net/url"

func MustURL(rawurl string) *url.URL {
	parsed, err := url.Parse(rawurl)
	if err != nil {
		panic(err)
	}
	return parsed
}

func SanitizeURL(in *url.URL) *url.URL {
	sanitized := *in
	if sanitized.User != nil {
		sanitized.User = url.UserPassword(sanitized.User.Username(), "****")
	}
	return &sanitized
}

// Error represents an error related to store operations.
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// ErrUnsupportedStore is returned when the store type is not supported.
var ErrUnsupportedStore = &Error{"unsupported store type"}

// Factory defines the interface for store factories.
type Factory interface {
	Match(in *url.URL) bool
	NewStore(in *url.URL, cfg Config) (Store, error)
}

// NewDetectStore creates a Store based on the provided URL by detecting the appropriate store type.
func NewDetectStore(in *url.URL, cfg Config) (Store, error) {
	// storeFactories holds the registered store factories.
	var storeFactories = []Factory{
		JSONDirFactory{},
		SQLiteFactory{},
		MySQLFactory{},
		PostgresFactory{},
	}

	for _, factory := range storeFactories {
		if factory.Match(in) {
			return factory.NewStore(in, cfg)
		}
	}
	return nil, ErrUnsupportedStore
}
