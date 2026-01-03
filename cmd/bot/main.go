package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ailinykh/reposter/v3/internal/log"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := log.NewLogger()

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
	startRunLoop(ctx, bot, logger)
	logger.Info("attempt to shutdown gracefully...")
}

func startRunLoop(ctx context.Context, bot *telegram.Bot, logger *slog.Logger) {
	var offset int64 = 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			updates, err := bot.GetUpdates(offset, 300)
			if err != nil {
				logger.Error("failed to get updates", "error", err)
				break
			}

			for _, update := range updates {
				logger.Info("got update", "update", update)
				offset = update.ID + 1
			}
		}
	}
}
