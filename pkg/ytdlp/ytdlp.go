package ytdlp

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
)

func New(opts ...func(*YtDlp)) *YtDlp {
	y := &YtDlp{
		proxies: &ProxyList{},
		l:       slog.Default(),
	}
	for _, o := range opts {
		o(y)
	}
	return y
}

type YtDlp struct {
	proxies *ProxyList
	l       *slog.Logger
}

func (yd *YtDlp) Exec(args ...string) (out []byte, err error) {
	args = append([]string{
		// default args
		"yt-dlp",
		"--ignore-config",
		"-t", "sleep",
		"-t", "mp4",
	}, args...)

	// add proxy if exists
	var proxy *Proxy
	if proxy = yd.proxies.Next(); proxy != nil {
		args = append(args, "--proxy", proxy.url)
	} else {
		yd.l.Warn("no working proxy found")
	}

	cmd := strings.Join(args, " ")
	yd.l.Debug("executing", "command", strings.Replace(cmd, os.TempDir(), "$TMPDIR/", 1))

	if out, err = exec.Command("/bin/sh", "-c", cmd).CombinedOutput(); err == nil {
		// This is a happy path!
		return out, err
	}

	yd.l.Error("failed to exec yt-dlp", "output", out, "error", err)
	err = NewError(err, out)

	// Mark proxy as banned
	if proxy != nil {
		var e *Error
		if errors.As(err, &e) && e.Code == 403 {
			yd.l.Debug("got 403 error using proxy", "error", e)
			yd.proxies.MarkBanned(proxy)
		}
	}

	return nil, err
}

func (yd *YtDlp) GetFormatRaw(url string) ([]byte, error) {
	out, err := yd.Exec("--quiet", "--no-warnings", "--dump-json", url)
	if err != nil {
		return nil, fmt.Errorf("failed to dump json: %w", err)
	}

	return out, nil
}

func (yd *YtDlp) GetFormat(url string) (r *Response, err error) {
	out, err := yd.GetFormatRaw(url)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out, &r)
	if err != nil {
		yd.l.Error("unexpected content", "text", out)
		return nil, fmt.Errorf("failed to parse json: %w", err)
	}

	return r, nil
}

func (yd *YtDlp) DownloadFormat(formatID string, resp *Response) (*LocalVideo, error) {
	dirPath, err := os.MkdirTemp("", "yt-dlp*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}

	if out, err := yd.Exec(
		"--embed-metadata",
		"--embed-thumbnail",
		"--convert-thumbnails", "jpg",
		"--write-thumbnail",
		"--write-info-json",
		// "--restrict-filenames", // removes all cyrillic letters
		"-f", formatID,
		"-P", dirPath,
		"-o", `"file.%(ext)s"`,
		resp.WebpageUrl,
	); err != nil {
		yd.l.Error("failed to download video", "extractor", resp.Extractor, "format_id", formatID, "url", resp.WebpageUrl, "output", out)
		return nil, fmt.Errorf("failed to download video: %w", err)
	}

	yd.l.Info("video downloaded successfully", "extractor", resp.Extractor, "format_id", formatID, "id", resp.ID)

	fPath := path.Join(dirPath, "file.mp4")
	f, err := os.Open(fPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local video %s: %w", fPath, err)
	}

	tPath := path.Join(dirPath, "file.jpg")
	t, err := os.Open(tPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local video thumb %s: %w", tPath, err)
	}

	return &LocalVideo{
		LocalFile: LocalFile{
			File: f,
			Name: "file.mp4",
			Path: fPath,
		},
		Thumb: LocalFile{
			File: t,
			Name: "file.jpg",
			Path: tPath,
		},
		dirPath: dirPath,
	}, nil
}
