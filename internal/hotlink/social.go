package hotlink

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/google/uuid"
)

func (h *Handler) handleSocial(ctx context.Context, urlString string, m *telegram.Message, bot *telegram.Bot) error {
	expanded, err := ExpandURL(urlString)
	if err != nil {
		return ErrURLNotSupported
	}

	if strings.Contains(expanded, "x.com/status") {
		return h.handleXcom(ctx, urlString, m, bot)
	}

	task := MediaTask{
		ID:   uuid.NewString(),
		Date: time.Now(),
		URL:  expanded,
	}
	h.l.Info("new task created", "task_id", task.ID, "url", task.URL)

	ch, err := h.q.Consume(ctx, task)
	if err != nil {
		return err
	}

	go func() {
		result := <-ch
		h.handleMediaTaskResult(ctx, result, urlString, m, bot)
	}()
	return nil
}

func (h *Handler) handleMediaTaskResult(
	ctx context.Context,
	result MediaTaskResult,
	urlString string,
	m *telegram.Message,
	bot *telegram.Bot,
) error {
	if !result.Ok {
		h.l.Error("failed to perform task", "task_id", result.TaskID, "error", result.Error)
		if !m.Chat.Private() {
			return nil // silent in group chat
		}

		if result.Error.NotifyUser {
			_, _ = bot.SendMessage(ctx, &telegram.SendMessageParams{
				ChatID:    m.Chat.ID,
				Text:      result.Error.Message,
				ParseMode: telegram.ParseModeHTML,
				ReplyParameters: &telegram.ReplyParameters{
					MessageID: m.ID,
					Quote:     urlString,
				},
			})
		}
		return nil
	}

	caption := fmt.Sprintf("<a href=\"%s\">🎞</a> <b>%s</b> <i>(by %s)</i>\n\n%s", urlString, result.Title, m.From.DisplayName(), result.Description)
	videos := VideoFromMessages(result.Messages)

	switch len(videos) {
	case 0:
		h.l.Error("expect at least one video", "task_id", result.TaskID, "messages", result.Messages)
	default:
		if _, err := bot.SendMediaGroup(ctx, &telegram.SendMediaGroupParams{
			ChatID: m.Chat.ID,
			Media:  MediaFromVideos(videos, telegram.NormalizeCaption(caption)),
		}); err != nil {
			h.l.Error("failed to send media group", "task_id", result.TaskID, "error", err)
		}
	}

	return nil
}
