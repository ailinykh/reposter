package hotlink

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/xcom"
)

func (h *Handler) handleXcom(ctx context.Context, urlString string, m *telegram.Message, bot *telegram.Bot) error {
	r := regexp.MustCompile(`https://(?i:twitter|x)\.com\S+/status/(\d+)`)
	match := r.FindStringSubmatch(urlString)

	if len(match) < 2 {
		h.l.Warn("can't find tweet id", "url", urlString)
		return nil
	}

	tweet, err := h.x.Get(ctx, match[1])
	if err != nil {
		h.l.Error("failed to get tweet", "error", err)
		return err
	}

	params := &telegram.SendMediaGroupParams{
		ChatID: m.Chat.ID,
		Media:  []telegram.InputMedia{},
	}

	var caption string
	if tweet.NoteTweet != nil {
		caption = fmt.Sprintf("<a href='%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", urlString, tweet.Core.UserResults.Result.Core.Name, m.From.DisplayName(), tweet.NoteTweet.NoteTweetResults.Result.Text)
	} else {
		re := regexp.MustCompile(`\s?http\S+$`)
		text := re.ReplaceAllString(tweet.Legacy.FullText, "")
		caption = fmt.Sprintf("<a href='%s'>🐦</a> <b>%s</b> <i>(by %s)</i>\n%s", urlString, tweet.Core.UserResults.Result.Core.Name, m.From.DisplayName(), text)
	}

	for i, m := range tweet.Legacy.Entities.Media {
		switch m.Type {
		case "photo":
			photo := &telegram.InputMediaPhoto{
				Type:                  "photo",
				Media:                 m.MediaUrlHttps,
				ParseMode:             telegram.ParseModeHTML,
				ShowCaptionAboveMedia: true,
			}
			if i == len(tweet.Legacy.Entities.Media)-1 {
				photo.Caption = caption
			}
			params.Media = append(params.Media, photo)
		case "video", "animated_gif":
			video := &telegram.InputMediaVideo{
				Type:                  "video",
				Media:                 m.VideoInfo.Best().URL,
				Duration:              int(m.VideoInfo.Duration / 1000),
				Thumbnail:             m.MediaUrlHttps,
				ParseMode:             telegram.ParseModeHTML,
				ShowCaptionAboveMedia: true,
			}
			if i == len(tweet.Legacy.Entities.Media)-1 {
				video.Caption = caption
			}
			params.Media = append(params.Media, video)
		default:
			h.l.Error("unexpected media type", "type", m.Type, "tweet", tweet.RestID)
		}
	}

	if len(params.Media) == 0 {
		h.l.Warn("no media found", "url", urlString)
		return &xcom.Error{TypeName: "No media found.", Reason: "Only tweets with media supported at the moment"}
	}

	_, err = bot.SendMediaGroup(ctx, params)
	return nil
}
