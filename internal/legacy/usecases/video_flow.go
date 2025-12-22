package usecases

import (
	"fmt"
	"math"
	"os"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateVideoFlow is a basic VideoFlow factory
func CreateVideoFlow(l core.Logger, videoFactory legacy.IVideoFactory, converter legacy.IVideoConverter) *VideoFlow {
	return &VideoFlow{l, converter, videoFactory}
}

// VideoFlow represents convert file to video logic
type VideoFlow struct {
	l            core.Logger
	converter    legacy.IVideoConverter
	videoFactory legacy.IVideoFactory
}

// HandleDocument is a core.IDocumentHandler protocol implementation
func (flow *VideoFlow) HandleDocument(document *legacy.Document, message *legacy.Message, bot legacy.IBot) error {
	vf, err := flow.videoFactory.CreateVideo(document.File.Path)
	if err != nil {
		return fmt.Errorf("failed to create video: %v", err)
	}
	defer vf.Dispose()

	expectedBitrate := int(math.Min(float64(vf.Bitrate), 568320))

	if expectedBitrate != vf.Bitrate {
		flow.l.Info("converting because of bitrate", "file_name", vf.Name)
		cvf, err := flow.converter.Convert(vf, expectedBitrate)
		if err != nil {
			return fmt.Errorf("failed to convert video: %v", err)
		}
		defer cvf.Dispose()
		fi1, _ := os.Stat(vf.Path)
		fi2, _ := os.Stat(cvf.Path)
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>src: %.2f MB (%d kb/s) %s\ndst: %.2f MB (%d kb/s) %s</i>", vf.Name, message.Sender.DisplayName(), float32(fi1.Size())/1048576, vf.Bitrate/1024, vf.Codec, float32(fi2.Size())/1048576, cvf.Bitrate/1024, cvf.Codec)
		_, err = bot.SendVideo(cvf, caption)
		if err != nil {
			return fmt.Errorf("failed to send video: %v", err)
		}
		return bot.Delete(message)
	}

	if vf.Codec != "h264" {
		flow.l.Info("converting because of codec", "file_name", vf.Name, "codec", vf.Codec)
		cvf, err := flow.converter.Convert(vf, 0)
		if err != nil {
			return fmt.Errorf("failed to convert video: %v", err)
		}
		defer cvf.Dispose()
		fi1, _ := os.Stat(vf.Path)
		fi2, _ := os.Stat(cvf.Path)
		caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>\n<i>src: %.2f MB (%d kb/s) %s\ndst: %.2f MB (%d kb/s) %s</i>", vf.Name, message.Sender.DisplayName(), float32(fi1.Size())/1048576, vf.Bitrate/1024, vf.Codec, float32(fi2.Size())/1048576, cvf.Bitrate/1024, cvf.Codec)
		_, err = bot.SendVideo(cvf, caption)
		if err != nil {
			return fmt.Errorf("failed to send video: %v", err)
		}
		return bot.Delete(message)
	}

	flow.l.Info("no need to convert", "file_name", vf.Name)
	caption := fmt.Sprintf("<b>%s</b> <i>(by %s)</i>", vf.Name, message.Sender.DisplayName())
	_, err = bot.SendVideo(vf, caption)
	if err != nil {
		return err
	}
	return bot.Delete(message)
}
