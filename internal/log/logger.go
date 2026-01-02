package log

import (
	"log/slog"
	"os"
	"path/filepath"
)

func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		AddSource:   true,
		ReplaceAttr: replaceAttr,
	}

	return slog.New(slog.NewTextHandler(os.Stderr, opts))
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.SourceKey {
		source := a.Value.Any().(*slog.Source)
		source.File = filepath.Base(source.File)
		source.Function = filepath.Base(source.Function)
	}
	return a
}
