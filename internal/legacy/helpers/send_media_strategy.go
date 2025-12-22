package helpers

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateSendMediaStrategy() *SendMediaStrategy {
	return &SendMediaStrategy{}
}

type SendMediaStrategy struct{}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (sms *SendMediaStrategy) SendMedia(media []*core.Media, bot core.IBot) error {
	switch len(media) {
	case 0:
		return fmt.Errorf("attempt to send an empty media")
	case 1:
		_, err := bot.SendMedia(media[0])
		return err
	default:
		_, err := bot.SendMediaAlbum(media)
		return err
	}
}
