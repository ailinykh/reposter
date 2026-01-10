package telegram

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
	if len(u.Username) > 0 {
		return u.Username
	}

	if len(u.LastName) == 0 {
		return u.FirstName
	}

	return u.FirstName + " " + u.LastName
}

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
	for _, e := range m.Entities {
		if e.Type == kind {
			urls = append(urls, m.Text[e.Offset:e.Offset+e.Length])
		}
	}
	return urls
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
