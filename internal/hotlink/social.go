package hotlink

import (
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/ffmpeg"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
	"github.com/ailinykh/reposter/v3/pkg/ytdlp"
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
		_, err = bot.SendMessage(telegram.SendMessageParams{
			ChatID:    m.Chat.ID,
			Text:      fmt.Sprintf("%s\n<b>⏳ video too long: %d sec</b>", r.Title, r.Duration),
			ParseMode: telegram.ParseModeHTML,
			ReplyParameters: &telegram.ReplyParameters{
				Quote: urlString,
			},
		})
		return err
	}

	video, err := h.yd.DownloadFormat(r.FormatID, r)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer video.Close()

	t := telegram.InputFileLocal{
		Name:   video.Thumb.Name,
		Reader: video.Thumb.File,
	}

	if r.MediaType == "short" {
		cropped, err := h.croppedThumb(r, video)
		if err == nil {
			t = telegram.InputFileLocal{
				Name:   video.Thumb.Name,
				Reader: cropped.File,
			}
			defer cropped.Close()
			defer os.Remove(cropped.Path)
		} else {
			h.l.Error("failed to crop thumbnail", "error", err)
		}
	}

	caption := fmt.Sprintf("<a href=\"%s\">🎞</a> <b>%s</b> <i>(by %s)</i>\n\n%s", r.OriginalUrl, r.Title, m.From.DisplayName(), r.Description)
	if len(caption) > 1024 {
		caption = caption[:1024]
	}
	caption = strings.ToValidUTF8(caption, "")

	_, err = bot.SendVideo(telegram.SendVideoParams{
		ChatID: m.Chat.ID,
		Video: telegram.InputFileLocal{
			Name:   video.Name,
			Reader: video.File,
		},
		Duration:          r.Duration,
		Width:             r.Width,
		Height:            r.Height,
		Thumbnail:         t,
		Caption:           caption,
		ParseMode:         telegram.ParseModeHTML,
		SupportsStreaming: true,
	})
	return err
}

func (h *Handler) croppedThumb(r *ytdlp.Response, v *ytdlp.LocalVideo) (*ytdlp.LocalFile, error) {
	info, err := ffmpeg.GetInfo(v.Thumb.Path)
	if err != nil {
		return nil, err
	}

	if len(info.Streams) < 1 {
		return nil, fmt.Errorf("no stream found at %s", v.Thumb.Path)
	}

	w := r.Width * info.Streams[0].Height / r.Height
	cropped, err := ffmpeg.Crop(v.Thumb.Path, w, info.Streams[0].Height)
	if err != nil {
		return nil, fmt.Errorf("failed to crop %s: %w", v.Thumb.Path, err)
	}

	f, err := os.Open(cropped)
	if err != nil {
		return nil, fmt.Errorf("failed to open cropped file: %w", err)
	}

	return &ytdlp.LocalFile{
		File: f,
		Path: cropped,
	}, nil
}
