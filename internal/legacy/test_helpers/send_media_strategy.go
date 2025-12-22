package test_helpers

import (
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateSendMediaStrategy() *FakeSendMediaStrategy {
	return &FakeSendMediaStrategy{[]string{}, nil}
}

type FakeSendMediaStrategy struct {
	SentMedia []string
	Err       error
}

// SendMedia is a core.ISendMediaStrategy interface implementation
func (fsms *FakeSendMediaStrategy) SendMedia(media []*core.Media, bot core.IBot) error {
	for _, m := range media {
		fsms.SentMedia = append(fsms.SentMedia, m.URL)
	}
	return fsms.Err
}
