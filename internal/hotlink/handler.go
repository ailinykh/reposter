package hotlink

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/xcom"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

type Repo interface {
	Set(ctx context.Context, arg repository.SetParams) (repository.Cache, error)
	Get(ctx context.Context, key string) (repository.Cache, error)
}

func New(l *slog.Logger, cache Repo, x *xcom.XComAPI, yd *ytdlp.YtDlp) *Handler {
	return &Handler{
		l:     l,
		cache: cache,
		x:     x,
		yd:    yd,
	}
}

type Handler struct {
	l     *slog.Logger
	cache Repo
	x     *xcom.XComAPI
	yd    *ytdlp.YtDlp
}

func (h *Handler) Handle(ctx context.Context, u *telegram.Update, bot *telegram.Bot) error {
	// TODO: respect type="text_link" as well
	if u.Message == nil || len(u.Message.URLs()) == 0 {
		return nil
	}

	canNotifyUser := func(err error) error {
		var tooLong *VideoTooLongError
		if errors.As(err, &tooLong) {
			return fmt.Errorf("%s\n<b>⏳ video too long: %d sec</b>", tooLong.Title, tooLong.Duration)
		}

		var xErr *xcom.Error
		if errors.As(err, &xErr) {
			return fmt.Errorf("😬 %s", xErr.Error())
		}

		var ytErr *ytdlp.Error
		if errors.As(err, &ytErr) {
			return fmt.Errorf("😬 %s", ytErr.Error())
		}

		return nil
	}

	for _, urlString := range u.Message.URLs() {
		if err := h.handleSocial(ctx, urlString, u.Message, bot); err != nil {
			if errors.Is(err, ErrURLNotSupported) {
				return h.handleHotlink(ctx, urlString, u.Message, bot)
			}

			if e := canNotifyUser(err); e != nil {
				_, _ = bot.SendMessage(ctx, &telegram.SendMessageParams{
					ChatID:    u.Message.Chat.ID,
					Text:      e.Error(),
					ParseMode: telegram.ParseModeHTML,
					ReplyParameters: &telegram.ReplyParameters{
						MessageID: u.Message.ID,
						Quote:     urlString,
					},
				})
			} else {
				h.l.Error("failed to process url", "error", err, "url", urlString)
				return err
			}
		}
	}

	return nil
}
