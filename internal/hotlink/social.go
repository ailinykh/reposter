package hotlink

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func (h *Handler) handleSocial(urlString string, m *telegram.Message, bot *telegram.Bot) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	h.l.Info("processing url", "hostname", url.Hostname(), "url", urlString)

	supportedHostnames := []string{
		"instagram.com",
		"tiktok.com",
		"twitter.com",
		"www.youtube.com",
		"youtube.com",
		"youtu.be",
		"x.com",
	}

	if !slices.Contains(supportedHostnames, url.Hostname()) {
		h.l.Info("url not supported yet", "hostname", url.Hostname(), "url", url)
		return ErrURLNotSupported
	}

	r, err := h.yd.GetFormat(urlString)
	if err != nil {
		return fmt.Errorf("failed to get format: %w", err)
	}

	const maxSize int64 = 50_000_000 // Telegram multipart/form-data limit
	if r.Filesize > maxSize {
		h.l.Warn("video too long", "id", r.ID, "extractor", r.Extractor, "size", r.Filesize, "duration", r.Duration)
		if !m.Chat.Private() {
			return nil // be silent in group chat
		}
		_, err := bot.SendMessage(m.Chat.ID, fmt.Sprintf("%s\n<b>⏳ video too long: %d sec</b>", r.Title, r.Duration), map[string]any{
			"reply_parameters": map[string]any{
				"message_id": m.ID,
				"quote":      urlString,
			},
			"parse_mode": "HTML",
		})
		return err
	}

	dirPath, err := h.yd.DownloadFormat(r.FormatID, r)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer os.RemoveAll(dirPath)

	caption := fmt.Sprintf("<a href=\"%s\">🎞</a> <b>%s</b> <i>(by %s)</i>\n\n%s", r.OriginalUrl, r.Title, m.From.DisplayName(), r.Description)
	if len(caption) > 1024 {
		caption = caption[:1024]
	}
	caption = strings.ToValidUTF8(caption, "")

	params := map[string]any{
		"caption":            caption,
		"duration":           r.Duration,
		"width":              r.Width,
		"height":             r.Height,
		"supports_streaming": true,
		"parse_mode":         "HTML",
		"chat_id":            m.Chat.ID,
	}

	var files []os.DirEntry
	if files, err = os.ReadDir(dirPath); err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, file := range files {
		filePath := path.Join(dirPath, file.Name())
		info, _ := file.Info()
		h.l.Info("checking file", "size", info.Size(), "file_path", strings.Replace(filePath, os.TempDir(), "$TMPDIR/", 1))
		if !strings.HasSuffix(filePath, ".mp4") && !strings.HasSuffix(filePath, ".jpg") {
			continue
		}
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", file.Name(), err)
		}
		if strings.HasSuffix(filePath, ".mp4") {
			params["video"] = f
		} else {
			params["thumb"] = f
		}
	}
	_, err = bot.SendVideoMultipart(params)
	return err
}
