package http_logger

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/ailinykh/pullanusbot/v2/internal/core"
	"github.com/ailinykh/pullanusbot/v2/internal/legacy/infrastructure"
)

func NewLoggingRoundTripper(rt http.RoundTripper, ls *infrastructure.LogStorage, l core.Logger) http.RoundTripper {
	return &LoggingRoundTripper{
		rt: rt,
		ls: ls,
		l:  l,
	}
}

type LoggingRoundTripper struct {
	rt http.RoundTripper
	ls *infrastructure.LogStorage
	l  core.Logger
}

func (l *LoggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	if err == nil {
		req.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	req.Body.Close()

	resp, err := l.rt.RoundTrip(req)

	if err != nil {
		l.l.Error("Request failed", "error", err, "method", req.Method, "url", req.URL)
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(data))

	if !utf8.Valid(data) {
		return resp, nil
	}

	rawJson := string(data)
	if rawJson == `{"ok":true,"result":[]}` {
		return resp, nil
	}

	parts := strings.Split(req.URL.Path, "/")
	err = l.ls.LogRecord(parts[len(parts)-1], string(body), rawJson)
	if err != nil {
		l.l.Error("failed to save logs", "error", err, "path", req.URL.Path, "data", data)
	}

	return resp, nil
}
