package telegram

import "io"

type Update struct {
	ID            int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Chat struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Title     string `json:"title,omitempty"`
	Type      string `json:"type"`
	Username  string `json:"username,omitempty"`
}

func (c *Chat) Private() bool {
	return c.Type == "private"
}

type User struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	IsBot        bool   `json:"is_bot"`
	IsPremium    bool   `json:"is_premium,omitempty"`
	LanguageCode string `json:"language_code"`
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

type ChatMember struct {
	Status string `json:"status"`
	User   *User  `json:"user"`
}

type CallbackQuery struct {
	ID           string `json:"id"`
	ChatInstance string `json:"chat_instance"`
	Data         string `json:"data"`
	From         *User  `json:"from"`
	// NOTE: to ensure message is accessible, check it's date > 0
	MaybeInaccessibleMessage *Message `json:"message"`
}

type ParseMode string

const (
	ParseModeMarkdown ParseMode = "MarkdownV2"
	ParseModeHTML     ParseMode = "HTML"
)

// InputFile
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

// ReplyParameters
type ReplyParameters struct {
	MessageID int64  `json:"message_id"`
	Quote     string `json:"quote,omitempty"`
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

// LinkPreviewOptions
type LinkPreviewOptions struct {
	IsDisabled bool   `json:"is_disabled,omitempty"`
	URL        string `json:"url,omitempty"`
}
