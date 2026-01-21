package info

import (
	"context"
	"fmt"
	"strings"

	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

func New() *Info {
	return &Info{}
}

type Info struct{}

func (i *Info) Handle(ctx context.Context, u *telegram.Update, bot *telegram.Bot) error {
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

	if m.ReplyToMessage != nil {
		if m.ReplyToMessage.ForwardFromChat != nil {
			info = append(info,
				"💬 forward from chat",
				fmt.Sprintf("ID: <b>%d</b>", m.ReplyToMessage.ForwardFromChat.ID),
				fmt.Sprintf("Title: <b>%s</b>", m.ReplyToMessage.ForwardFromChat.Title),
				fmt.Sprintf("Type: <b>%s</b>", m.ReplyToMessage.ForwardFromChat.Type),
				"",
			)
		}
		if m.ReplyToMessage.ForwardFrom != nil {
			info = append(info,
				"👤 forward from",
				fmt.Sprintf("ID: <b>%d</b>", m.ReplyToMessage.ForwardFrom.ID),
				fmt.Sprintf("First: <b>%s</b>", m.ReplyToMessage.ForwardFrom.FirstName),
				fmt.Sprintf("Last: <b>%s</b>", m.ReplyToMessage.ForwardFrom.LastName),
				fmt.Sprintf("Username: <b>%s</b>", m.ReplyToMessage.ForwardFrom.Username),
				fmt.Sprintf("SenderName: <b>%s</b>", m.ReplyToMessage.ForwardSenderName),
				"",
			)
		}
	}

	_, err := bot.SendMessage(ctx, &telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      strings.Join(info, "\n"),
		ParseMode: telegram.ParseModeHTML,
	})
	return err
}
