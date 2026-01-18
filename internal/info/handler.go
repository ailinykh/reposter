package info

import (
	"fmt"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func New() *Info {
	return &Info{}
}

type Info struct{}

func (i *Info) Handle(u *telegram.Update, bot *telegram.Bot) error {
	m := u.Message
	if m == nil || len(m.Commands()) == 0 || m.Commands()[0] != "/info" {
		return nil
	}

	info := []string{
		"💬 Chat",
		fmt.Sprintf("ID: <b>%d</b>", m.Chat.ID),
		fmt.Sprintf("Title: <b>%s</b>", m.Chat.Title),
		fmt.Sprintf("Type: <b>%s</b>", m.Chat.Type),
		"",
		"👤 Sender",
		fmt.Sprintf("ID: <b>%d</b>", m.From.ID),
		fmt.Sprintf("First: <b>%s</b>", m.From.FirstName),
		fmt.Sprintf("Last: <b>%s</b>", m.From.LastName),
		fmt.Sprintf("Username: <b>%s</b>", m.From.Username),
		"",
	}

	if m.ReplyTo != nil {
		if m.ReplyTo.ForwardFromChat != nil {
			info = append(info,
				"💬 forward from chat",
				fmt.Sprintf("ID: <b>%d</b>", m.ReplyTo.ForwardFromChat.ID),
				fmt.Sprintf("Title: <b>%s</b>", m.ReplyTo.ForwardFromChat.Title),
				fmt.Sprintf("Type: <b>%s</b>", m.ReplyTo.ForwardFromChat.Type),
				"",
			)
		}
		if m.ReplyTo.ForwardFrom != nil {
			info = append(info,
				"👤 forward from",
				fmt.Sprintf("ID: <b>%d</b>", m.ReplyTo.ForwardFrom.ID),
				fmt.Sprintf("First: <b>%s</b>", m.ReplyTo.ForwardFrom.FirstName),
				fmt.Sprintf("Last: <b>%s</b>", m.ReplyTo.ForwardFrom.LastName),
				fmt.Sprintf("Username: <b>%s</b>", m.ReplyTo.ForwardFrom.Username),
				fmt.Sprintf("SenderName: <b>%s</b>", m.ReplyTo.ForwardSenderName),
				"",
			)
		}
	}

	_, err := bot.SendMessage(telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      strings.Join(info, "\n"),
		ParseMode: telegram.ParseModeHTML,
	})
	return err
}
