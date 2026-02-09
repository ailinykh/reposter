package hotlink

import (
	"context"
	"time"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

type MediaTask struct {
	ID   string    `json:"id"`
	Date time.Time `json:"date"`
	URL  string    `json:"url"`
}

type MediaTaskResult struct {
	TaskID      string `json:"task_id"`
	Ok          bool   `json:"ok"`
	Error       error  `json:"error,omitempty"`
	Title       string
	Description string
	Messages    []*telegram.Message `json:"messages,omitempty"`
}

type Queue interface {
	Consume(context.Context, MediaTask, chan MediaTaskResult) error
}
