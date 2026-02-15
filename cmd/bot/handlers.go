package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/ailinykh/reposter/v3/internal/fotd"
	"github.com/ailinykh/reposter/v3/internal/hotlink"
	"github.com/ailinykh/reposter/v3/internal/info"
	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/internal/xui"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/xcom"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

type UpdateHandler interface {
	Handle(context.Context, *telegram.Update, *telegram.Bot) error
}

func makeHandlers(
	logger *slog.Logger,
	repo *repository.Queries,
	bot *telegram.Bot,
) []UpdateHandler {
	handlers := []UpdateHandler{
		fotd.NewGame(logger.With("handler", "fotd"), repo),
		info.New(),
	}

	chatID, err := strconv.ParseInt(os.Getenv("DEFAULT_CHAT_ID"), 10, 64)
	if err != nil {
		logger.Warn("hotlink logic disabled", "error", err)
	} else {
		logger.Info("running hotlink", "chat_id", chatID)
		handlers = append(handlers, hotlink.New(
			logger.With("handler", "hotlink"),
			hotlink.NewLocalQueue(
				logger,
				hotlink.NewTaskHandler(
					chatID,
					logger,
					bot,
					repo,
					ytdlp.New(
						ytdlp.WithProxyList(proxyList()),
						ytdlp.WithLogger(logger.With("tool", "yt-dlp")),
					),
				),
			),
			xcom.New(logger),
		),
		)
	}

	baseUrl := os.Getenv("XUI_BASE_URL")
	login := os.Getenv("XUI_LOGIN")
	password := os.Getenv("XUI_PASSWORD")
	inboundID, err := strconv.Atoi(os.Getenv("XUI_INBOUND_ID"))
	if baseUrl != "" && login != "" && password != "" && err == nil {
		logger.Info("xui vpn logic enabled", "username", login)
		client := xui.NewClient(logger.With("handler", "xui"), baseUrl, login, password)
		handlers = append(handlers, xui.New(client, inboundID, logger.With("handler", "xui"), repo))
	} else {
		logger.Info("xui vpn logic disabled")
	}

	return handlers
}

func proxyList() []string {
	var list = []string{}
	for _, env := range os.Environ() {
		if idx := strings.Index(env, "="); idx > 0 {
			if strings.HasPrefix(env[:idx], "PROXY") {
				list = append(list, env[idx+1:])
			}
		}
	}
	return list
}
