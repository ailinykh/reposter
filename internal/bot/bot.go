package bot

import (
	"context"
	"log/slog"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func New(ctx context.Context, token string, logger *slog.Logger) *telegram.Bot {
	bot, err := telegram.NewBot(
		ctx,
		telegram.WithToken(token),
		telegram.WithLogger(logger),
		telegram.WithClient(NewHttpClient(logger)),
	)

	if err != nil {
		panic(err)
	}

	logger.Info("bot created", "username", bot.Username)

	return bot
}
