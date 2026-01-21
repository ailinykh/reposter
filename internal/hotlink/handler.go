package hotlink

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

type Repo interface {
	Set(ctx context.Context, arg repository.SetParams) (repository.Cache, error)
	Get(ctx context.Context, key string) (repository.Cache, error)
}

func New(l *slog.Logger, cache Repo, yd *ytdlp.YtDlp) *Handler {
	return &Handler{
		l:     l,
		cache: cache,
		yd:    yd,
	}
}

type Handler struct {
	l     *slog.Logger
	cache Repo
	yd    *ytdlp.YtDlp
}

func (h *Handler) Handle(u *telegram.Update, bot *telegram.Bot) error {
	// TODO: respect type="text_link" as well
	if u.Message == nil || len(u.Message.URLs()) == 0 {
		return nil
	}

	for _, urlString := range u.Message.URLs() {
		if err := h.handleSocial(urlString, u.Message, bot); err != nil {
			if errors.Is(err, ErrURLNotSupported) {
				return h.handleHotlink(urlString, u.Message, bot)
			}

			var tooLong *VideoTooLongError
			if errors.As(err, &tooLong) {
				_, _ = bot.SendMessage(telegram.SendMessageParams{
					ChatID:    u.Message.Chat.ID,
					Text:      fmt.Sprintf("%s\n<b>⏳ video too long: %d sec</b>", tooLong.Title, tooLong.Duration),
					ParseMode: telegram.ParseModeHTML,
					ReplyParameters: &telegram.ReplyParameters{
						MessageID: u.Message.ID,
						Quote:     urlString,
					},
				})
				continue
			}

			h.l.Error("failed to process url", "error", err, "url", urlString)

			var ytErr *ytdlp.Error
			if !errors.As(err, &ytErr) {
				return err
			}
			// Youtube related error occured, we can notify user
			_, _ = bot.SendMessage(telegram.SendMessageParams{
				ChatID: u.Message.Chat.ID,
				Text:   "😬 " + ytErr.Error(),
				LinkPreviewOptions: &telegram.LinkPreviewOptions{
					IsDisabled: true,
				},
				ReplyParameters: &telegram.ReplyParameters{
					MessageID: u.Message.ID,
					Quote:     urlString,
				},
			})
		}
	}

	return nil
}
