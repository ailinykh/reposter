package helpers

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateUploadMediaDecorator(l core.Logger, decoratee legacy.ISendMediaStrategy, fileDownloader legacy.IFileDownloader, videoFactory legacy.IVideoFactory, sendVideo legacy.ISendVideoStrategy) legacy.ISendMediaStrategy {
	return &UploadMediaDecorator{l, decoratee, fileDownloader, videoFactory, sendVideo}
}

type UploadMediaDecorator struct {
	l              core.Logger
	decoratee      legacy.ISendMediaStrategy
	fileDownloader legacy.IFileDownloader
	videoFactory   legacy.IVideoFactory
	sendVideo      legacy.ISendVideoStrategy
}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (decorator *UploadMediaDecorator) SendMedia(media []*legacy.Media, bot legacy.IBot) error {
	err := decorator.decoratee.SendMedia(media, bot)
	if err != nil {
		fallbackErrors := []string{
			"failed to get HTTP URL content",
			"wrong file identifier/HTTP URL specified",
			"wrong type of the web page content",
		}
		for _, e := range fallbackErrors {
			if strings.Contains(err.Error(), e) {
				return decorator.fallbackToUploading(media[0], bot)
			}
		}
	}

	return err
}

func (decorator *UploadMediaDecorator) fallbackToUploading(media *legacy.Media, bot legacy.IBot) error {
	decorator.l.Info("send by uploading")
	file, err := decorator.downloadMedia(media)
	if err != nil {
		return err
	}
	defer file.Dispose()

	switch media.Type {
	case legacy.TText:
		decorator.l.Warn("unexpected media type", "type", media.Type)
	case legacy.TPhoto:
		image := &legacy.Image{File: *file}
		_, err = bot.SendImage(image, media.Caption)
		return err
	case legacy.TVideo:
		vf, err := decorator.videoFactory.CreateVideo(file.Path)
		if err != nil {
			return fmt.Errorf("can't create video file for %s, %v", file.Path, err)
		}
		return decorator.sendVideo.SendVideo(vf, media.Caption, bot)
	}
	return err
}

func (decorator *UploadMediaDecorator) downloadMedia(media *legacy.Media) (*legacy.File, error) {
	//TODO: duplicated code
	filename := path.Base(media.ResourceURL)
	if strings.Contains(filename, "?") {
		parts := strings.Split(media.ResourceURL, "?")
		filename = path.Base(parts[0])
	}

	if !strings.HasSuffix(filename, ".mp4") {
		filename = filename + ".mp4"
	}

	mediaPath := path.Join(os.TempDir(), filename)
	file, err := decorator.fileDownloader.Download(media.ResourceURL, mediaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %v", err)
	}

	decorator.l.Info("file downloaded", "file_name", file.Name, "file_size", fmt.Sprintf("%0.2fMB", float64(file.Size)/1024/1024))

	return file, nil
}
