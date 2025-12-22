package usecases_test

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleText_CreateChatPayload(t *testing.T) {
	logger := test_helpers.CreateLogger()
	loc := test_helpers.CreateLocalizer(map[string]string{})
	settingsProvider := test_helpers.CreateSettingsProvider()
	commandService := test_helpers.CreateCommandService()
	startFlow := usecases.CreateStartFlow(logger, loc, settingsProvider, commandService)

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
			startFlow.Start(makeMessage(text), bot)
			wg.Done()
		}(message)
	}

	wg.Wait()

	assert.Equal(t, 1, len(settingsProvider.Data))

	message := makeMessage("/start")
	data, _ := settingsProvider.GetData(message.Chat.ID, core.SPayloadList)
	var settingsV1 struct {
		Payload []string
	}
	_ = json.Unmarshal(data, &settingsV1)
	assert.Equal(t, true, contains("payload", settingsV1.Payload))
	assert.Equal(t, true, contains("another_payload", settingsV1.Payload))

	expected := []string{
		"enable commands 42 [{help show help message}]",
		"enable commands 42 [{help show help message}]",
		"enable commands 42 [{help show help message}]",
	}
	assert.Equal(t, expected, commandService.ActionLog)
}

func makeMessage(text string) *core.Message {
	chat := core.Chat{ID: 42, Title: "Vinny The Pooh", Type: "private"}
	sender := core.User{ID: 1, FirstName: "Vinny", LastName: "The Pooh"}
	return &core.Message{Text: text, Chat: &chat, Sender: &sender}
}

func contains(message string, messages []string) bool {
	for _, m := range messages {
		if m == message {
			return true
		}
	}
	return false
}
