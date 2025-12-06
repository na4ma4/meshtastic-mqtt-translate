package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/na4ma4/meshtastic-mqtt-translate/internal/mtypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type gormMessage struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid()"`
	NodeFrom  uint32
	NodeTo    uint32
	MessageID string
	PortNum   string
	Payload   []byte
	JSONData  *mtypes.Message `gorm:"type:jsonb"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name used by gormMessage to `messages`.
func (gormMessage) TableName() string {
	return "messages"
}

type GormStore struct {
	db     *gorm.DB
	Logger *slog.Logger
}

func newGormStore(in gorm.Dialector, cfg Config) (*GormStore, error) {
	gcfg := &gorm.Config{}

	if cfg.Logger != nil {
		gcfg.Logger = logger.NewSlogLogger(cfg.Logger, logger.Config{
			SlowThreshold: cfg.SlowThreshold,
			LogLevel:      logger.LogLevel(cfg.LogLevel),
		})
	}
	var db *gorm.DB
	{
		var err error
		db, err = gorm.Open(in, gcfg)
		if err != nil {
			return nil, fmt.Errorf("failed to open gorm DB: %w", err)
		}
	}

	if err := db.AutoMigrate(&gormMessage{}); err != nil {
		return nil, fmt.Errorf("failed to migrate gorm DB: %w", err)
	}

	return &GormStore{db: db}, nil
}

// func (s *GormStore) SaveOld(messageID, portNum string, payload, jsonData []byte) error {
// 	item := gormMessage{
// 		MessageID: messageID,
// 		PortNum:   portNum,
// 		Payload:   payload,
// 		JSONData:  jsonData,
// 	}
// 	return s.db.Create(&item).Error
// }

func (s *GormStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (s *GormStore) Save(ctx context.Context, messageID, portNum string, payload []byte, msg *mtypes.Message) error {
	item := gormMessage{
		MessageID: messageID,
		NodeFrom:  msg.GetFrom(),
		NodeTo:    msg.GetTo(),
		PortNum:   portNum,
		Payload:   payload,
		JSONData:  msg,
	}
	return s.db.WithContext(ctx).Create(&item).Error
}

func (s *GormStore) Get(ctx context.Context, messageID string) (*mtypes.Message, error) {
	item, err := gorm.G[gormMessage](s.db).Where("message_id = ?", messageID).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get message by ID %s: %w", messageID, err)
	}
	return item.JSONData, nil
}

func (s *GormStore) GetPayload(ctx context.Context, messageID string) ([]byte, error) {
	item, err := gorm.G[gormMessage](s.db).Where("message_id = ?", messageID).First(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get message by ID %s: %w", messageID, err)
	}
	return item.Payload, nil
}

func (s *GormStore) Iterate(ctx context.Context, f func(*mtypes.Message) error) error {
	var messages []gormMessage
	if err := s.db.WithContext(ctx).Order("created_at desc").Find(&messages); err != nil {
		return fmt.Errorf("failed to iterate messages: %w", err.Error)
	}
	for _, msg := range messages {
		if err := f(msg.JSONData); err != nil {
			return err
		}
	}
	return nil
}

// // JSONB Interface for JSONB Field of yourTableName Table
// type JSONB map[string]any

// // Value Marshal
// func (a JSONB) Value() (driver.Value, error) {
// 	return json.Marshal(a)
// }

// // Scan Unmarshal
// func (a *JSONB) Scan(value interface{}) error {
// 	b, ok := value.([]byte)
// 	if !ok {
// 		return errors.New("type assertion to []byte failed")
// 	}
// 	return json.Unmarshal(b, &a)
// }
