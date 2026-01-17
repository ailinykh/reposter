package telegram

// GetUpdatesParams https://core.telegram.org/bots/api#getting-updates
type GetUpdatesParams struct {
	Offset         int64    `json:"offset,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

// SendPhotoParams https://core.telegram.org/bots/api#sendphoto
type SendPhotoParams struct {
	ChatID    int64     `json:"chat_id"`
	Photo     InputFile `json:"photo"`
	Caption   string    `json:"caption,omitempty"`
	ParseMode ParseMode `json:"parse_mode,omitempty"`
}

// SendVideoParams https://core.telegram.org/bots/api#sendvideo
type SendVideoParams struct {
	ChatID            int64     `json:"chat_id"`
	Video             InputFile `json:"video"`
	Duration          int       `json:"duration,omitempty"`
	Width             int       `json:"width,omitempty"`
	Height            int       `json:"height,omitempty"`
	Thumbnail         InputFile `json:"thumbnail,omitempty"`
	Caption           string    `json:"caption,omitempty"`
	ParseMode         ParseMode `json:"parse_mode,omitempty"`
	SupportsStreaming bool      `json:"supports_streaming,omitempty"`
}

// GetChatMemberParams https://core.telegram.org/bots/api#getchatmember
type GetChatMemberParams struct {
	ChatID int64 `json:"chat_id"`
	UserID int64 `json:"user_id"`
}
