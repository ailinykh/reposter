package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ailinykh/reposter/v3/internal/log"
)

func NewHttpClient(logger *slog.Logger) *http.Client {
	transport := http.DefaultTransport
	if _, ok := os.LookupEnv("EXTENDED_LOGGING_ENABLED"); ok {
		logger.Info("extended logging enabled")
		transport = log.NewLoggingRoundTripper(
			http.DefaultTransport,
			logger,
		)
	}
	return &http.Client{
		Transport: transport,
		Timeout:   time.Minute,
	}
}
