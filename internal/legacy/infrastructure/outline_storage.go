package infrastructure

import (
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CreateOutlineStorage is a default OutlineStorage factory
func CreateOutlineStorage(dbFile string, l core.Logger) *OutlineStorage {
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(err)
	}

	conn.AutoMigrate(&VpnKey{})
	return &OutlineStorage{conn, l}
}

type OutlineStorage struct {
	conn *gorm.DB
	l    core.Logger
}

type VpnKey struct {
	ID        string `gorm:"primaryKey"`
	ChatID    int64  `gorm:"primaryKey"`
	Host      string `gorm:"primaryKey"`
	Title     string
	Key       string
	CreatedAt time.Time `gorm:"autoUpdateTime"`
	UpdatedAt time.Time `gorm:"autoCreateTime"`
}

func (storage *OutlineStorage) GetKeys(chatID int64) ([]*VpnKey, error) {
	var keys []*VpnKey
	res := storage.conn.Where("chat_id = ?", chatID).Find(&keys)

	if res.Error != nil {
		storage.l.Error(res.Error)
		return nil, res.Error
	}

	return keys, nil
}

func (storage *OutlineStorage) CreateKey(id string, chatID int64, host string, title string, key string) error {
	storage.l.Info("creating key", "id", id, "chat_id", chatID, "host", host, "title", title, "key", key)
	k := VpnKey{
		ID:     id,
		ChatID: chatID,
		Host:   host,
		Title:  title,
		Key:    key,
	}
	res := storage.conn.Create(&k)
	return res.Error
}

func (storage *OutlineStorage) DeleteKey(key *legacy.VpnKey, host string) error {
	res := storage.conn.Delete(VpnKey{
		ID:     key.ID,
		ChatID: key.ChatID,
		Host:   host,
		Title:  key.Title,
		Key:    key.Key,
	})
	return res.Error
}
