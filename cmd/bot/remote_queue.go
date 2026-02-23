package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ailinykh/reposter/v3/internal/hotlink"
	"github.com/ailinykh/reposter/v3/pkg/rabbitmq"
)

func NewRemoteQueue(url string, logger *slog.Logger) (*RemoteQueue, error) {
	r, err := rabbitmq.New(url)
	if err != nil {
		return nil, err
	}

	cb := make(chan hotlink.MediaTaskResult)

	msgs, err := r.Consume("reposter_callback")
	if err != nil {
		return nil, err
	}

	go func() {
		for d := range msgs {
			logger.Debug("⬅️ task completed", "data", d.Body)

			var result hotlink.MediaTaskResult
			if err := json.Unmarshal(d.Body, &result); err != nil {
				logger.Error("failed to unmarshal task result", "error", err)
				continue
			}

			cb <- result
		}
	}()

	return &RemoteQueue{
		cb: cb,
		l:  logger,
		r:  r,
	}, nil
}

type RemoteQueue struct {
	cb chan hotlink.MediaTaskResult
	l  *slog.Logger
	r  *rabbitmq.Rabbit
}

func (q *RemoteQueue) Consume(_ context.Context, task hotlink.MediaTask) (chan hotlink.MediaTaskResult, error) {
	body, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := q.r.Publish("reposter", body); err != nil {
		return nil, fmt.Errorf("failed to publish a task: %w", err)
	}

	return q.cb, nil
}
