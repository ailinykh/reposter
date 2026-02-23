package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/ailinykh/reposter/v3/internal/bot"
	"github.com/ailinykh/reposter/v3/internal/log"
	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := log.NewLogger()
	bot := bot.New(ctx, os.Getenv("TELEGRAM_BOT_TOKEN"), logger)
	repo := repository.New(NewDB(logger))

	startRunLoop(ctx, logger, bot, makeHandlers(logger, repo, bot))
	logger.Info("attempt to shutdown gracefully...")
}

func startRunLoop(
	ctx context.Context,
	logger *slog.Logger,
	bot *telegram.Bot,
	handlers []UpdateHandler,
) {
	var offset int64 = 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
			updates, err := bot.GetUpdates(ctx, &telegram.GetUpdatesParams{Offset: offset, Timeout: 300})
			if err != nil {
				logger.Error("failed to get updates", "error", err)
				break
			}

			for _, update := range updates {
				logger.Info("processing update", "update_id", update.ID, "message", update.Message)
				for _, handler := range handlers {
					if err := handler.Handle(ctx, update, bot); err != nil {
						logger.Error("❌ failed to process update", "handler", handler, "error", err)
					}
				}
				offset = update.ID + 1
			}
			logger.Info("all updates processed", "current_offset", offset)
		}
	}
}
