package logger

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
	"github.com/google/logger"
)

func NewGoogleLogger(ctx context.Context, workingDir string) core.Logger {
	logFilePath := path.Join(workingDir, "reposter.log")
	lf, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		time.Sleep(1 * time.Millisecond)
		logger.Info("closing logger")
		logger.Close()
	}()

	return &GoogleLogger{
		l: logger.Init("reposter", true, false, lf),
	}
}

type GoogleLogger struct {
	l *logger.Logger
}

func (GoogleLogger) join(values ...interface{}) string {
	strs := make([]string, len(values))
	for i, v := range values {
		sep := " "
		if i%2 != 0 {
			sep = "="
		}
		strs[i] = fmt.Sprintf("%v%s", v, sep)
	}
	return strings.Join(strs, "")
}

func (l *GoogleLogger) Debug(v ...interface{}) {
	l.l.InfoDepth(1, v...)
}

func (l *GoogleLogger) Error(v ...interface{}) {
	l.l.ErrorDepth(1, l.join(v...))
}

func (l *GoogleLogger) Info(v ...interface{}) {
	l.l.InfoDepth(1, l.join(v...))
}

func (l *GoogleLogger) Warn(v ...interface{}) {
	l.l.WarningDepth(1, l.join(v...))
}
