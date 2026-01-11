package hotlink

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"slices"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

var ErrURLNotSupported = errors.New("url not supported")

func New(l *slog.Logger) *Handler {
	return &Handler{
		l: l,
	}
}

type Handler struct {
	l *slog.Logger
}

func (h *Handler) Handle(u *telegram.Update, bot *telegram.Bot) error {
	if u.Message == nil || len(u.Message.URLs()) == 0 {
		return nil
	}

	// TODO: Multiple `url` support
	urlString := u.Message.URLs()[0]

	err := h.handleSocial(urlString, u.Message, bot)
	if errors.Is(err, ErrURLNotSupported) {
		return h.handleHotlink(urlString, u.Message, bot)
	}
	return err
}

func (h *Handler) handleSocial(urlString string, m *telegram.Message, bot *telegram.Bot) error {
	url, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	h.l.Info("got url", "hostname", url.Hostname())

	supportedHostnames := []string{
		"instagram.com",
		"tiktok.com",
		"twitter.com",
		"youtube.com",
		"youtu.be",
		"x.com",
	}

	if !slices.Contains(supportedHostnames, url.Hostname()) {
		h.l.Info("url not supported yet", "hostname", url.Hostname(), "url", url)
		return ErrURLNotSupported
	}

	return nil
}

func (h *Handler) handleHotlink(url string, m *telegram.Message, bot *telegram.Bot) error {
	res, err := http.DefaultClient.Head(url)
	if err != nil {
		h.l.Error("failed to perform HEAD request", "url", url, "error", err)
		return nil
	}

	contentType := res.Header.Get("Content-Type")
	if len(contentType) == 0 {
		h.l.Error("content type not found", "url", url)
		return nil
	}

	h.l.Info("got contentType", "contentType", contentType)

	if strings.HasPrefix(contentType, "video") {
		_, err = bot.SendVideo(m.Chat.ID, url, fmt.Sprintf(`<a href="%s">🔗</a> <b>%s</b> <i>(by %s)</i>`, url, path.Base(url), m.From.DisplayName()))
		return err
	}

	if strings.HasPrefix(contentType, "image") {
		_, err = bot.SendPhoto(m.Chat.ID, url, fmt.Sprintf(`<a href="%s">🖼</a> <b>%s</b> <i>(by %s)</i>`, url, path.Base(url), m.From.DisplayName()))
		return err
	}

	h.l.Info("unsupported content-type", "content_type", contentType)
	return nil
}
