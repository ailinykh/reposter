package usecases

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateLinkFlow is a basic LinkFlow factory
func CreateLinkFlow(l core.Logger, httpClient legacy.IHttpClient, mediaFactory legacy.IMediaFactory, sendMediaStrategy legacy.ISendMediaStrategy) *LinkFlow {
	return &LinkFlow{l, httpClient, mediaFactory, sendMediaStrategy}
}

// LinkFlow converts hotlink to video/photo attachment
type LinkFlow struct {
	l                 core.Logger
	httpClient        legacy.IHttpClient
	mediaFactory      legacy.IMediaFactory
	sendMediaStrategy legacy.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *LinkFlow) HandleText(message *legacy.Message, bot legacy.IBot) error {
	r := regexp.MustCompile(`^http(\S+)$`)
	if r.MatchString(message.Text) {
		return flow.handleURL(message.Text, message, bot)
	}
	return fmt.Errorf("not implemented")
}

func (flow *LinkFlow) handleURL(url legacy.URL, message *legacy.Message, bot legacy.IBot) error {
	contentType, err := flow.httpClient.GetContentType(url)
	if err != nil {
		return fmt.Errorf("failed to get content from %s: %v", url, err)
	}

	if !strings.HasPrefix(contentType, "video") && !strings.HasPrefix(contentType, "image") {
		return fmt.Errorf("not implemented")
	}

	media, err := flow.mediaFactory.CreateMedia(url)
	if err != nil {
		return fmt.Errorf("failed to create media from %s: %v", url, err)
	}

	for _, m := range media {
		switch m.Type {
		case legacy.TPhoto:
			m.Caption = fmt.Sprintf(`<a href="%s">🖼</a> <b>%s</b> <i>(by %s)</i>`, m.URL, path.Base(m.URL), message.Sender.DisplayName())
		case legacy.TVideo:
			m.Caption = fmt.Sprintf(`<a href="%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, m.URL, path.Base(m.URL), message.Sender.DisplayName())
		case legacy.TText:
			flow.l.Warn("Unexpected content type", "type", m)
		}
	}

	err = flow.sendMediaStrategy.SendMedia(media, bot)
	if err != nil {
		return fmt.Errorf("failed to send media: %v", err)
	}

	return nil
}
