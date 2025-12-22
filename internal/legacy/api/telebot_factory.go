package api

import (
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	tb "gopkg.in/telebot.v3"
)

func makeTbMessage(m *core.Message) *tb.Message {
	message := &tb.Message{
		ID:     m.ID,
		Chat:   &tb.Chat{ID: m.Chat.ID},
		Sender: makeTbUser(m.Sender),
	}
	if m.ReplyTo != nil {
		message.ReplyTo = makeTbMessage(m.ReplyTo)
	}
	if m.Video != nil {
		message.Video = makeTbVideo(m.Video, m.Text)
	}
	return message
}

func makeTbVideo(vf *core.Video, caption string) *tb.Video {
	var video *tb.Video
	if len(vf.ID) > 0 {
		video = &tb.Video{File: tb.File{FileID: vf.ID}}
		video.Caption = makeCaption(caption)
	} else {
		video = &tb.Video{File: tb.FromDisk(vf.Path)}
		video.FileName = vf.File.Name
		video.Width = vf.Width
		video.Height = vf.Height
		video.Caption = makeCaption(caption)
		video.Duration = vf.Duration
		video.Streaming = true
		video.Thumbnail = &tb.Photo{
			File:   tb.FromDisk(vf.Thumb.Path),
			Width:  vf.Thumb.Width,
			Height: vf.Thumb.Height,
		}
	}
	return video
}

func makeTbPhoto(image *core.Image, caption string) *tb.Photo {
	photo := &tb.Photo{File: tb.FromDisk(image.File.Path)}
	if len(image.ID) > 0 {
		photo = &tb.Photo{File: tb.File{FileID: image.ID}}
	}
	photo.Caption = makeCaption(caption)
	return photo
}

func makeTbUser(user *core.User) *tb.User {
	return &tb.User{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	}
}

func makeInlineKeyboard(k core.Keyboard) [][]tb.InlineButton {
	keyboard := [][]tb.InlineButton{}
	for _, buttons := range k {
		btns := []tb.InlineButton{}
		for _, b := range buttons {
			btn := tb.InlineButton{Unique: b.ID, Text: b.Text, Data: b.Payload}
			btns = append(btns, btn)
		}
		keyboard = append(keyboard, btns)
	}
	return keyboard
}

func makeTbCommands(commands []core.Command) []tb.Command {
	comands := []tb.Command{}
	for _, command := range commands {
		c := tb.Command{
			Text:        command.Text,
			Description: command.Description,
		}
		comands = append(comands, c)
	}
	return comands
}

func makeCaption(caption string) string {
	if len(caption) > 1024 {
		return caption[:1024]
	}
	return caption
}
