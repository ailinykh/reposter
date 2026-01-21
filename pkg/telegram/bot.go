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
	var rv []*Update
	err := b.raw("getUpdates", params, &rv)
	return rv, err
}

// GetMe https://core.telegram.org/bots/api#getme
func (b *Bot) GetMe() (*User, error) {
	var rv *User
	err := b.raw("getMe", nil, &rv)
	return rv, err
}

// SendMessage https://core.telegram.org/bots/api#sendmessage
func (b *Bot) SendMessage(params SendMessageParams) (*Message, error) {
	var rv *Message
	err := b.raw("sendMessage", params, &rv)
	return rv, err
}

// SendVideo https://core.telegram.org/bots/api#sendvideo
func (b *Bot) SendVideo(params SendVideoParams) (*Message, error) {
	var rv *Message
	if _, ok := params.Video.(InputFileLocal); ok {
		err := b.rawMultipart("sendVideo", params, &rv)
		return rv, err
	}
	err := b.raw("sendVideo", params, &rv)
	return rv, err
}

// GetChatMember https://core.telegram.org/bots/api#getchatmember
func (b *Bot) GetChatMember(params GetChatMemberParams) (*ChatMember, error) {
	var rv *ChatMember
	err := b.raw("getChatMember", params, &rv)
	return rv, err
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
	var rv *Message
	if _, ok := params.Photo.(InputFileLocal); ok {
		err := b.rawMultipart("sendPhoto", params, &rv)
		return rv, err
	}
	err := b.raw("sendPhoto", params, &rv)
	return rv, err
}

// EditMessageText https://core.telegram.org/bots/api#editmessagetext
func (b *Bot) EditMessageText(params EditMessageTextParams) (*Message, error) {
	var rv *Message
	err := b.raw("editMessageText", params, &rv)
	return rv, err
}

// DeleteMessage https://core.telegram.org/bots/api#deletemessage
func (b *Bot) DeleteMessage(params DeleteMessageParams) (bool, error) {
	var rv bool
	err := b.raw("deleteMessage", params, &rv)
	return rv, err
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

	var r apiResponse
	if err = json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if !r.OK {
		return fmt.Errorf("telegram error: %s", r.Description)
	}

	return json.Unmarshal(r.Result, in)
}

type apiResponse struct {
	OK          bool               `json:"ok"`
	Result      json.RawMessage    `json:"result,omitempty"`
	Description string             `json:"description,omitempty"`
	ErrorCode   int                `json:"error_code,omitempty"`
	Parameters  *apiResponseParams `json:"parameters,omitempty"`
}

type apiResponseParams struct {
	RetryAfter      int `json:"retry_after,omitempty"`
	MigrateToChatID int `json:"migrate_to_chat_id,omitempty"`
}
