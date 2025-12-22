package test_helpers

import (
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateSendVideoStrategy() *FakeSendVideoStrategy {
	return &FakeSendVideoStrategy{[]string{}, nil}
}

type FakeSendVideoStrategy struct {
	SentVideos []string
	Err        error
}

// SendVideo is a core.ISendVideoStrategy interface implementation
func (fsms *FakeSendVideoStrategy) SendVideo(video *core.Video, caption string, bot core.IBot) error {
	fsms.SentVideos = append(fsms.SentVideos, video.Name)
	return fsms.Err
}
