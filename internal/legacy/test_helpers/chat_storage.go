package test_helpers

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateChatStorage() *FakeChatStorage {
	return &FakeChatStorage{make(map[int64]*core.Chat), nil}
}

type FakeChatStorage struct {
	Chats map[int64]*core.Chat
	Err   error
}

// GetChatByID is a core.IChatStorage interface implementation
func (storage *FakeChatStorage) GetChatByID(chatID int64) (*core.Chat, error) {
	if user, ok := storage.Chats[chatID]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("record not found")
}

// CreateChat is a core.IChatStorage interface implementation
func (s *FakeChatStorage) CreateChat(chatID int64, title string, type_ string) error {
	s.Chats[chatID] = &core.Chat{ID: chatID, Title: title, Type: type_}
	return nil
}
