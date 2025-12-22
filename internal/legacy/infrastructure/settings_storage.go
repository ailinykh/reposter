package infrastructure

import (
	"time"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Settings
type Settings struct {
	ChatID    int64  `gorm:"primaryKey"`
	Key       string `gorm:"primaryKey"`
	Data      []byte
	CreatedAt time.Time `gorm:"autoUpdateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}

func CreateSettingsStorage(dbFile string) core.ISettingsProvider {
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(err)
	}

	conn.AutoMigrate(&Settings{})

	return &SettingsStorage{conn}
}

// SettingsStorage implements core.ISettingsProvider interface
type SettingsStorage struct {
	conn *gorm.DB
}

// GetData is a core.ISettingsProvider interface implementation
func (storage *SettingsStorage) GetData(chatID core.ChatID, key core.SettingKey) ([]byte, error) {
	var sessings Settings
	err := storage.conn.First(&sessings, chatID, key).Error
	if err != nil {
		return nil, err
	}
	return sessings.Data, nil
}

// SetData is a core.ISettingsProvider interface implementation
func (storage *SettingsStorage) SetData(chatID core.ChatID, key core.SettingKey, data []byte) error {
	settings := Settings{
		ChatID: chatID,
		Key:    string(key),
		Data:   data,
	}
	return storage.conn.Save(&settings).Error
}
