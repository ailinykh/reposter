package ytdlp

import (
	"os/exec"
	"strings"
)

type Error struct {
	Code        int
	Description string
}

func (err *Error) Error() string {
	return err.Description
}

func NewError(err error, data []byte) *Error {
	execErr, ok := err.(*exec.ExitError)
	if !ok {
		return &Error{Code: -1, Description: err.Error()}
	}

	stdErr := strings.ToLower(string(execErr.Stderr))
	if stdErr == "" {
		stdErr = string(data)
	}

	if strings.Contains(stdErr, "sign in to confirm you’re not a bot") {
		return &Error{Code: 403, Description: "Sign in to confirm that you're not a bot"}
	}
	if strings.Contains(stdErr, "sign in to confirm your age") {
		return &Error{Code: 403, Description: "Sign in to confirm your age. This video may be inappropriate for some users"}
	}
	if strings.Contains(stdErr, "video unavailable") {
		return &Error{Code: 404, Description: "This video isn't available any more"}
	}
	if strings.Contains(stdErr, "HTTP Error 503: Service Unavailable") {
		return &Error{Code: 503, Description: "Service Unavailable"}
	}
	if strings.Contains(stdErr, "private video") {
		return &Error{Code: 401, Description: "Video unavailable. This video is private"}
	}

	return &Error{Code: -1, Description: err.Error()}
}
