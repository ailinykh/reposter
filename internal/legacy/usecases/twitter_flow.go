package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

type ITweetHandler interface {
	Process(string, *legacy.Message, legacy.IBot) error
}

// CreateTwitterFlow is a basic TwitterFlow factory
func CreateTwitterFlow(l core.Logger, mediaFactory legacy.IMediaFactory, sendMediaStrategy legacy.ISendMediaStrategy) *TwitterFlow {
	return &TwitterFlow{l, mediaFactory, sendMediaStrategy}
}

// TwitterFlow represents tweet processing logic
type TwitterFlow struct {
	l                 core.Logger
	mediaFactory      legacy.IMediaFactory
	sendMediaStrategy legacy.ISendMediaStrategy
}

// Process is a ITweetHandler protocol implementation
func (flow *TwitterFlow) Process(tweetID string, message *legacy.Message, bot legacy.IBot) error {
	flow.l.Info("processing tweet", "tweet_id", tweetID)
	media, err := flow.mediaFactory.CreateMedia(tweetID)
	if err != nil {
		return fmt.Errorf("failed to create media: %v", err)
	}

	for _, m := range media {
		re := regexp.MustCompile(`\s?http\S+$`)
		text := re.ReplaceAllString(m.Description, "")
		m.Caption = fmt.Sprintf("<a href='%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", m.URL, m.Title, message.Sender.DisplayName(), text)
	}

	err = flow.sendMediaStrategy.SendMedia(media, bot)
	if err != nil {
		return fmt.Errorf("failed to send media: %v", err)
	}

	return nil
}
