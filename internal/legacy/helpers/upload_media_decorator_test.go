package helpers_test

import (
	"fmt"
	"testing"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/stretchr/testify/assert"
)

func Test_UploadMedia_DoesNotFailOnEmptyMedia(t *testing.T) {
	strategy, _, bot := makeUploadMediaDecoratorSUT()
	media := []*core.Media{}

	strategy.SendMedia(media, bot)

	assert.Equal(t, []string{}, bot.SentMedias)
}

func Test_UploadMedia_DoesNotFallbackOnGenericError(t *testing.T) {
	strategy, proxy, bot := makeUploadMediaDecoratorSUT()
	media := []*core.Media{}
	proxy.Err = fmt.Errorf("an error")

	err := strategy.SendMedia(media, bot)

	assert.Equal(t, proxy.Err, err)
}

func Test_UploadMedia_FallbackOnSpecificError(t *testing.T) {
	strategy, proxy, bot := makeUploadMediaDecoratorSUT()
	media := []*core.Media{{ResourceURL: "https://a-url.com"}}
	proxy.Err = fmt.Errorf("failed to get HTTP URL content")

	err := strategy.SendMedia(media, bot)

	assert.Equal(t, nil, err)
}

// Helpers
func makeUploadMediaDecoratorSUT() (core.ISendMediaStrategy, *test_helpers.FakeSendMediaStrategy, *test_helpers.FakeBot) {
	logger := test_helpers.CreateLogger()
	send_media_strategy := test_helpers.CreateSendMediaStrategy()
	file_downloader := test_helpers.CreateFileDownloader()
	video_factory := test_helpers.CreateVideoFactory()
	send_video := test_helpers.CreateSendVideoStrategy()
	strategy := helpers.CreateUploadMediaDecorator(logger, send_media_strategy, file_downloader, video_factory, send_video)
	bot := test_helpers.CreateBot()
	return strategy, send_media_strategy, bot
}
