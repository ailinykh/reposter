package log

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
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
	l.l.Debug("request", "url", req.URL.Path)
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			l.l.Debug("sending", "body", body)
		}
	}

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

	l.l.Debug("success", "json", string(data))

	return resp, nil
}
