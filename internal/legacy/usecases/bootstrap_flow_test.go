package usecases_test

import (
	"sync"
	"testing"

	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleText_CreateUserData(t *testing.T) {
	logger := test_helpers.CreateLogger()
	chatStorage := test_helpers.CreateChatStorage()
	userStorage := test_helpers.CreateUserStorage()
	bootstrapFlow := usecases.CreateBootstrapFlow(logger, chatStorage, userStorage)

	bot := test_helpers.CreateBot()

	messages := []string{
		"/start",
		"/start payload",
		"/start another_payload",
	}
	wg := sync.WaitGroup{}

	for _, message := range messages {
		wg.Add(1)
		go func(text string) {
			bootstrapFlow.HandleText(makeMessage(text), bot)
			wg.Done()
		}(message)
	}

	wg.Wait()

	assert.Equal(t, 1, len(userStorage.Users))

	message := makeMessage("/start")
	user, _ := userStorage.GetUserById(message.Sender.ID)
	assert.Equal(t, message.Sender, user)
}

func Test_HandleText_CreateChatData(t *testing.T) {
	logger := test_helpers.CreateLogger()
	chatStorage := test_helpers.CreateChatStorage()
	userStorage := test_helpers.CreateUserStorage()
	bootstrapFlow := usecases.CreateBootstrapFlow(logger, chatStorage, userStorage)

	bot := test_helpers.CreateBot()

	messages := []string{
		"/start",
		"/some_command",
		"some text message",
	}
	wg := sync.WaitGroup{}

	for _, message := range messages {
		wg.Add(1)
		go func(text string) {
			bootstrapFlow.HandleText(makeMessage(text), bot)
			wg.Done()
		}(message)
	}

	wg.Wait()

	assert.Equal(t, 1, len(chatStorage.Chats))
}
