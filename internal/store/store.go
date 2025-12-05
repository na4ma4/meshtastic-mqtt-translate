package store

import (
	"context"
	"database/sql/driver"
	"log/slog"
	"time"
)

type MessageType interface {
	Value() (driver.Value, error)
	Scan(src any) error
	GetFrom() uint32
	GetTo() uint32
}

type Store interface {
	// Save(messageID, portNum string, payload, jsonData []byte) error
	Save(ctx context.Context, messageID, portNum string, payload []byte, jsonData MessageType) error
	Get(ctx context.Context, messageID string) (MessageType, error)
	Iterate(ctx context.Context, f func(MessageType) error) error
	Close() error
}

type Config struct {
	SlowThreshold time.Duration
	LogLevel      slog.Level
	Logger        *slog.Logger
}

// type Logger interface {
// 	LogMode() Logger
// 	Info(context.Context, string, ...interface{})
// 	Warn(context.Context, string, ...interface{})
// 	Error(context.Context, string, ...interface{})
// 	Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
// }
