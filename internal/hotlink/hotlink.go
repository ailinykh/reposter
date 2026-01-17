package hotlink

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func (h *Handler) handleHotlink(url string, m *telegram.Message, bot *telegram.Bot) error {
	res, err := http.DefaultClient.Head(url)
	if err != nil {
		h.l.Error("failed to perform HEAD request", "url", url, "error", err)
		return nil
	}

	contentType := res.Header.Get("Content-Type")
	if len(contentType) == 0 {
		h.l.Error("content type not found", "url", url)
		return nil
	}

	h.l.Info("got contentType", "contentType", contentType)

	if strings.HasPrefix(contentType, "video") {
		_, err = bot.SendVideo(m.Chat.ID, url, fmt.Sprintf(`<a href="%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, url, path.Base(url), m.From.DisplayName()))
		return err
	}

	if strings.HasPrefix(contentType, "image") {
		_, err = bot.SendPhoto(telegram.SendPhotoParams{
			ChatID:    m.Chat.ID,
			Photo:     url,
			Caption:   fmt.Sprintf(`<a href="%s">🖼</a> <b>%s</b> <i>(by %s)</i>`, url, path.Base(url), m.From.DisplayName()),
			ParseMode: telegram.ParseModeMarkdown,
		})
		return err
	}

	h.l.Info("unsupported content-type", "content_type", contentType)
	return nil
}
