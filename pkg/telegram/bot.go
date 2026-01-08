package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"maps"
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
		return nil, fmt.Errorf("failed to get bot data: %w", err)
	}
	defer resp.Body.Close()

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
		Result      User   `json:"result"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}

	if !r.Ok {
		return nil, fmt.Errorf("telegram error: %s", r.Description)
	}

	return &r.Result, nil
}

func chkErr(data []byte) error {
	var e struct {
		Ok          bool           `json:"ok"`
		Code        int            `json:"error_code"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}
	if err := json.Unmarshal(data, &e); err != nil {
		return fmt.Errorf("failed to parse error: %w", err)
	}
	if e.Ok {
		return nil
	}
	return fmt.Errorf("telegram error: %s", e.Description)
}

func (b *Bot) GetUpdates(offset, timeout int64) ([]*Update, error) {
	url := b.endpoint + "/bot" + b.token + "/getUpdates"
	b.l.Debug("🗳️ start polling...", "offset", offset, "timeout", timeout)

	o := map[string]any{
		"offset":  offset,
		"timeout": timeout,
	}

	var i struct {
		Result []*Update `json:"result"`
	}

	if err := b.do("POST", url, o, &i); err != nil {
		return nil, err
	}

	return i.Result, nil
}

func (b *Bot) SendMessage(chatID int64, text string, opts ...any) (*Message, error) {
	url := b.endpoint + "/bot" + b.token + "/sendMessage"
	req := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
		"link_preview_options": map[string]any{
			"is_disabled": true,
		},
	}
	for _, opt := range opts {
		switch o := opt.(type) {
		case map[string]any:
			maps.Copy(req, o)
		default:
			break
		}
	}

	var res struct {
		Result *Message `json:"result"`
	}

	if err := b.do("POST", url, req, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (b *Bot) EditMessageText(chatID, messageID int64, text string, opts ...any) (*Message, error) {
	url := b.endpoint + "/bot" + b.token + "/editMessageText"
	req := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
		"link_preview_options": map[string]any{
			"is_disabled": true,
		},
	}
	for _, opt := range opts {
		switch o := opt.(type) {
		case map[string]any:
			maps.Copy(req, o)
		default:
			break
		}
	}

	var res struct {
		Result *Message `json:"result"`
	}

	if err := b.do("POST", url, req, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (b *Bot) IsUserMemberOfChat(userID, chatID int64) bool {
	chatMember, err := b.GetChatMember(userID, chatID)
	if err != nil {
		b.l.Error("failed to get ChatMember", "error", err)
		return false
	}

	return chatMember != nil && chatMember.Status != "left" && chatMember.Status != "kicked"
}

func (b *Bot) GetChatMember(userID, chatID int64) (*ChatMember, error) {
	url := b.endpoint + "/bot" + b.token + "/getChatMember"
	o := map[string]any{
		"user_id": userID,
		"chat_id": chatID,
	}

	var i struct {
		Result *ChatMember `json:"result"`
	}

	if err := b.do("POST", url, o, &i); err != nil {
		return nil, err
	}
	return i.Result, nil
}

func (b *Bot) do(method, url string, o, i any) error {
	var body io.Reader
	if o != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(&o)
		if err != nil {
			return fmt.Errorf("failed to pack data %w", err)
		}
		body = io.NopCloser(buf)
	}

	request, err := http.NewRequestWithContext(b.ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if method == "POST" {
		request.Header.Add("Content-Type", "application/json")
	}

	resp, err := b.client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	if err = chkErr(data); err != nil {
		return err
	}

	return json.Unmarshal(data, i)
}
