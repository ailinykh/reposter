package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"

	"github.com/ailinykh/reposter/v3/internal/bot"
	"github.com/ailinykh/reposter/v3/internal/database"
	"github.com/ailinykh/reposter/v3/internal/hotlink"
	"github.com/ailinykh/reposter/v3/internal/log"
	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/rabbitmq"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := log.NewLogger()
	bot := bot.New(ctx, os.Getenv("TELEGRAM_BOT_TOKEN_2"), logger)
	repo := repository.New(NewDB(logger))

	chatID, err := strconv.ParseInt(os.Getenv("DEFAULT_CHAT_ID"), 10, 64)
	if err != nil {
		logger.Warn("hotlink logic disabled", "error", err)
	}

	handler := hotlink.NewTaskHandler(
		chatID,
		logger,
		bot,
		repo,
		ytdlp.New(
			ytdlp.WithProxyList(ytdlp.NewProxyList("PROXY")),
			ytdlp.WithLogger(logger.With("tool", "yt-dlp")),
		),
	)

	if err := run(ctx, logger, handler); err != nil {
		logger.Error("worker failed", "error", err)
	}

	logger.Info("attempt to shutdown gracefully...")
}

func run(ctx context.Context, logger *slog.Logger, taskHandler *hotlink.TaskHandler) error {
	rabbit, err := rabbitmq.New(os.Getenv("AMQP_URL"))
	if err != nil {
		return err
	}

	msgs, err := rabbit.Consume("reposter")

	logger.Info("running worker")

	for {
		select {
		case <-ctx.Done():
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel is closed and drained")
			}

			logger.Debug("⬅️ processing", "data", d.Body)
			var task hotlink.MediaTask
			if err := json.Unmarshal(d.Body, &task); err != nil {
				return fmt.Errorf("failed to unmarshal task %w", err)
			}

			logger.Info("processing task", "task_id", task.ID)
			taskResult := taskHandler.Process(ctx, task)
			body, err := json.Marshal(taskResult)
			if err != nil {
				return fmt.Errorf("failed to marshal task result %w", err)
			}

			if err := rabbit.Publish("reposter_callback", body); err != nil {
				return fmt.Errorf("failed to publish message: %w", err)
			}

			logger.Debug("➡️ task processed", "data", body)
		}
	}
}

func NewDB(logger *slog.Logger) *sql.DB {
	db, err := database.New(logger,
		database.WithURL(os.Getenv("DATABASE_URL")),
		database.WithMigrations(database.Migrations),
	)
	if err != nil {
		panic(err)
	}
	return db
}
