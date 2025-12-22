package test_helpers

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateUserStorage() *FakeUserStorage {
	return &FakeUserStorage{make(map[int64]*core.User), nil}
}

type FakeUserStorage struct {
	Users map[int64]*core.User
	Err   error
}

// GetUserById is a core.IUserStorage interface implementation
func (storage *FakeUserStorage) GetUserById(userID int64) (*core.User, error) {
	if user, ok := storage.Users[userID]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("failed to find user: record not found")
}

// CreateUser is a core.IUserStorage interface implementation
func (storage *FakeUserStorage) CreateUser(user *core.User) error {
	storage.Users[user.ID] = user
	return nil
}
