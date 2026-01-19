package ytdlp

import (
	"fmt"
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

func NewError(err error, data []byte) error {
	execErr, ok := err.(*exec.ExitError)
	if !ok {
		return &Error{Code: -1, Description: err.Error()}
	}

	stdErr := strings.ToLower(string(execErr.Stderr))
	if stdErr == "" {
		stdErr = string(data)
	}

	errors := []struct {
		code        int
		substring   string
		description string
	}{
		{403, "sign in to confirm you’re not a bot", "Sign in to confirm that you're not a bot"},
		{403, "sign in to confirm your age", "Sign in to confirm your age. This video may be inappropriate for some users"},
		{404, "removed by the uploader", "Video unavailable. This video has been removed by the uploader"},
		{404, "account associated with this video", "Video unavailable. This video is no longer available because the YouTube account associated with this video has been terminated."},
		{404, "removed for violating", "This video has been removed for violating YouTube's Terms of Service"},
		{404, "video unavailable", "This video isn't available any more"},
		{503, "503: Service Unavailable", "HTTP Error 503: Service Unavailable"},
		{401, "private video", "Video unavailable. This video is private"},
	}

	for _, e := range errors {
		if strings.Contains(stdErr, e.substring) {
			return &Error{Code: e.code, Description: e.description}
		}
	}

	return fmt.Errorf("unexpected error: %w", err)
}
