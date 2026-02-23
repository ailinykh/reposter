package hotlink

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

func NewTaskHandler(
	chatID int64,
	l *slog.Logger,
	bot *telegram.Bot,
	cache Repo,
	yd *ytdlp.YtDlp,
) *TaskHandler {
	return &TaskHandler{
		bot:    bot,
		cache:  cache,
		chatID: chatID,
		l:      l,
		yd:     yd,
	}
}

type TaskHandler struct {
	bot    *telegram.Bot
	cache  Repo
	chatID int64
	l      *slog.Logger
	yd     *ytdlp.YtDlp
}

func (h *TaskHandler) Process(ctx context.Context, task MediaTask) MediaTaskResult {
	messages, result, err := h.handle(ctx, task)
	if err != nil {
		return MediaTaskResult{
			TaskID: task.ID,
			Ok:     false,
			Error:  NewMediaTaskError(err),
		}
	}
	return MediaTaskResult{
		TaskID:      task.ID,
		Ok:          true,
		Title:       result.Title,
		Description: result.Description,
		Messages:    messages,
	}
}

func (h *TaskHandler) handle(ctx context.Context, task MediaTask) ([]*telegram.Message, *ytdlp.Response, error) {
	r, err := h.GetFormat(ctx, task.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get format: %w", err)
	}

	if r.MediaType == "livestream" {
		return nil, nil, &MediaTaskError{
			NotifyUser: true,
			Message:    "😬 live stream is not supported yet",
		}
	}

	key := fmt.Sprintf("%s.id.%s.bot.%s.messages", strings.ToLower(r.Extractor), r.ID, h.bot.Username)
	cache, err := h.cache.Get(ctx, key)
	if err == nil {
		var messages []*telegram.Message
		if err = json.Unmarshal(cache.Value, &messages); err != nil {
			return nil, nil, err
		}
		h.l.Info("got messages from cache", "key", key, "count", len(messages))
		return messages, r, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	const maxSize int64 = 50_000_000 // Telegram multipart/form-data limit
	if r.Filesize > maxSize || r.Duration > 360 {
		h.l.Warn("video too long", "id", r.ID, "extractor", r.Extractor, "size", r.Filesize, "duration", r.Duration)
		return nil, nil, &VideoTooLongError{
			Duration: time.Duration(r.Duration),
			Title:    r.Title,
		}
	}

	// No messages in cache found, try to upload it to channel
	video, err := h.yd.DownloadFormat(r.FormatID, r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer video.Dispose()

	t := telegram.InputFileLocal{
		Name:   video.Thumb.Name,
		Reader: video.Thumb.File,
	}

	if r.MediaType == "short" {
		cropped, err := CroppedThumb(r, video)
		if err == nil {
			t = telegram.InputFileLocal{
				Name:   video.Thumb.Name,
				Reader: cropped.File,
			}
			defer cropped.Dispose()
		} else {
			h.l.Error("failed to crop thumbnail", "error", err)
		}
	}

	m, err := h.bot.SendVideo(ctx, &telegram.SendVideoParams{
		ChatID: h.chatID,
		Video: telegram.InputFileLocal{
			Name:   video.Name,
			Reader: video.File,
		},
		Duration:          int(r.Duration),
		Width:             r.Width,
		Height:            r.Height,
		Thumbnail:         t,
		Caption:           telegram.NormalizeCaption(fmt.Sprintf("<a href=\"%s\"><b>%s</b></a>\n%s", task.URL, r.Title, r.Description)),
		ParseMode:         telegram.ParseModeHTML,
		SupportsStreaming: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send video: %w", err)
	}

	h.l.Info("video sent successfully", "extractor", r.Extractor, "size", r.Filesize, "duration", r.Duration)

	if m.Video == nil {
		return nil, nil, fmt.Errorf("no video in outgoing message found")
	}

	messages := []*telegram.Message{m}
	data, err := json.Marshal(messages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode videos: %w", err)
	}

	_, err = h.cache.Set(ctx, repository.SetParams{
		Key:   key,
		Value: data,
	})
	return messages, r, err
}

func (h *TaskHandler) GetFormat(ctx context.Context, url string) (*ytdlp.Response, error) {
	cache, err := h.cache.Get(ctx, url)

	if err == nil {
		var r *ytdlp.Response
		if err = json.Unmarshal(cache.Value, &r); err != nil {
			return nil, err
		}
		h.l.Info("got yt-dlp response from cache", "key", url)
		return r, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	data, err := h.yd.GetFormatRaw(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get format: %w", err)
	}

	var r *ytdlp.Response
	if err = json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("failed to encode yt-dlp response: %w", err)
	}

	_, err = h.cache.Set(ctx, repository.SetParams{
		Key:   url,
		Value: data,
	})

	return r, nil
}
