package usecases

import (
	"strings"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateIDoNotCare() *IDoNotCare {
	return &IDoNotCare{}
}

type IDoNotCare struct{}

// HandleText is a core.ITextHandler protocol implementation
func (IDoNotCare) HandleText(message *core.Message, bot core.IBot) error {
	if strings.Contains(strings.ToLower(message.Text), "мне всё равно") {
		_, err := bot.SendVideo(&core.Video{ID: "BAACAgIAAxkBAAEDfu1kFdKdAAHM4iO92LOC3muxi2yyvosAAgQoAAIZvLFIVaKgRXqfmVgvBA"}, "")
		if err != nil {
			media := &core.Media{
				ResourceURL: "https://telegra.ph/file/182c624365bea4df6842a.mp4",
				Type:        core.TVideo,
			}
			_, err = bot.SendMedia(media)
		}
		return err
	}
	if strings.Contains(strings.ToLower(message.Text), "привет, андрей") {
		_, err := bot.SendVideo(&core.Video{ID: "BAACAgIAAxkBAAIziWEeZBqlM1_1n2AVaxedGFn3vS-sAAKgDwACSl7xSImLuE-s8DMbIAQ"}, "")
		return err
	}
	return nil
}
