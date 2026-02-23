package bot

import (
	"context"
	"log/slog"

	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

type Repo interface {
	CreateBotTrace(ctx context.Context, arg repository.CreateBotTraceParams) error
}

func New(ctx context.Context, token string, logger *slog.Logger, repo Repo) *telegram.Bot {
	bot, err := telegram.NewBot(
		ctx,
		telegram.WithToken(token),
		telegram.WithLogger(logger),
		telegram.WithClient(NewHttpClient(logger, repo)),
	)

	if err != nil {
		panic(err)
	}

	logger.Info("bot created", "username", bot.Username)

	return bot
}
