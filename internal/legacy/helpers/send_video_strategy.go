package helpers

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateSendVideoStrategy() core.ISendVideoStrategy {
	return &SendVideoStrategy{}
}

type SendVideoStrategy struct{}

// SendMedia is a core.ISendVideoStrategy interface implementation
func (strategy *SendVideoStrategy) SendVideo(video *core.Video, caption string, bot core.IBot) error {
	_, err := bot.SendVideo(video, caption)

	if err != nil {
		return fmt.Errorf("failed to send video: %v", err)
	}

	return nil
}
