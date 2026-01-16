package telegram

// GetUpdatesParams https://core.telegram.org/bots/api#getting-updates
type GetUpdatesParams struct {
	Offset         int64    `json:"offset,omitempty"`
	Limit          int      `json:"limit,omitempty"`
	Timeout        int      `json:"timeout,omitempty"`
	AllowedUpdates []string `json:"allowed_updates,omitempty"`
}

// GetChatMemberParams https://core.telegram.org/bots/api#getchatmember
type GetChatMemberParams struct {
	ChatID int64 `json:"chat_id"`
	UserID int64 `json:"user_id"`
}
