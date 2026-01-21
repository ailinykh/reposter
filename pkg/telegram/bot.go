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

// GetUpdates https://core.telegram.org/bots/api#getupdates
func (b *Bot) GetUpdates(params GetUpdatesParams) ([]*Update, error) {
	b.l.Debug("🗳️ start polling...", "offset", params.Offset, "timeout", params.Timeout)

	var i struct {
		Result []*Update `json:"result"`
	}

	if err := b.raw("getUpdates", &params, &i); err != nil {
		return nil, err
	}

	return i.Result, nil
}

// GetMe https://core.telegram.org/bots/api#getme
func (b *Bot) GetMe() (*User, error) {
	var i struct {
		Result *User `json:"result"`
	}

	if err := b.raw("getMe", nil, &i); err != nil {
		return nil, err
	}

	return i.Result, nil
}

// SendMessage https://core.telegram.org/bots/api#sendmessage
func (b *Bot) SendMessage(params SendMessageParams) (*Message, error) {
	var res struct {
		Result *Message `json:"result"`
	}

	if err := b.raw("sendMessage", params, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

// SendVideo https://core.telegram.org/bots/api#sendvideo
func (b *Bot) SendVideo(params SendVideoParams) (*Message, error) {
	var i struct {
		Result *Message `json:"result"`
	}

	if _, ok := params.Video.(InputFileLocal); ok {
		if err := b.rawMultipart("sendVideo", params, &i); err != nil {
			return nil, err
		}
		return i.Result, nil
	}

	if err := b.raw("sendVideo", params, &i); err != nil {
		return nil, err
	}
	return i.Result, nil
}

// GetChatMember https://core.telegram.org/bots/api#getchatmember
func (b *Bot) GetChatMember(params GetChatMemberParams) (*ChatMember, error) {
	var i struct {
		Result *ChatMember `json:"result"`
	}

	if err := b.raw("getChatMember", params, &i); err != nil {
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

// AnswerCallbackQuery https://core.telegram.org/bots/api#answercallbackquery
func (b *Bot) AnswerCallbackQuery(queryID, text string) error {
	o := map[string]any{
		"callback_query_id": queryID,
		"text":              text,
	}
	return b.raw("answerCallbackQuery", o, nil)
}

func (b *Bot) SendPhoto(params SendPhotoParams) (*Message, error) {
	var i struct {
		Result *Message `json:"result"`
	}

	if _, ok := params.Photo.(InputFileLocal); ok {
		if err := b.rawMultipart("sendPhoto", params, &i); err != nil {
			return nil, err
		}
		return i.Result, nil
	}

	if err := b.raw("sendPhoto", params, &i); err != nil {
		return nil, err
	}
	return i.Result, nil
}

// EditMessageText https://core.telegram.org/bots/api#editmessagetext
func (b *Bot) EditMessageText(params EditMessageTextParams) (*Message, error) {
	var res struct {
		Result *Message `json:"result"`
	}

	if err := b.raw("editMessageText", params, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

// DeleteMessage https://core.telegram.org/bots/api#deletemessage
func (b *Bot) DeleteMessage(params DeleteMessageParams) (bool, error) {
	var i struct {
		Result bool `json:"result"`
	}
	if err := b.raw("deleteMessage", params, &i); err != nil {
		return false, err
	}
	return i.Result, nil
}

func (b *Bot) raw(method string, out, in any) error {
	url := b.endpoint + "/bot" + b.token + "/" + method
	var body io.Reader
	if out != nil {
		buf := new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(&out)
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

	return json.Unmarshal(data, in)
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
