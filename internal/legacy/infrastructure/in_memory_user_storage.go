package infrastructure

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateInMemoryUserStorage() core.IUserStorage {
	return &InMemoryUserStorage{make(map[int64]*core.User)}
}

type InMemoryUserStorage struct {
	cache map[int64]*core.User
}

// GetUserById is a core.IUserStorage interface implementation
func (storage *InMemoryUserStorage) GetUserById(userID int64) (*core.User, error) {
	if user, ok := storage.cache[userID]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("record not found")
}

// CreateUser is a core.IUserStorage interface implementation
func (storage *InMemoryUserStorage) CreateUser(user *core.User) error {
	storage.cache[user.ID] = user
	return nil
}
