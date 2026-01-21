package telegram

import "io"

// Update https://core.telegram.org/bots/api#update
type Update struct {
	ID            int64          `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// User https://core.telegram.org/bots/api#user
type User struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	IsBot        bool   `json:"is_bot"`
	IsPremium    bool   `json:"is_premium,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

func (u *User) DisplayName() string {
	if u.Username != "" {
		return u.Username
	}

	if u.LastName != "" {
		return u.FirstName
	}

	return u.FirstName + " " + u.LastName
}

// Chat https://core.telegram.org/bots/api#chat
type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title,omitempty"`
	Username  string `json:"username,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}

func (c *Chat) Private() bool {
	return c.Type == "private"
}

// Message https://core.telegram.org/bots/api#message
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
	Photo             []*PhotoSize    `json:"photo,omitempty"`
	Video             *Video          `json:"video,omitempty"`
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

// ParseMode https://core.telegram.org/bots/api#formatting-options
type ParseMode string

const (
	ParseModeMarkdown ParseMode = "MarkdownV2"
	ParseModeHTML     ParseMode = "HTML"
)

// MessageEntity https://core.telegram.org/bots/api#messageentity
type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

// PhotoSize https://core.telegram.org/bots/api#photosize
type PhotoSize struct {
	FileID       string `json:"file_id"`
	FileUniqueID string `json:"file_unique_id"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FileSize     int64  `json:"file_size"`
}

// Video https://core.telegram.org/bots/api#video
type Video struct {
	FileID         string       `json:"file_id"`
	FileUniqueID   string       `json:"file_unique_id"`
	Width          int          `json:"width"`
	Height         int          `json:"height"`
	Duration       int          `json:"duration"`
	Thumbnail      *PhotoSize   `json:"thumbnail,omitempty"`
	Cover          []*PhotoSize `json:"cover,omitempty"`
	StartTimestamp int          `json:"start_timestamp,omitempty"`
	FileName       string       `json:"file_name,omitempty"`
	MimeType       string       `json:"mime_type,omitempty"`
	FileSize       int64        `json:"file_size"`
}

// ReplyParameters https://core.telegram.org/bots/api#replyparameters
type ReplyParameters struct {
	MessageID int64  `json:"message_id"`
	Quote     string `json:"quote,omitempty"`
}

// LinkPreviewOptions https://core.telegram.org/bots/api#linkpreviewoptions
type LinkPreviewOptions struct {
	IsDisabled bool   `json:"is_disabled,omitempty"`
	URL        string `json:"url,omitempty"`
}

// CallbackQuery https://core.telegram.org/bots/api#callbackquery
type CallbackQuery struct {
	ID   string `json:"id"`
	From *User  `json:"from"`
	// NOTE: to ensure message is accessible, check it's date > 0
	MaybeInaccessibleMessage *Message `json:"message,omitempty"`

	ChatInstance string `json:"chat_instance"`
	Data         string `json:"data,omitempty"`
}

// ReplyMarkup
type ReplyMarkup interface {
	isReplyMarkup() // marker interface
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

func (InlineKeyboardMarkup) isReplyMarkup() {}

// ChatMember https://core.telegram.org/bots/api#chatmember
type ChatMember struct {
	Status string `json:"status"`
	User   *User  `json:"user"`
}

// InputFile https://core.telegram.org/bots/api#inputfile
type InputFile interface {
	isInputFile() // marker interface
}

type InputFileURL string

func (InputFileURL) isInputFile() {}

type InputFileLocal struct {
	Name   string
	Reader io.Reader
}

func (InputFileLocal) isInputFile() {}
