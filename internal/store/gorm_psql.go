package store

import (
	"net/url"

	"gorm.io/driver/postgres"
)

// PostgresFactory is a Factory for PostgreSQL Directory Store.
type PostgresFactory struct{}

func (f PostgresFactory) Match(in *url.URL) bool {
	if in.Scheme == "postgres" ||
		in.Scheme == "postgresql" ||
		in.Scheme == "pg" ||
		in.Scheme == "psql" {
		in.Scheme = "postgresql"
		return true
	}

	return false
}

func (f PostgresFactory) NewStore(in *url.URL, cfg Config) (Store, error) {
	// Implementation for creating a PostgreSQL Store
	return NewPostgresStore(in, cfg)
}

func NewPostgresStore(dsn *url.URL, cfg Config) (*GormStore, error) {
	return newGormStore(postgres.Open(dsn.String()), cfg)
}
