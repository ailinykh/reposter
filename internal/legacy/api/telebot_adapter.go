package api

import (
	"encoding/json"
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	tb "gopkg.in/telebot.v3"
)

// TelebotAdapter combines Telebot and core.IBot
type TelebotAdapter struct {
	m *tb.Message
	t *Telebot
}

// SendText is a core.IBot interface implementation
func (a *TelebotAdapter) SendText(text string, params ...interface{}) (*core.Message, error) {
	opts := tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true}
	for _, param := range params {
		switch m := param.(type) {
		case *core.Message:
			opts.ReplyTo = &tb.Message{ID: m.ID}
		case bool:
			opts.DisableWebPagePreview = m
		case core.Keyboard:
			opts.ReplyMarkup = &tb.ReplyMarkup{InlineKeyboard: makeInlineKeyboard(m)}
		default:
			break
		}
	}
	sent, err := a.t.bot.Send(a.m.Chat, text, &opts)
	if err != nil {
		return nil, err
	}
	return a.t.coreFactory.makeMessage(sent), err
}

// Delete is a core.IBot interface implementation
func (a *TelebotAdapter) Delete(message *core.Message) error {
	return a.t.bot.Delete(&tb.Message{ID: message.ID, Chat: &tb.Chat{ID: message.Chat.ID}})
}

// Edit is a core.IBot interface implementation
func (a *TelebotAdapter) Edit(message *core.Message, what interface{}, options ...interface{}) (*core.Message, error) {
	switch v := what.(type) {
	case core.Keyboard:
		replyMarkup := &tb.ReplyMarkup{InlineKeyboard: makeInlineKeyboard(v)}
		m, err := a.t.bot.EditReplyMarkup(makeTbMessage(message), replyMarkup)
		if err != nil {
			return nil, err
		}
		return a.t.coreFactory.makeMessage(m), nil
	case string:
		opts := &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true}
		for _, opt := range options {
			switch o := opt.(type) {
			case core.Keyboard:
				opts.ReplyMarkup = &tb.ReplyMarkup{InlineKeyboard: makeInlineKeyboard(o)}
			default:
				break
			}
		}
		m, err := a.t.bot.Edit(makeTbMessage(message), v, opts)
		if err != nil {
			return nil, err
		}
		return a.t.coreFactory.makeMessage(m), nil
	default:
	}
	return nil, fmt.Errorf("not implemented")
}

// SendImage is a core.IBot interface implementation
func (a *TelebotAdapter) SendImage(image *core.Image, caption string) (*core.Message, error) {
	photo := makeTbPhoto(image, caption)
	sent, err := photo.Send(a.t.bot, a.m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err != nil {
		return nil, err
	}
	return a.t.coreFactory.makeMessage(sent), err
}

// SendAlbum is a core.IBot interface implementation
func (a *TelebotAdapter) SendAlbum(images []*core.Image) ([]*core.Message, error) {
	album := tb.Album{}
	for _, i := range images {
		photo := &tb.Photo{File: tb.File{FileID: i.ID}}
		album = append(album, photo)
	}

	sent, err := a.t.bot.SendAlbum(a.m.Chat, album)
	if err != nil {
		return nil, err
	}

	var messages []*core.Message
	for _, m := range sent {
		messages = append(messages, a.t.coreFactory.makeMessage(&m))
	}
	return messages, err
}

// SendMedia is a core.IBot interface implementation
func (a *TelebotAdapter) SendMedia(media *core.Media) (*core.Message, error) {
	var sent *tb.Message
	var err error
	opts := &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true}
	switch media.Type {
	case core.TPhoto:
		a.t.logger.Info("sending media as photo", "media", media)
		file := &tb.Photo{File: tb.FromURL(media.ResourceURL)}
		file.Caption = makeCaption(media.Caption)
		a.t.bot.Notify(a.m.Chat, tb.UploadingPhoto)
		sent, err = a.t.bot.Send(a.m.Chat, file, opts)
	case core.TVideo:
		a.t.logger.Info("sending media as video", "media", media)
		file := &tb.Video{File: tb.FromURL(media.ResourceURL)}
		file.Caption = makeCaption(media.Caption)
		a.t.bot.Notify(a.m.Chat, tb.UploadingVideo)
		sent, err = a.t.bot.Send(a.m.Chat, file, opts)
	case core.TText:
		a.t.logger.Info("sending media as text", "media", media)
		sent, err = a.t.bot.Send(a.m.Chat, makeCaption(media.Caption), opts)
	}

	if err != nil {
		return nil, err
	}
	return a.t.coreFactory.makeMessage(sent), err
}

