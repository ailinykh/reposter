package bot

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ailinykh/reposter/v3/internal/repository"
)

func NewLoggingRoundTripper(rt http.RoundTripper, l *slog.Logger, repo Repo) http.RoundTripper {
	return &LoggingRoundTripper{
		T:      rt,
		logger: l,
		repo:   repo,
	}
}

type LoggingRoundTripper struct {
	T      http.RoundTripper
	logger *slog.Logger
	repo   Repo
}

func (rt *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.LogAndRoundTrip(req, rt.logger)
}

func (rt *LoggingRoundTripper) LogAndRoundTrip(req *http.Request, logger *slog.Logger) (resp *http.Response, err error) {
	telegramMethod := req.URL.Path[strings.LastIndex(req.URL.Path, "/"):]
	logger.Debug("performing request", "method", req.Method, "telegram_method", telegramMethod, "content_type", req.Header.Get("Content-Type"))
	var body []byte
	if req.Body != nil {
		body, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			logger.Error("failed to read request body", "error", err, "method", req.Method, "telegram_method", telegramMethod)
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		if req.Header.Get("Content-Type") == "application/json" {
			logger.Debug("sending data", "data", body)
		} else {
			logger.Debug("sending bytes", "count", len(body))
		}
	}

	resp, err = rt.T.RoundTrip(req)
	if err != nil {
		logger.Error("roundtrip failed", "error", err, "method", req.Method, "telegram_method", telegramMethod)
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body", "error", err, "method", req.Method, "telegram_method", telegramMethod)
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(data))

	if !utf8.Valid(data) {
		logger.Error("invalid utf8 data", "data", data)
		return resp, nil
	}

	// skip empty updates traces
	if bytes.Equal(data, []byte(`{"ok":true,"result":[]}`)) {
		return resp, nil
	}

	botID, _ := strconv.ParseInt((strings.TrimPrefix(req.URL.Path[:strings.Index(req.URL.Path, ":")], "/bot")), 10, 64)
	if err := rt.repo.CreateBotTrace(req.Context(), repository.CreateBotTraceParams{
		BotID:    botID,
		Method:   telegramMethod,
		Request:  ToRawMessagePtr(body),
		Response: data,
	}); err != nil {
		logger.Error("failed to trace request", "error", err)
	}

	logger.Debug("success", "json", string(data))

	return resp, nil
}

func ToRawMessagePtr(b []byte) *json.RawMessage {
	if b == nil {
		return nil
	}
	m := json.RawMessage(b)
	return &m
}
