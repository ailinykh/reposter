package usecases

import (
	"strings"
	"sync"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateBootstrapFlow(l core.Logger, chatStorage legacy.IChatStorage, userStorage legacy.IUserStorage) legacy.ITextHandler {
	return &BootstrapFlow{l, chatStorage, userStorage, sync.Mutex{}}
}

type BootstrapFlow struct {
	l           core.Logger
	chatStorage legacy.IChatStorage
	userStorage legacy.IUserStorage
	lock        sync.Mutex
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *BootstrapFlow) HandleText(message *legacy.Message, bot legacy.IBot) error {
	flow.lock.Lock()
	defer flow.lock.Unlock()

	err := flow.ensureUserExists(message.Sender)
	if err != nil {
		flow.l.Error(err)
		//Do not return?
	}

	err = flow.ensureChatExists(message.Chat)
	if err != nil {
		flow.l.Error(err)
		//Do not return?
	}

	return err
}

func (flow *BootstrapFlow) ensureChatExists(chat *legacy.Chat) error {
	_, err := flow.chatStorage.GetChatByID(chat.ID)
	if err != nil {
		if err.Error() == "record not found" {
			return flow.chatStorage.CreateChat(chat.ID, chat.Title, chat.Type)
		}
		flow.l.Error(err)
	}
	return err
}

func (flow *BootstrapFlow) ensureUserExists(user *legacy.User) error {
	_, err := flow.userStorage.GetUserById(user.ID)
	if err != nil {
		if strings.HasSuffix(err.Error(), "record not found") {
			return flow.userStorage.CreateUser(user)
		}
		flow.l.Error(err)
	}
	return err
}