// SendMediaAlbum is a core.IBot interface implementation
func (a *TelebotAdapter) SendMediaAlbum(medias []*core.Media) ([]*core.Message, error) {
	var album = tb.Album{}
	opts := &tb.SendOptions{ParseMode: tb.ModeHTML, DisableWebPagePreview: true}

	for i, m := range medias {
		switch m.Type {
		case core.TVideo:
			video := &tb.Video{File: tb.FromURL(m.ResourceURL)}
			if i == len(medias)-1 {
				video.Caption = m.Caption
			}
			album = append(album, video)
		case core.TPhoto:
			photo := &tb.Photo{File: tb.FromURL(m.ResourceURL)}
			if i == len(medias)-1 {
				photo.Caption = m.Caption
			}
			album = append(album, photo)
		case core.TText, core.TAudio:
			a.t.logger.Error("%+v not supported", m.Type)
		}
	}

	sent, err := a.t.bot.SendAlbum(a.m.Chat, album, opts)
	if err != nil {
		return nil, err
	}

	var messages []*core.Message
	for _, m := range sent {
		messages = append(messages, a.t.coreFactory.makeMessage(&m))
	}
	return messages, err
}

// SendVideo is a core.IBot interface implementation
func (a *TelebotAdapter) SendVideo(vf *core.Video, caption string) (*core.Message, error) {
	if vf.Size > 50*1024*1024 && a.t.multipart != nil {
		body, err := a.t.multipart.SendVideo(vf, caption, a.m.Chat.ID)
		if err != nil {
			return nil, err
		}
		var resp struct {
			Result *tb.Message
		}
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return nil, err
		}
		return a.t.coreFactory.makeMessage(resp.Result), err
	}
	a.t.logger.Info("uploading video", "file_name", vf.Name, "size", fmt.Sprintf("(%.2f MB)", float64(vf.Size)/1024/1024))
	video := makeTbVideo(vf, caption)
	a.t.bot.Notify(a.m.Chat, tb.UploadingVideo)
	sent, err := video.Send(a.t.bot, a.m.Chat, &tb.SendOptions{ParseMode: tb.ModeHTML})
	if err != nil {
		return nil, err
	}
	a.t.logger.Info("video successfully sent", "file_name", vf.Name)
	return a.t.coreFactory.makeMessage(sent), err
}

// IsUserMemberOfChat is a core.IBot interface implementation
func (a *TelebotAdapter) IsUserMemberOfChat(user *core.User, chatID int64) bool {
	chat := &tb.Chat{ID: chatID}
	member, err := a.t.bot.ChatMemberOf(chat, makeTbUser(user))
	if err != nil {
		a.t.logger.Error(err, member)
	}
	return member != nil &&
		member.Role != tb.Left &&
		member.Role != tb.Kicked
}

// GetCommands is a core.IBot interface implementation
func (a *TelebotAdapter) GetCommands(chatID int64) ([]core.Command, error) {
	scope := tb.CommandScope{
		Type:   tb.CommandScopeChat,
		ChatID: chatID,
	}
	commands, err := a.t.bot.Commands(scope)
	if err != nil {
		return nil, err
	}
	return a.t.coreFactory.makeCommands(commands), nil
}

// SetCommands is a core.IBot interface implementation
func (a *TelebotAdapter) SetCommands(chatID int64, commands []core.Command) error {
	scope := tb.CommandScope{
		Type:   tb.CommandScopeChat,
		ChatID: chatID,
	}
	return a.t.bot.SetCommands(makeTbCommands(commands), scope)
}
