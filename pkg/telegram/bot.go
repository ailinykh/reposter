package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func NewBot(opts ...func(*BotConfig)) (*Bot, error) {
	config := NewBotConfig(opts...)

	me, err := getMe(config)
	if err != nil {
		return nil, err
	}
	return &Bot{
		User:     me,
		client:   config.client,
		ctx:      config.ctx,
		endpoint: config.endpoint,
		token:    config.token,
		l:        config.logger.With("username", me.Username),
	}, nil
}

type Bot struct {
	*User
	client   *http.Client
	ctx      context.Context
	endpoint string
	token    string
	l        *slog.Logger
}

func getMe(config *BotConfig) (*User, error) {
	resp, err := config.client.Get(config.endpoint + "/bot" + config.token + "/getMe")
	if err != nil {
		return nil, fmt.Errorf("failed to connect Telegram API %w", err)
	}
	defer resp.Body.Close()

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
		Result      User   `json:"result"`
	}
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json %w", err)
	}

	if !r.Ok {
		return nil, fmt.Errorf("telegram error: %s", r.Description)
	}

	return &r.Result, nil
}

func (b *Bot) GetUpdates(offset, timeout int64) ([]*Update, error) {
	urlString := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d&timeout=%d", b.endpoint, b.token, offset, timeout)
	b.l.Debug("start polling...", "offset", offset, "timeout", timeout)

	var r struct {
		Ok          bool      `json:"ok"`
		Description string    `json:"description"`
		Result      []*Update `json:"result"`
	}
	err := b.do("GET", urlString, &r)
	if err != nil {
		return nil, err
	}

	if !r.Ok {
		return nil, fmt.Errorf("telegram error: %s", r.Description)
	}

	return r.Result, nil
}

func (b *Bot) do(method, url string, result any) error {
	request, err := http.NewRequestWithContext(b.ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := b.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}
