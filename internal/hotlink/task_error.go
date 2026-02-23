package hotlink

import (
	"errors"
	"fmt"

	"github.com/ailinykh/reposter/v3/pkg/xcom"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
)

type MediaTaskError struct {
	NotifyUser bool   `json:"notify_user"`
	Message    string `json:"message"`
}

func (e *MediaTaskError) Error() string {
	return e.Message
}

func NewMediaTaskError(err error) *MediaTaskError {
	var e *MediaTaskError
	if errors.As(err, &e) {
		return e
	}

	var long *VideoTooLongError
	if errors.As(err, &long) {
		return &MediaTaskError{
			NotifyUser: true,
			Message:    fmt.Sprintf("%s\n<b>⏳ video too long: %d sec</b>", long.Title, long.Duration),
		}
	}

	var xErr *xcom.Error
	if errors.As(err, &xErr) {
		return &MediaTaskError{
			NotifyUser: true,
			Message:    fmt.Sprintf("😬 %s", xErr.Error()),
		}
	}

	var ytErr *ytdlp.Error
	if errors.As(err, &ytErr) {
		return &MediaTaskError{
			NotifyUser: true,
			Message:    fmt.Sprintf("😬 %s", ytErr.Error()),
		}
	}

	return &MediaTaskError{
		NotifyUser: false,
		Message:    err.Error(),
	}
}
