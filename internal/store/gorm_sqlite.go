package store

import (
	"net/url"

	"gorm.io/driver/sqlite"
)

// SQLiteFactory is a Factory for SQLite Directory Store.
type SQLiteFactory struct{}

func (f SQLiteFactory) Match(in *url.URL) bool {
	return in.Scheme == "sqlite" ||
		in.Scheme == "sqlite3"
}

func (f SQLiteFactory) NewStore(in *url.URL, cfg Config) (Store, error) {
	// Implementation for creating a SQLite Store
	return NewSQLiteStore(in, cfg)
}

func NewSQLiteStore(dsn *url.URL, cfg Config) (*GormStore, error) {
	return newGormStore(sqlite.Open(dsn.String()), cfg)
}
