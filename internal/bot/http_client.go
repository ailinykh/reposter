package bot

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

func NewHttpClient(logger *slog.Logger, repo Repo) *http.Client {
	transport := http.DefaultTransport
	if _, ok := os.LookupEnv("ENABLE_TRACES"); ok {
		logger.Info("http traces enabled")
		transport = NewLoggingRoundTripper(
			http.DefaultTransport,
			logger,
			repo,
		)
	}
	return &http.Client{
		Transport: transport,
		Timeout:   time.Minute * 10,
	}
}
