package helpers

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
)

func CreateConvertMediaStrategy(
	l core.Logger,
	sms legacy.ISendMediaStrategy,
	fd legacy.IFileDownloader,
	vf legacy.IVideoFactory,
	vc legacy.IVideoConverter,
) *ConvertMediaStrategy {
	return &ConvertMediaStrategy{l, sms, fd, vf, vc}
}

type ConvertMediaStrategy struct {
	l   core.Logger
	sms legacy.ISendMediaStrategy
	fd  legacy.IFileDownloader
	vf  legacy.IVideoFactory
	vc  legacy.IVideoConverter
}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (cms *ConvertMediaStrategy) SendMedia(media []*legacy.Media, bot legacy.IBot) error {
	for _, m := range media {
		if cms.needToConvert(m) {
			cms.l.Info("fallback ot convert", "codec", m.Codec)
			return cms.fallbackToConverting(m, bot)
		}
	}
	return cms.sms.SendMedia(media, bot)
}

func (cms *ConvertMediaStrategy) needToConvert(media *legacy.Media) bool {
	if media.Type != legacy.TVideo {
		return false
	}

	for _, codec := range []string{"mp4", "h264"} {
		if media.Codec == codec {
			return false
		}
	}
	return true
}

func (cms *ConvertMediaStrategy) fallbackToConverting(media *legacy.Media, bot legacy.IBot) error {
	cms.l.Info("send by converting")
	file, err := cms.downloadMedia(media)
	if err != nil {
		return err
	}
	defer file.Dispose()

	vf, err := cms.vf.CreateVideo(file.Path)
	if err != nil {
		return fmt.Errorf("failed to create video file at %s: %v", file.Path, err)
	}
	defer vf.Dispose()

	vfc, err := cms.vc.Convert(vf, 0)
	if err != nil {
		return fmt.Errorf("failed to convert video file: %v", err)
	}
	defer vfc.Dispose()

	_, err = bot.SendVideo(vfc, media.Caption)
	return err
}

func (cms *ConvertMediaStrategy) downloadMedia(media *legacy.Media) (*legacy.File, error) {
	//TODO: duplicated code
	filename := path.Base(media.ResourceURL)
	if strings.Contains(filename, "?") {
		parts := strings.Split(media.ResourceURL, "?")
		filename = path.Base(parts[0])
	}
	mediaPath := path.Join(os.TempDir(), filename)
	file, err := cms.fd.Download(media.ResourceURL, mediaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %v", err)
	}

	cms.l.Info("file downloaded", "file_name", file.Name, "file_size", fmt.Sprintf("%0.2fMB", float64(file.Size)/1024/1024))

	return file, nil
}
