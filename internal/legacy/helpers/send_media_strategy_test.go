package helpers_test

import (
	"testing"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/stretchr/testify/assert"
)

func Test_SendMedia_DoesNotFailOnEmptyMedia(t *testing.T) {
	strategy, bot := makeMediaStrategySUT()
	media := []*core.Media{}

	strategy.SendMedia(media, bot)

	assert.Equal(t, []string{}, bot.SentMedias)
}

func Test_SendMedia_SendsASingleMediaTroughABot(t *testing.T) {
	strategy, bot := makeMediaStrategySUT()
	media := []*core.Media{{ResourceURL: "https://a-url.com"}}

	strategy.SendMedia(media, bot)

	assert.Equal(t, []string{"https://a-url.com"}, bot.SentMedias)
}

func Test_SendMedia_SendsAGroupMediaTroughABot(t *testing.T) {
	strategy, bot := makeMediaStrategySUT()
	media := []*core.Media{{ResourceURL: "https://a-url.com"}, {ResourceURL: "https://another-url.com"}}

	strategy.SendMedia(media, bot)

	assert.Equal(t, []string{"https://a-url.com", "https://another-url.com"}, bot.SentMedias)
}

// Helpers
func makeMediaStrategySUT() (core.ISendMediaStrategy, *test_helpers.FakeBot) {
	strategy := helpers.CreateSendMediaStrategy()
	bot := test_helpers.CreateBot()
	return strategy, bot
}
