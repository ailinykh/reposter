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

	ch := make(chan MediaTaskResult)
	go func() {
		result := <-ch
		h.handleMediaTaskResult(ctx, result, urlString, m, bot)
	}()

	task := MediaTask{
		ID:   uuid.NewString(),
		Date: time.Now(),
		URL:  expanded,
	}
	h.l.Info("new task created", "task_id", task.ID, "url", task.URL)

	return h.q.Consume(ctx, task, ch)
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
		if err := h.CanNotifyUser(result.Error); err != nil {
			_, _ = bot.SendMessage(ctx, &telegram.SendMessageParams{
				ChatID:    m.Chat.ID,
				Text:      err.Error(),
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
	if len(caption) > 1024 {
		caption = caption[:1024]
	}
	caption = strings.ToValidUTF8(caption, "")

	videos := VideoFromMessages(result.Messages)

	switch len(videos) {
	case 0:
		h.l.Error("expect at least one video", "task_id", result.TaskID, "messages", result.Messages)
	default:
		if _, err := bot.SendMediaGroup(ctx, &telegram.SendMediaGroupParams{
			ChatID: m.Chat.ID,
			Media:  MediaFromVideos(videos, caption),
		}); err != nil {
			h.l.Error("failed to send media group", "task_id", result.TaskID, "error", err)
		}
	}

	return nil
}
