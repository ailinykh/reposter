package telegram

import (
	"context"
	"log/slog"
	"net/http"
)

func NewBotConfig(opts ...func(*BotConfig)) *BotConfig {
	config := &BotConfig{
		endpoint: "https://api.telegram.org",
		token:    "",
		client:   http.DefaultClient,
		ctx:      context.Background(),
		logger:   slog.Default(),
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

type BotConfig struct {
	endpoint string
	token    string

	client *http.Client
	ctx    context.Context
	logger *slog.Logger
}

func WithToken(token string) func(*BotConfig) {
	return func(bc *BotConfig) {
		bc.token = token
	}
}

func WithClient(client *http.Client) func(*BotConfig) {
	return func(bc *BotConfig) {
		bc.client = client
	}
}

func WithContext(ctx context.Context) func(*BotConfig) {
	return func(bc *BotConfig) {
		bc.ctx = ctx
	}
}

func WithLogger(logger *slog.Logger) func(*BotConfig) {
	return func(bc *BotConfig) {
		bc.logger = logger
	}
}
