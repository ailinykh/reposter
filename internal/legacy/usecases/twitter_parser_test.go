package usecases_test

import (
	"fmt"
	"testing"

	"github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/test_helpers"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleText_NotFoundAnyLinkByDefault(t *testing.T) {
	parser, handler, bot := makeTwitterSUT()
	m := makeTweetMessage("a message without any links")

	parser.HandleText(m, bot)

	assert.Equal(t, []string{}, handler.tweets)
}

func Test_HandleText_FoundTweetLink(t *testing.T) {
	parser, handler, bot := makeTwitterSUT()
	m1 := makeTweetMessage("a message with https://twitter.com/username/status/123456")
	m2 := makeTweetMessage("a message with https://x.com/x/status/654321/photo/1")

	parser.HandleText(m1, bot)
	parser.HandleText(m2, bot)

	assert.Equal(t, []string{"123456", "654321"}, handler.tweets)
}

func Test_HandleText_FoundTweetLinkX(t *testing.T) {
	parser, handler, bot := makeTwitterSUT()
	m := makeTweetMessage("a message with https://x.com/username/status/123456")

	parser.HandleText(m, bot)

	assert.Equal(t, []string{"123456"}, handler.tweets)
}

func Test_HandleText_FoundMultipleTweetLinks(t *testing.T) {
	parser, handler, bot := makeTwitterSUT()
	m := makeTweetMessage("a message with https://twitter.com/username/status/123456 and https://x.com/username/status/789010 and some text")
	parser.HandleText(m, bot)

	assert.Equal(t, []string{"123456", "789010"}, handler.tweets)
}

func Test_HandleText_DoesNotRemoveOriginalMessage(t *testing.T) {
	parser, _, bot := makeTwitterSUT()
	m := makeTweetMessage("https://x.com/username/status/123456 and some other text")

	parser.HandleText(m, bot)

	assert.Equal(t, []string{}, bot.RemovedMessages)
}

func Test_HandleText_ReturnsErrorOnError(t *testing.T) {
	parser, handler, bot := makeTwitterSUT()
	m := makeTweetMessage("a message with https://x.com/username/status/123456")
	handler.err = fmt.Errorf("an error")

	err := parser.HandleText(m, bot)

	assert.Equal(t, "an error", err.Error())
}

func makeTwitterSUT() (*usecases.TwitterParser, *FakeTweetHandler, *test_helpers.FakeBot) {
	logger := test_helpers.CreateLogger()
	handler := &FakeTweetHandler{[]string{}, nil}
	parser := usecases.CreateTwitterParser(logger, handler)
	bot := test_helpers.CreateBot()
	return parser, handler, bot
}

func makeTweetMessage(text string) *core.Message {
	return &core.Message{ID: 0, Text: text}
}

type FakeTweetHandler struct {
	tweets []string
	err    error
}

// Process is a ITweetHandler protocol implementation
func (fth *FakeTweetHandler) Process(tweetID string, message *core.Message, bot core.IBot) error {
	fth.tweets = append(fth.tweets, tweetID)
	return fth.err
}
