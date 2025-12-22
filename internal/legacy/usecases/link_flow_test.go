package usecases_test

import (
	"testing"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleUrl_ConvertsVideoUrlToVideo(t *testing.T) {
	bot := test_helpers.CreateBot()
	logger := test_helpers.CreateLogger()
	http_client := test_helpers.CreateHttpClient()
	media_factory := test_helpers.CreateMediaFactory()
	send_message_strategy := test_helpers.CreateSendMediaStrategy()
	link_flow := usecases.CreateLinkFlow(logger, http_client, media_factory, send_message_strategy)

	url := "http://an-url.com"
	http_client.ContentTypeForURL[url] = "video"
	message := core.Message{Text: url, Sender: &core.User{Username: "Username"}}
	err := link_flow.HandleText(&message, bot)

	assert.Equal(t, nil, err)
	assert.Equal(t, []core.URL{url}, media_factory.URLs)
	assert.Equal(t, []string{url}, send_message_strategy.SentMedia)
}
