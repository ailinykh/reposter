package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func NewBot(opts ...func(*Bot)) (*Bot, error) {
	b := &Bot{
		client:   http.DefaultClient,
		ctx:      context.Background(),
		endpoint: "https://api.telegram.org",
		l:        slog.Default(),
	}

	for _, o := range opts {
		o(b)
	}

	me, err := b.GetMe()
	if err != nil {
		return nil, err
	}

	b.User = me
	b.l = b.l.With("username", me.Username)
	return b, nil
}

type Bot struct {
	*User
	client   *http.Client
	ctx      context.Context
	endpoint string
	token    string
	l        *slog.Logger
}

func (b *Bot) GetMe() (*User, error) {
	var i struct {
		Result *User `json:"result"`
	}

	if err := b.do("getMe", nil, &i); err != nil {
		return nil, err
	}

	return i.Result, nil
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

func (b *Bot) GetUpdates(params GetUpdatesParams) ([]*Update, error) {
	b.l.Debug("🗳️ start polling...", "offset", params.Offset, "timeout", params.Timeout)

	var i struct {
		Result []*Update `json:"result"`
	}

	if err := b.do("getUpdates", &params, &i); err != nil {
		return nil, err
	}

	return i.Result, nil
}

func (b *Bot) GetChatMember(params GetChatMemberParams) (*ChatMember, error) {
	var i struct {
		Result *ChatMember `json:"result"`
	}

	if err := b.do("getChatMember", params, &i); err != nil {
		return nil, err
	}
	return i.Result, nil
}

func (b *Bot) IsUserMemberOfChat(params GetChatMemberParams) bool {
	chatMember, err := b.GetChatMember(params)
	if err != nil {
		b.l.Error("failed to get ChatMember", "error", err)
		return false
	}

	return chatMember != nil && chatMember.Status != "left" && chatMember.Status != "kicked"
}

func (b *Bot) AnswerCallbackQuery(queryID, text string) error {
	o := map[string]any{
		"callback_query_id": queryID,
		"text":              text,
	}
	return b.do("answerCallbackQuery", o, nil)
}

func (b *Bot) SendPhoto(params SendPhotoParams) (*Message, error) {
	var i struct {
		Result *Message `json:"result"`
	}

	if err := b.do("sendPhoto", params, &i); err != nil {
		return nil, err
	}
	return i.Result, nil
}

func (b *Bot) SendVideo(chatID int64, url, caption string) (*Message, error) {
	o := map[string]any{
		"chat_id":    chatID,
		"video":      url,
		"caption":    caption,
		"parse_mode": "HTML",
	}

	var i struct {
		Result *Message `json:"result"`
	}

	if err := b.do("sendVideo", o, &i); err != nil {
		return nil, err
	}
	return i.Result, nil
}

func (b *Bot) do(method string, o, i any) error {
	url := b.endpoint + "/bot" + b.token + "/" + method
	var body io.Reader
	if o != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(&o)
		if err != nil {
			return fmt.Errorf("failed to pack data %w", err)
		}
		body = io.NopCloser(buf)
	}

	request, err := http.NewRequestWithContext(b.ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Add("Content-Type", "application/json")

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
