package usecases

import (
	"os"

	"github.com/ailinykh/reposter/v2/internal/api/image_uploader"
	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateImageFlow is a basic ImageFlow factory
func CreateImageFlow(l core.Logger, imageUploader image_uploader.Uploader, imageDownloader legacy.IImageDownloader) *ImageFlow {
	return &ImageFlow{l, imageUploader, imageDownloader}
}

// ImageFlow represents convert image to hotlink logic
type ImageFlow struct {
	l               core.Logger
	imageUploader   image_uploader.Uploader
	imageDownloader legacy.IImageDownloader
}

// HandleImage is a core.IImageHandler protocol implementation
func (flow *ImageFlow) HandleImage(image *legacy.Image, message *legacy.Message, bot legacy.IBot) error {
	if !message.IsPrivate {
		return nil
	}

	file, err := flow.imageDownloader.Download(image)
	if err != nil {
		return err
	}
	//TODO: memory management
	defer func() {
		flow.l.Info("removing", "file", file.Name())
		err = os.Remove(file.Name())
		if err != nil {
			flow.l.Error(err)
		}
	}()

	url, err := flow.imageUploader.Upload(file)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	flow.l.Info("image uploaded", "url", url)
	_, err = bot.SendText(url.String())
	return err
}
