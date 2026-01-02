package telegram

import (
	"log/slog"
	"net/http"
)

func NewBotConfig(opts ...func(*BotConfig)) *BotConfig {
	config := &BotConfig{
		endpoint: "https://api.telegram.org",
		token:    "",
		client:   http.DefaultClient,
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
	logger *slog.Logger
}

func WithToken(token string) func(*BotConfig) {
	return func(bc *BotConfig) {
		bc.token = token
	}
}

func WithLogger(logger *slog.Logger) func(*BotConfig) {
	return func(bc *BotConfig) {
		bc.logger = logger
	}
}
