package store

import (
	"net/url"

	"gorm.io/driver/mysql"
)

// MySQLFactory is a Factory for MySQL Directory Store.
type MySQLFactory struct{}

func (f MySQLFactory) Match(in *url.URL) bool {
	return in.Scheme == "mysql" || in.Scheme == "mariadb"
}

func (f MySQLFactory) NewStore(in *url.URL, cfg Config) (Store, error) {
	// Implementation for creating a MySQL Store
	return NewMySQLStore(in, cfg)
}

func NewMySQLStore(dsn *url.URL, cfg Config) (*GormStore, error) {
	return newGormStore(mysql.Open(dsn.String()), cfg)
}
