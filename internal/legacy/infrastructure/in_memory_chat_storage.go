package infrastructure

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateInMemoryChatStorage() core.IChatStorage {
	return &InMemoryChatStorage{make(map[int64]*core.Chat)}
}

type InMemoryChatStorage struct {
	cache map[int64]*core.Chat
}

// GetChatByID is a core.IChatStorage interface implementation
func (storage *InMemoryChatStorage) GetChatByID(chatID int64) (*core.Chat, error) {
	if chat, ok := storage.cache[chatID]; ok {
		return chat, nil
	}
	return nil, fmt.Errorf("record not found")
}

// CreateChat is a core.IChatStorage interface implementation
func (storage *InMemoryChatStorage) CreateChat(chatID int64, title string, type_ string) error {
	storage.cache[chatID] = &core.Chat{
		ID:    chatID,
		Title: title,
		Type:  type_,
	}
	return nil
}
