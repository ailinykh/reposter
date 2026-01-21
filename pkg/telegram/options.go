package telegram

import (
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

func WithLogger(logger *slog.Logger) func(*Bot) {
	return func(b *Bot) {
		b.l = logger
	}
}
