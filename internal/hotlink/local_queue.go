package hotlink

import (
	"context"
	"log/slog"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func NewLocalQueue(
	l *slog.Logger,
	handler *TaskHandler,
) Queue {
	ch := make(chan LocalQueueTask)
	queue := LocalQueue{
		l:       l,
		ch:      ch,
		handler: handler,
	}

	go func() {
		for t := range ch {
			queue.l.Info("🔹 processing task", "task_id", t.task.ID, "url", t.task.URL)
			queue.process(t.ctx, t.task, t.callback)
			queue.l.Info("🔸 task processed", "task_id", t.task.ID, "url", t.task.URL)
		}
	}()

	return &queue
}

type LocalQueue struct {
	bot     *telegram.Bot
	l       *slog.Logger
	ch      chan LocalQueueTask
	handler *TaskHandler
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
	messages, result, err := q.handler.Process(ctx, task)
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
