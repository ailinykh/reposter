package telegram

type Update struct {
	ID      int64    `json:"update_id"`
	Message *Message `json:"message,omitempty"`
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
	Chat     *Chat  `json:"chat"`
	Date     int    `json:"date"`
	From     *User  `json:"from,omitempty"`
	ID       int64  `json:"message_id"`
	Text     string `json:"text,omitempty"`
	Entities []struct {
		Offset int    `json:"offset"`
		Length int    `json:"length"`
		Type   string `json:"type"`
	}
}

func (m *Message) Commands() []string {
	var commands = []string{}
	for _, e := range m.Entities {
		if e.Type == "bot_command" {
			commands = append(commands, m.Text[e.Offset:e.Offset+e.Length])
		}
	}
	return commands
}

type ChatMember struct {
	Status string `json:"status"`
	User   *User  `json:"user"`
}
