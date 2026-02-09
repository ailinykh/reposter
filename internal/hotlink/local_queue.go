package hotlink

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

func NewLocalQueue(
	chatID int64,
	l *slog.Logger,
	bot *telegram.Bot,
	cache Repo,
	yd *ytdlp.YtDlp,
) Queue {
	ch := make(chan LocalQueueTask)
	queue := LocalQueue{
		chatID: chatID,
		bot:    bot,
		cache:  cache,
		l:      l,
		yd:     yd,
		ch:     ch,
	}

	go func() {
		for task := range ch {
			queue.process(task.ctx, task.task, task.callback)
		}
	}()

	return &queue
}

type LocalQueue struct {
	bot    *telegram.Bot
	cache  Repo
	chatID int64
	l      *slog.Logger
	yd     *ytdlp.YtDlp
	ch     chan LocalQueueTask
}

type LocalQueueTask struct {
	task     MediaTask
	callback chan MediaTaskResult
	ctx      context.Context
}

func (q *LocalQueue) Consume(ctx context.Context, task MediaTask, cb chan MediaTaskResult) error {
	go func() {
		q.ch <- LocalQueueTask{
			task:     task,
			callback: cb,
			ctx:      ctx,
		}
	}()
	return nil
}

func (q *LocalQueue) process(ctx context.Context, task MediaTask, cb chan MediaTaskResult) {
	q.l.Info("processing task", "task_id", task.ID, "url", task.URL)
	messages, result, err := q.handle(ctx, task)
	if err != nil {
		cb <- MediaTaskResult{
			TaskID: task.ID,
			Ok:     false,
			Error:  err,
		}
	} else {
		cb <- MediaTaskResult{
			TaskID:      task.ID,
			Ok:          true,
			Title:       result.Title,
			Description: result.Description,
			Messages:    messages,
		}
	}
}

func (q *LocalQueue) handle(ctx context.Context, task MediaTask) ([]*telegram.Message, *ytdlp.Response, error) {
	r, err := q.GetFormat(ctx, task.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get format: %w", err)
	}

	if r.MediaType == "livestream" {
		return nil, nil, fmt.Errorf("live stream is not supported yet")
	}

	key := fmt.Sprintf("%s.id.%s.bot.%s.messages", strings.ToLower(r.Extractor), r.ID, q.bot.Username)
	cache, err := q.cache.Get(ctx, key)
	if err == nil {
		var messages []*telegram.Message
		if err = json.Unmarshal(cache.Value, &messages); err != nil {
			return nil, nil, err
		}
		q.l.Info("got messages from cache", "key", key, "count", len(messages))
		return messages, r, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	// No messages in cache found, try to upload it to channel
	video, err := q.yd.DownloadFormat(r.FormatID, r)
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
			q.l.Error("failed to crop thumbnail", "error", err)
		}
	}

	caption, _ := json.MarshalIndent(task, "", "  ")

	m, err := q.bot.SendVideo(ctx, &telegram.SendVideoParams{
		ChatID: q.chatID,
		Video: telegram.InputFileLocal{
			Name:   video.Name,
			Reader: video.File,
		},
		Duration:          int(r.Duration),
		Width:             r.Width,
		Height:            r.Height,
		Thumbnail:         t,
		Caption:           fmt.Sprintf("<pre>%s</pre>", caption),
		ParseMode:         telegram.ParseModeHTML,
		SupportsStreaming: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send video: %w", err)
	}

	q.l.Info("video sent successfully", "extractor", r.Extractor, "size", r.Filesize, "duration", r.Duration)

	if m.Video == nil {
		return nil, nil, fmt.Errorf("no video in outgoing message found")
	}

	messages := []*telegram.Message{m}
	data, err := json.Marshal(messages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode videos: %w", err)
	}

	_, err = q.cache.Set(ctx, repository.SetParams{
		Key:   key,
		Value: data,
	})
	return messages, r, err
}

func (q *LocalQueue) GetFormat(ctx context.Context, url string) (*ytdlp.Response, error) {
	cache, err := q.cache.Get(ctx, url)

	if err == nil {
		var r *ytdlp.Response
		if err = json.Unmarshal(cache.Value, &r); err != nil {
			return nil, err
		}
		q.l.Info("got yt-dlp response from cache", "key", url)
		return r, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	data, err := q.yd.GetFormatRaw(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get format: %w", err)
	}

	var r *ytdlp.Response
	if err = json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("failed to encode yt-dlp response: %w", err)
	}

	_, err = q.cache.Set(ctx, repository.SetParams{
		Key:   url,
		Value: data,
	})

	return r, nil
}
