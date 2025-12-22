package logger

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
)

func NewSlogLogger() core.Logger {
	// https://pkg.go.dev/golang.org/x/exp/slog#example-package-Wrapping
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
			source.Function = filepath.Base(source.Function)
		}
		return a
	}
	opts := &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		AddSource:   true,
		ReplaceAttr: replace,
	}
	return &SLogger{
		l: *slog.New(slog.NewTextHandler(os.Stderr, opts)),
	}
}

type SLogger struct {
	l slog.Logger
}

func (l *SLogger) Debug(v ...interface{}) {
	l.log(slog.LevelDebug, v...)
}

func (l *SLogger) Error(v ...interface{}) {
	l.log(slog.LevelError, v...)
}

func (l *SLogger) Warn(v ...interface{}) {
	l.log(slog.LevelWarn, v...)
}

func (l *SLogger) Info(v ...interface{}) {
	l.log(slog.LevelInfo, v...)
}

func (l *SLogger) log(level slog.Level, v ...any) {
	msg := "no content"
	if len(v) > 0 {
		if m, ok := v[0].(string); ok {
			msg = m
			v = v[1:]
		}
	}
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, log, Info]
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(v...)
	l.l.Handler().Handle(context.Background(), r)
}
