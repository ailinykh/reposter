package log

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"
)

func NewLoggingRoundTripper(rt http.RoundTripper, l *slog.Logger) http.RoundTripper {
	return &LoggingRoundTripper{
		rt: rt,
		l:  l,
	}
}

type LoggingRoundTripper struct {
	rt http.RoundTripper
	l  *slog.Logger
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	path := req.URL.Path[strings.LastIndex(req.URL.Path, "/"):]
	l.l.Debug("performing request", "method", req.Method, "telegram_method", path, "content_type", req.Header.Get("Content-Type"))
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			l.l.Error("failed to read request body", "error", err, "method", req.Method, "telegram_method", path)
			return nil, err
		}
		req.Body = io.NopCloser(bytes.NewBuffer(body))
		if req.Header.Get("Content-Type") == "application/json" {
			l.l.Debug("sending data", "data", body)
		} else {
			l.l.Debug("sending bytes", "count", len(body))
		}
	}

	resp, err := l.rt.RoundTrip(req)
	if err != nil {
		l.l.Error("roundtrip failed", "error", err, "method", req.Method, "telegram_method", path)
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		l.l.Error("failed to read response body", "error", err, "method", req.Method, "telegram_method", path)
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(data))

	if !utf8.Valid(data) {
		return resp, nil
	}

	l.l.Debug("success", "json", string(data))

	return resp, nil
}
