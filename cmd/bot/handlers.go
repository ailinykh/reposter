package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/ailinykh/reposter/v3/internal/fotd"
	"github.com/ailinykh/reposter/v3/internal/hotlink"
	"github.com/ailinykh/reposter/v3/internal/info"
	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/internal/xui"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

type UpdateHandler interface {
	Handle(context.Context, *telegram.Update, *telegram.Bot) error
}

func makeHandlers(
	ctx context.Context,
	logger *slog.Logger,
	repo *repository.Queries,
) []UpdateHandler {
	handlers := []UpdateHandler{
		fotd.NewGame(logger.With("handler", "fotd"), repo),
		info.New(),
		hotlink.New(
			logger.With("handler", "hotlink"),
			repo,
			ytdlp.New(
				ytdlp.WithArgs(getYtDlpArgs()),
				ytdlp.WithLogger(logger.With("tool", "yt-dlp")),
			),
		),
	}

	baseUrl := os.Getenv("XUI_BASE_URL")
	login := os.Getenv("XUI_LOGIN")
	password := os.Getenv("XUI_PASSWORD")
	if baseUrl != "" && login != "" && password != "" {
		logger.Info("xui vpn logic enabled", "username", login)
		client := xui.NewClient(logger.With("handler", "xui"), baseUrl, login, password)
		handlers = append(handlers, xui.NewHandler(client, logger.With("handler", "xui"), repo))
	} else {
		logger.Info("xui vpn logic disabled")
	}

	return handlers
}

func getYtDlpArgs() []string {
	var args = []string{}
	if value, ok := os.LookupEnv("PROXY"); ok {
		args = append(args, "--proxy", value)
	}
	return args
}
