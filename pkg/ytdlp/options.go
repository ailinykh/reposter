package ytdlp

import (
	"log/slog"
)

func WithProxyList(proxyList *ProxyList) func(*YtDlp) {
	return func(yd *YtDlp) {
		yd.proxies = proxyList
	}
}

func WithLogger(l *slog.Logger) func(*YtDlp) {
	return func(yd *YtDlp) {
		yd.l = l
	}
}
