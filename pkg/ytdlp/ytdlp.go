package ytdlp

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func New(opts ...func(*YtDlp)) *YtDlp {
	y := &YtDlp{
		args: []string{
			"yt-dlp",
			"--ignore-config",
			"-t", "sleep",
			"-t", "mp4",
		},
		l: slog.Default(),
	}
	for _, o := range opts {
		o(y)
	}
	return y
}

type YtDlp struct {
	args []string
	l    *slog.Logger
}

func (yd *YtDlp) GetFormat(url string) (r *Response, err error) {
	cmd := strings.Join(append(yd.args, "--dump-json", url), " ")
	yd.l.Debug("executing", "command", cmd)

	out, err := exec.Command("/bin/sh", "-c", cmd).Output()
	if err != nil {
		yd.l.Error("failed to dump json", "url", url, "output", out, "error", err)
		return nil, fmt.Errorf("failed to dump json: %w", NewError(err, out))
	}

	err = json.Unmarshal(out, &r)
	if err != nil {
		yd.l.Error("unexpected content", "text", out)
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	return r, nil
}

func (yd *YtDlp) DownloadFormat(formatID string, resp *Response) (string, error) {
	dirPath, err := os.MkdirTemp("", "yt-dlp*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary directory: %w", err)
	}

	cmd := strings.Join(append(yd.args,
		"--embed-metadata",
		"--embed-thumbnail",
		"--convert-thumbnails", "jpg",
		"--write-thumbnail",
		"--write-info-json",
		// "--restrict-filenames", // removes all cyrillic letters
		"-f", formatID,
		"-P", dirPath,
		"-o", `"%(title)s.%(ext)s"`,
		resp.WebpageUrl,
	), " ")
	yd.l.Debug("executing", "command", strings.Replace(cmd, os.TempDir(), "$TMPDIR/", 1))

	if out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err != nil {
		yd.l.Error("failed to download video", "extractor", resp.Extractor, "format_id", formatID, "url", resp.WebpageUrl, "output", out)
		return "", fmt.Errorf("failed to dump json: %w", NewError(err, out))
	}

	yd.l.Info("video downloaded successfully", "extractor", resp.Extractor, "format_id", formatID, "id", resp.ID)

	return dirPath, nil
}
