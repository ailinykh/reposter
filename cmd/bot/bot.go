package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func NewBot(ctx context.Context, logger *slog.Logger) *telegram.Bot {
	bot, err := telegram.NewBot(
		telegram.WithToken(os.Getenv("TELEGRAM_BOT_TOKEN")),
		telegram.WithLogger(logger),
		telegram.WithClient(NewHttpClient(logger)),
		telegram.WithContext(ctx),
	)

	if err != nil {
		panic(err)
	}

	logger.Info("bot created", "username", bot.Username)

	return bot
}
