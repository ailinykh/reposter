package telegram

type MessageEntity struct {
		Type   string `json:"type"`
		Offset int    `json:"offset"`
		Length int    `json:"length"`
	}

type Message struct {
	ID                int64           `json:"message_id"`
	From              *User           `json:"from,omitempty"`
	Date              int             `json:"date"`
	Chat              *Chat           `json:"chat"`
	ForwardFrom       *User           `json:"forward_from,omitempty"`
	ForwardFromChat   *Chat           `json:"forward_from_chat,omitempty"`
	ForwardSenderName string          `json:"forward_sender_name,omitempty"`
	ReplyToMessage    *Message        `json:"reply_to_message,omitempty"`
	Text              string          `json:"text,omitempty"`
	Entities          []MessageEntity `json:"entities,omitempty"`
	Caption           string          `json:"caption,omitempty"`
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

func (b *Bot) SendMessage(params SendMessageParams) (*Message, error) {
	var res struct {
		Result *Message `json:"result"`
	}

	if err := b.raw("sendMessage", params, &res); err != nil {
		return nil, err
	}

	return res.Result, nil
}

func (b *Bot) EditMessageText(params EditMessageTextParams) (*Message, error) {
	var res struct {
		Result *Message `json:"result"`
	}

	if err := b.raw("editMessageText", params, &res); err != nil {
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
	if err := b.raw("deleteMessage", o, &i); err != nil {
		return false, err
	}
	return i.Result, nil
}

// ParseMode https://core.telegram.org/bots/api#formatting-options
type ParseMode string

const (
	ParseModeMarkdown ParseMode = "MarkdownV2"
	ParseModeHTML     ParseMode = "HTML"
)
