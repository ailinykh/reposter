package hotlink

import (
	"fmt"
	"net/url"
	"os"
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

	if r.MediaType == "livestream" {
		return fmt.Errorf("live stream is not supported yet")
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

	video, err := h.yd.DownloadFormat(r.FormatID, r)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer video.Close()

	caption := fmt.Sprintf("<a href=\"%s\">🎞</a> <b>%s</b> <i>(by %s)</i>\n\n%s", r.OriginalUrl, r.Title, m.From.DisplayName(), r.Description)
	if len(caption) > 1024 {
		caption = caption[:1024]
	}
	caption = strings.ToValidUTF8(caption, "")

	v, err := os.Open(video.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open local video %s: %w", video.FilePath, err)
	}
	defer video.Close()

	thumb, err := os.Open(video.Thumb.FilePath)
	if err != nil {
		return fmt.Errorf("failed to open local video thumb %s: %w", video.Thumb.FilePath, err)
	}
	defer thumb.Close()

	_, err = bot.SendVideo(telegram.SendVideoParams{
		ChatID: m.Chat.ID,
		Video: telegram.InputFileLocal{
			Name:   video.FileName,
			Reader: v,
		},
		Duration: r.Duration,
		Width:    r.Width,
		Height:   r.Height,
		Thumbnail: telegram.InputFileLocal{
			Name:   video.Thumb.FileName,
			Reader: thumb,
		},
		Caption:           caption,
		ParseMode:         telegram.ParseModeHTML,
		SupportsStreaming: true,
	})
	return err
}
