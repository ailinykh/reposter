package ytdlp

import (
	"log/slog"
)

func WithProxyList(proxyList []string) func(*YtDlp) {
	return func(yd *YtDlp) {
		yd.proxies = NewProxyList(proxyList)
	}
}

func WithLogger(l *slog.Logger) func(*YtDlp) {
	return func(yd *YtDlp) {
		yd.l = l
	}
}
