package ytdlp

import "log/slog"

func WithArgs(args []string) func(*YtDlp) {
	return func(yd *YtDlp) {
		yd.args = append(yd.args, args...)
	}
}

func WithLogger(l *slog.Logger) func(*YtDlp) {
	return func(yd *YtDlp) {
		yd.l = l
	}
}
