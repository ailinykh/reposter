package main

import (
	"context"
	"log/slog"

	"github.com/ailinykh/reposter/v3/internal/fotd"
	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

type UpdateHandler interface {
	Handle(*telegram.Update, *telegram.Bot) error
}

func makeHandlers(
	ctx context.Context,
	logger *slog.Logger,
	repo *repository.Queries,
) []UpdateHandler {
	handlers := []UpdateHandler{
		fotd.NewGame(ctx, logger, repo),
	}

	return handlers
}
