package main

import (
	"context"
	"log/slog"

	"github.com/ailinykh/reposter/v3/internal/hotlink"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func NewLocalQueue(
	l *slog.Logger,
	handler *hotlink.TaskHandler,
) hotlink.Queue {
	ch := make(chan LocalQueueTask)
	queue := LocalQueue{
		l:       l,
		ch:      ch,
		cb:      make(chan hotlink.MediaTaskResult),
		handler: handler,
	}

	go func() {
		for t := range ch {
			queue.l.Info("🔹 processing task", "task_id", t.task.ID, "url", t.task.URL)
			queue.process(t.ctx, t.task)
			queue.l.Info("🔸 task processed", "task_id", t.task.ID, "url", t.task.URL)
		}
	}()

	return &queue
}

type LocalQueue struct {
	bot     *telegram.Bot
	l       *slog.Logger
	ch      chan LocalQueueTask
	cb      chan hotlink.MediaTaskResult
	handler *hotlink.TaskHandler
}

type LocalQueueTask struct {
	task hotlink.MediaTask
	ctx  context.Context
}

func (q *LocalQueue) Consume(ctx context.Context, task hotlink.MediaTask) (chan hotlink.MediaTaskResult, error) {
	go func() {
		q.ch <- LocalQueueTask{
			task: task,
			ctx:  ctx,
		}
	}()
	return q.cb, nil
}

func (q *LocalQueue) process(ctx context.Context, task hotlink.MediaTask) {
	q.cb <- q.handler.Process(ctx, task)
}
