package telegram

import "maps"

type Message struct {
	ID                int64    `json:"message_id"`
	Date              int      `json:"date"`
	Chat              *Chat    `json:"chat"`
	From              *User    `json:"from"`
	ForwardFrom       *User    `json:"forward_from"`
	ForwardFromChat   *Chat    `json:"forward_from_chat"`
	ForwardSenderName string   `json:"forward_sender_name"`
	ReplyTo           *Message `json:"reply_to_message"`
	Text              string   `json:"text"`
	Entities          []struct {
		Offset int    `json:"offset"`
		Length int    `json:"length"`
		Type   string `json:"type"`
	}
}

func (m *Message) Commands() []string {
	return m.entities("bot_command")
}

func (m *Message) URLs() []string {
	return m.entities("url")
}

func (m *Message) entities(kind string) []string {
	var urls = []string{}
	runes := []rune(m.Text)
	for _, e := range m.Entities {
		if e.Type == kind {
			urls = append(urls, string(runes[e.Offset:e.Offset+e.Length]))
		}
	}
	return urls
}

func (b *Bot) SendMessage(chatID int64, text string, opts ...any) (*Message, error) {
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

	if err := b.do("sendMessage", req, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (b *Bot) EditMessageText(chatID, messageID int64, text string, opts ...any) (*Message, error) {
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

	if err := b.do("editMessageText", req, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (b *Bot) DeleteMessage(chatID, messageID int64) (bool, error) {
	o := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
	}
	var i struct {
		Result bool `json:"result"`
	}
	if err := b.do("deleteMessage", o, &i); err != nil {
		return false, err
	}
	return i.Result, nil
}
