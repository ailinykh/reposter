package hotlink

import (
	"errors"
	"fmt"
	"time"
)

var ErrURLNotSupported = errors.New("url not supported")

type VideoTooLongError struct {
	Duration time.Duration
	Title    string
}

func (e *VideoTooLongError) Error() string {
	return fmt.Sprintf("video %s too long: %d sec", e.Title, e.Duration)
}
