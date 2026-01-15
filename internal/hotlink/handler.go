package hotlink

import (
	"errors"
	"log/slog"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

var ErrURLNotSupported = errors.New("url not supported")

func New(l *slog.Logger, yd *ytdlp.YtDlp) *Handler {
	return &Handler{
		l:  l,
		yd: yd,
	}
}

type Handler struct {
	l  *slog.Logger
	yd *ytdlp.YtDlp
}

func (h *Handler) Handle(u *telegram.Update, bot *telegram.Bot) error {
	if u.Message == nil || len(u.Message.URLs()) == 0 {
		return nil
	}

	for _, urlString := range u.Message.URLs() {
		if err := h.handleSocial(urlString, u.Message, bot); err != nil {
			if errors.Is(err, ErrURLNotSupported) {
				return h.handleHotlink(urlString, u.Message, bot)
			}
			h.l.Error("failed to process url", "error", err, "url", urlString)
			// _, _ = bot.SendMessage(u.Message.Chat.ID, "😬 "+err.Error(), map[string]any{
			// 	"reply_parameters": map[string]any{
			// 		"message_id": u.Message.ID,
			// 		"quote":      urlString,
			// 		"link_preview_options": map[string]any{
			// 			"is_disabled": true,
			// 		},
			// 	},
			// })
			return err
		}
	}

	return nil
}
