package telegram

import (
	"context"
	"log/slog"
	"net/http"
)

func WithToken(token string) func(*Bot) {
	return func(b *Bot) {
		b.token = token
	}
}

func WithClient(client *http.Client) func(*Bot) {
	return func(b *Bot) {
		b.client = client
	}
}

func WithContext(ctx context.Context) func(*Bot) {
	return func(b *Bot) {
		b.ctx = ctx
	}
}

func WithLogger(logger *slog.Logger) func(*Bot) {
	return func(b *Bot) {
		b.l = logger
	}
}
