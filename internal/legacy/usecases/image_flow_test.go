package usecases_test

import (
	"os"
	"testing"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_HandleImage_DownloadsAndUploadsImage(t *testing.T) {
	logger := test_helpers.CreateLogger()
	image_uploader := test_helpers.CreateImageUploader()
	image_downloader := test_helpers.CreateImageDownloader()
	image_flow := usecases.CreateImageFlow(logger, image_uploader, image_downloader)

	url := "http://an-image-url.com"
	path, _ := os.CreateTemp(t.TempDir(), t.Name())
	image := &core.Image{FileURL: url, File: core.File{Path: path.Name()}}

	message := &core.Message{IsPrivate: true}
	bot := test_helpers.CreateBot()

	image_flow.HandleImage(image, message, bot)

	assert.Equal(t, []string{url}, image_downloader.Downloaded)
	assert.Equal(t, []string{path.Name()}, image_uploader.Uploaded)

	_, err := os.Stat(path.Name())
	assert.True(t, os.IsNotExist(err))

}
