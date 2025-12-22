package helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateSendVideoStrategySplitDecorator(l core.Logger, decoratee legacy.ISendVideoStrategy, splitter legacy.IVideoSplitter) legacy.ISendVideoStrategy {
	return &SendVideoStrategySplitDecorator{l, decoratee, splitter}
}

type SendVideoStrategySplitDecorator struct {
	l         core.Logger
	decoratee legacy.ISendVideoStrategy
	splitter  legacy.IVideoSplitter
}

// SendMedia is a core.ISendVideoStrategy interface implementation
func (strategy *SendVideoStrategySplitDecorator) SendVideo(video *legacy.Video, caption string, bot legacy.IBot) error {
	err := strategy.decoratee.SendVideo(video, caption, bot)
	if err != nil && strings.HasSuffix(err.Error(), "telegram: Request Entity Too Large (400)") {
		strategy.l.Info("fallback to splitting")
		// https://lists.ffmpeg.org/pipermail/ffmpeg-user/2014-January/019556.html
		files, err := strategy.splitter.Split(video, 49)
		if err != nil {
			return err
		}

		for _, file := range files {
			defer file.Dispose()
		}

		r := regexp.MustCompile(`<b>(.*)</b>`)
		match := r.FindStringSubmatch(caption)

		for i, file := range files {
			c := caption
			if len(match) > 0 {
				c = r.ReplaceAllString(caption, fmt.Sprintf("<b>[%d/%d] %s</b>", i+1, len(files), match[1]))
			}
			_, err := bot.SendVideo(file, c)
			if err != nil {
				return err
			}
		}

		strategy.l.Info("all parts successfully sent")
		return nil
	}
	return err
}
