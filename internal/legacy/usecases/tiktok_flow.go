package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateTikTokFlow(httpClient core.IHttpClient, mediaFactory core.IMediaFactory, sendMediaStrategy core.ISendMediaStrategy) *TikTokFlow {
	return &TikTokFlow{httpClient, mediaFactory, sendMediaStrategy}
}

type TikTokFlow struct {
	httpClient        core.IHttpClient
	mediaFactory      core.IMediaFactory
	sendMediaStrategy core.ISendMediaStrategy
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *TikTokFlow) HandleText(message *core.Message, bot core.IBot) error {
	r := regexp.MustCompile(`https?://\w+\.tiktok.com/\S+`)
	links := r.FindAllString(message.Text, -1)
	for _, l := range links {
		err := flow.handleURL(l, message, bot)
		if err != nil {
			return fmt.Errorf("failed to process %s: %v", l, err)
		}
	}

	if len(links) > 0 {
		return bot.Delete(message)
	}
	return nil
}

func (flow *TikTokFlow) handleURL(url string, message *core.Message, bot core.IBot) error {
	media, err := flow.mediaFactory.CreateMedia(url)
	if err != nil {
		if err.Error() == "Video currently unavailable" {
			_, err := bot.SendText(url + "\nV" + err.Error())
			return err
		}
		return err
	}

	for _, m := range media {
		m.URL = url
		m.Caption = fmt.Sprintf("<a href='%s'>🎵</a> <b>%s</b> (by %s)\n%s", url, m.Title, message.Sender.DisplayName(), m.Description)
	}

	return flow.sendMediaStrategy.SendMedia(media, bot)
}
