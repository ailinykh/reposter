package usecases

import (
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateChatStorageDecorator(cache core.IChatStorage, db core.IChatStorage) core.IChatStorage {
	return &ChatStorageDecorator{cache, db}
}

type ChatStorageDecorator struct {
	cache core.IChatStorage
	db    core.IChatStorage
}

// GetChatByID is a core.IChatStorage interface implementation
func (decorator *ChatStorageDecorator) GetChatByID(chatID int64) (*core.Chat, error) {
	chat, err := decorator.cache.GetChatByID(chatID)
	if err != nil {
		chat, err := decorator.db.GetChatByID(chatID)
		if err != nil {
			return nil, err
		}
		_ = decorator.cache.CreateChat(chat.ID, chat.Title, chat.Type)
		return chat, nil
	}
	return chat, nil
}

// CreateChat is a core.IChatStorage interface implementation
func (decorator *ChatStorageDecorator) CreateChat(chatID int64, title string, type_ string) error {
	_ = decorator.cache.CreateChat(chatID, title, type_)
	return decorator.db.CreateChat(chatID, title, type_)
}
