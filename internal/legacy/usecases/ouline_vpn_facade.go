package usecases

import (
	"fmt"
	"net/url"

	"github.com/ailinykh/reposter/v2/internal/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/api"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/infrastructure"
)

func CreateOutlineVpnFacade(apiUrl string, dbFile string, l core.Logger, userStorage legacy.IUserStorage) legacy.IVpnAPI {
	u, err := url.Parse(apiUrl)
	if err != nil {
		panic(err)
	}

	api := api.CreateOutlineAPI(l, apiUrl)
	outlineStorage := infrastructure.CreateOutlineStorage(dbFile, l)
	return &OutlineVpnFacade{l, api, u.Host, outlineStorage, userStorage}
}

type OutlineVpnFacade struct {
	l              core.Logger
	api            *api.OutlineAPI
	host           string
	outlineStorage *infrastructure.OutlineStorage
	userStorage    legacy.IUserStorage
}

// GetKeys is a core.IVpnAPI interface implementation
func (facade *OutlineVpnFacade) GetKeys(chatID int64) ([]*legacy.VpnKey, error) {
	keys, err := facade.outlineStorage.GetKeys(chatID)
	if err != nil {
		facade.l.Error(err)
		return nil, err
	}

	keys2 := []*legacy.VpnKey{}
	for _, k := range keys {
		keys2 = append(keys2, &legacy.VpnKey{
			ID:     k.ID,
			ChatID: k.ChatID,
			Title:  k.Title,
			Key:    k.Key,
		})
	}
	return keys2, nil
}

// CreateKey is a core.IVpnAPI interface implementation
func (facade *OutlineVpnFacade) CreateKey(title string, chatID int64, user *legacy.User) (*legacy.VpnKey, error) {
	keys, err := facade.outlineStorage.GetKeys(chatID)
	if err != nil {
		facade.l.Error(err)
		return nil, err
	}

	key, err := facade.api.CreateKey(chatID, fmt.Sprintf("%s %d", user.DisplayName(), len(keys)))
	if err != nil {
		facade.l.Error(err)
		return nil, err
	}

	err = facade.outlineStorage.CreateKey(key.ID, chatID, facade.host, title, key.AccessURL)
	if err != nil {
		facade.l.Error(err)
		return nil, err
	}

	return &legacy.VpnKey{
		ID:     key.ID,
		ChatID: chatID,
		Title:  title,
		Key:    key.AccessURL,
	}, nil
}

// DeleteKey is a core.IVpnAPI interface implementation
func (facade *OutlineVpnFacade) DeleteKey(key *legacy.VpnKey) error {
	err := facade.api.DeleteKey(key)
	if err != nil {
		facade.l.Error(err)
		return err
	}

	return facade.outlineStorage.DeleteKey(key, facade.host)
}
