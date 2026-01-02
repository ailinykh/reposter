package main

import (
	"context"
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
	)

	if err != nil {
		panic(err)
	}

	logger.Info("bot created", "username", bot.Username)

	<-ctx.Done()
	logger.Info("attempt to shutdown gracefully...")
}
