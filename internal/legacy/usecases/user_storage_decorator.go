package usecases

import "github.com/ailinykh/reposter/v2/internal/legacy/core"

func CreateUserStorageDecorator(primary core.IUserStorage, secondary core.IUserStorage) core.IUserStorage {
	return &UserStorageDecorator{primary, secondary}
}

type UserStorageDecorator struct {
	cache core.IUserStorage
	db    core.IUserStorage
}

// GetUserById is a core.IUserStorage interface implementation
func (decorator *UserStorageDecorator) GetUserById(userID int64) (*core.User, error) {
	user, err := decorator.cache.GetUserById(userID)
	if err != nil {
		user, err := decorator.db.GetUserById(userID)
		if err != nil {
			return nil, err
		}
		_ = decorator.cache.CreateUser(user)
		return user, err
	}
	return user, err
}

// CreateUser is a core.IUserStorage interface implementation
func (decorator *UserStorageDecorator) CreateUser(user *core.User) error {
	_ = decorator.cache.CreateUser(user)
	return decorator.db.CreateUser(user)
}
