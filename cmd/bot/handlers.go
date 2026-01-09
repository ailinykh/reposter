package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ailinykh/reposter/v3/internal/fotd"
	"github.com/ailinykh/reposter/v3/internal/info"
	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/internal/xui"
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
		info.New(),
	}

	baseUrl := os.Getenv("XUI_BASE_URL")
	login := os.Getenv("XUI_LOGIN")
	password := os.Getenv("XUI_PASSWORD")
	if len(baseUrl) > 0 && len(login) > 0 && len(password) > 0 {
		logger.Info("xui vpn logic enabled", "username", login)
		client := xui.NewClient(ctx, logger, baseUrl, login, password)
		handlers = append(handlers, xui.NewHandler(client, logger, repo))
	} else {
		logger.Info("xui vpn logic disabled")
	}

	return handlers
}
