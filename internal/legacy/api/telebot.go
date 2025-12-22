package api

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/helpers"
	tb "gopkg.in/telebot.v3"
)

// Telebot is a telegram API
type Telebot struct {
	config           *TelebotConfig
	bot              *tb.Bot
	logger           core.Logger
	coreFactory      *CoreFactory
	multipart        *helpers.SendMultipartVideo
	commandHandlers  []string
	textHandlers     []legacy.ITextHandler
	documentHandlers []legacy.IDocumentHandler
	imageHandlers    []legacy.IImageHandler
	videoHandlers    []legacy.IVideoHandler
}

// CreateTelebot is a default Telebot factory
func CreateTelebot(logger core.Logger, opts ...TelebotConfigOption) *Telebot {
	config := TelebotConfig{}

	for _, opt := range opts {
		opt(&config)
	}

	poller := tb.NewMiddlewarePoller(&tb.LongPoller{Timeout: 10 * time.Second}, func(upd *tb.Update) bool {
		return true
	})

	var err error
	bot, err := tb.NewBot(tb.Settings{
		Token:  config.BotToken,
		Poller: poller,
		Client: config.Client,
	})

	if err != nil {
		panic(err)
	}

	var multipart *helpers.SendMultipartVideo
	if len(config.BotAPIUrl) > 0 {
		apiURL := fmt.Sprintf("%s/bot%s/sendVideo", config.BotAPIUrl, config.BotToken)
		multipart = helpers.CreateSendMultipartVideo(logger, apiURL)
	}

	telebot := &Telebot{
		&config,
		bot,
		logger,
		&CoreFactory{},
		multipart,
		[]string{},
		[]legacy.ITextHandler{},
		[]legacy.IDocumentHandler{},
		[]legacy.IImageHandler{},
		[]legacy.IVideoHandler{},
	}

	bot.Handle(tb.OnText, func(c tb.Context) error {
		var err error
		var message = telebot.coreFactory.makeMessage(c.Message())
		var bot = telebot.coreFactory.makeIBot(c.Message(), telebot)
		for _, h := range telebot.textHandlers {
			err = h.HandleText(message, bot)
			if err != nil {
				if err.Error() == "not implemented" {
					err = nil // skip "not implemented" error
				} else {
					logger.Error(fmt.Sprintf("%T: %s", h, err))
					telebot.reportError(c.Message(), err)
				}
			}
		}
		return err
	})

	var mutex sync.Mutex

	bot.Handle(tb.OnDocument, func(c tb.Context) error {
		var err error
		var m = c.Message()
		// TODO: inject `download` to get rid of MIME cheking
		if m.Document.MIME[:5] == "video" || m.Document.MIME == "image/gif" {
			mutex.Lock()
			defer mutex.Unlock()

			logger.Info("attempt to download document", "file_name", m.Document.FileName, "mime", m.Document.MIME, "username", m.Sender.Username)

			path := path.Join(os.TempDir(), m.Document.FileName)
			err := bot.Download(&m.Document.File, path)
			if err != nil {
				logger.Error(err)
				return err
			}

			logger.Info("document downloaded", "path", strings.ReplaceAll(path, os.TempDir(), "$TMPDIR/"))
			defer os.Remove(path)

			for _, h := range telebot.documentHandlers {
				err = h.HandleDocument(&legacy.Document{
					File: legacy.File{Name: m.Document.FileName, Path: path},
					MIME: m.Document.MIME,
				}, telebot.coreFactory.makeMessage(m), telebot.coreFactory.makeIBot(m, telebot))
				if err != nil {
					logger.Error("%T: %s", h, err)
					telebot.reportError(m, err)
				}
			}
		}
		return err
	})

	bot.Handle(tb.OnPhoto, func(c tb.Context) error {
		var err error
		var m = c.Message()
		image := &legacy.Image{
			File: legacy.File{
				Name: m.Photo.File.UniqueID + ".jpg",
				Path: m.Photo.File.FilePath,
				Size: m.Photo.File.FileSize,
			},
			ID:      m.Photo.FileID,
			FileURL: m.Photo.FileURL,
			Width:   m.Photo.Width,
			Height:  m.Photo.Height,
		}

		for _, h := range telebot.imageHandlers {
			err = h.HandleImage(image, telebot.coreFactory.makeMessage(m), telebot.coreFactory.makeIBot(m, telebot))
			if err != nil {
				logger.Error("%T: %s", h, err)
				telebot.reportError(m, err)
			}
		}
		return err
	})

	bot.Handle(tb.OnVideo, func(c tb.Context) error {
		var err error
		var m = c.Message()
		video := &legacy.Video{
			ID:     m.Video.FileID,
			Width:  m.Video.Width,
			Height: m.Video.Height,
		}
		logger.Info("handle video", "message", m, "video", video)
		for _, h := range telebot.videoHandlers {
			err = h.HandleVideo(video, telebot.coreFactory.makeMessage(m), telebot.coreFactory.makeIBot(m, telebot))
			if err != nil {
				logger.Error(fmt.Sprintf("%T: %s", h, err))
				telebot.reportError(m, err)
			}
		}
		return err
	})

	return telebot
}

// Download is a core.IImageDownloader interface implementation
func (t *Telebot) Download(image *legacy.Image) (*os.File, error) {
	//TODO: potential race condition
	file := tb.FromURL(image.FileURL)
	file.FileID = image.ID
	path := path.Join(os.TempDir(), image.Name)
	err := t.bot.Download(&file, path)
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}

	osFile, err := os.Open(path)
	if err != nil {
		t.logger.Error(err)
		return nil, err
	}

	t.logger.Info("image downloaded", "file_id", file.UniqueID, "file_path", path, "size", image.File.Size)

	return osFile, nil
}

// AddHandler register object as one of core.Handler's
func (t *Telebot) AddHandler(handler ...interface{}) {
	switch h := handler[0].(type) {
	case legacy.IDocumentHandler:
		t.documentHandlers = append(t.documentHandlers, h)
	case legacy.ITextHandler:
		t.textHandlers = append(t.textHandlers, h)
	case legacy.IImageHandler:
		t.imageHandlers = append(t.imageHandlers, h)
	case legacy.IVideoHandler:
		t.videoHandlers = append(t.videoHandlers, h)
	case string:
		t.registerCommand(h)
		if f, ok := handler[1].(func(*legacy.Message, legacy.IBot) error); ok {
			t.bot.Handle(h, func(c tb.Context) error {
				m := c.Message()
				return f(t.coreFactory.makeMessage(m), &TelebotAdapter{m, t})
			})
		} else {
			panic("interface must implement func(*core.Message, core.IBot) error")
		}
	default:
		panic(fmt.Sprintf("something wrong with %s", h))
	}

	if h, ok := handler[0].(legacy.IButtonHandler); ok {
		for _, id := range h.GetButtonIds() {
			t.bot.Handle("\f"+id, func(c tb.Context) error {
				m := c.Message()
				cb := c.Callback()
				button := legacy.Button{
					ID:      cb.Unique,
					Text:    c.Text(),
					Payload: c.Data(),
				}
				err := h.ButtonPressed(
					&button,
					t.coreFactory.makeMessage(m),
					t.coreFactory.makeUser(c.Sender()),
					t.coreFactory.makeIBot(m, t),
				)
				if err != nil {
					t.logger.Error(err)
					t.reportError(m, err)
					resp := tb.CallbackResponse{
						CallbackID: cb.ID,
						Text:       err.Error(),
					}
					return t.bot.Respond(cb, &resp)
				}
				return t.bot.Respond(cb, &tb.CallbackResponse{CallbackID: cb.ID})
			})
		}
	}
}

// Run bot loop
func (t *Telebot) Run() {
	t.bot.Start()
}

func (t *Telebot) Stop() {
	t.bot.Stop()
}

func (t *Telebot) registerCommand(command string) {
	for _, c := range t.commandHandlers {
		if c == command {
			panic("Handler for " + command + " already set!")
		}
	}
	t.commandHandlers = append(t.commandHandlers, command)
}

func (t *Telebot) reportError(m *tb.Message, e error) {
	if t.config.ReportChatId == 0 {
		return
	}
	chat := &tb.Chat{ID: t.config.ReportChatId}
	opts := &tb.SendOptions{DisableWebPagePreview: true}
	_, err := t.bot.Forward(chat, m, opts)
	if err != nil {
		t.logger.Error("forward message failed", "message", m, "error", err)
	}

	_, err = t.bot.Send(chat, e.Error(), opts)
	if err != nil {
		t.logger.Error("send message failed", "error", err)
	}
}

type CoreFactory struct {
}

func (factory *CoreFactory) makeMessage(m *tb.Message) *legacy.Message {
	text := m.Text
	if m.Document != nil {
		text = m.Caption
	}
	message := &legacy.Message{
		ID:        m.ID,
		Chat:      factory.makeChat(m.Chat),
		IsPrivate: m.Private(),
		Sender:    factory.makeUser(m.Sender),
		Text:      text,
	}

	if m.ReplyTo != nil {
		message.ReplyTo = factory.makeMessage(m.ReplyTo)
	}

	if m.Video != nil {
		message.Video = factory.makeVideo(m.Video)
	}

	return message
}

func (factory *CoreFactory) makeChat(c *tb.Chat) *legacy.Chat {
	title := c.Title
	if c.Type == tb.ChatPrivate {
		title = c.FirstName + " " + c.LastName
	}
	return &legacy.Chat{
		ID:    c.ID,
		Title: title,
		Type:  string(c.Type),
	}
}

func (CoreFactory) makeUser(u *tb.User) *legacy.User {
	return &legacy.User{
		ID:           u.ID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     u.Username,
		LanguageCode: u.LanguageCode,
	}
}

func (factory *CoreFactory) makeVideo(v *tb.Video) *legacy.Video {
	return &legacy.Video{
		File: legacy.File{
			Name: v.FileName,
			Path: v.FileURL,
		},
		ID:       v.FileID,
		Width:    v.Width,
		Height:   v.Height,
		Bitrate:  0,
		Duration: v.Duration,
		Codec:    "",
		Thumb:    factory.makePhoto(v.Thumbnail),
	}
}

func (CoreFactory) makePhoto(p *tb.Photo) *legacy.Image {
	return &legacy.Image{
		File: legacy.File{
			Name: p.FileLocal,
			Path: p.FilePath,
		},
		ID:      p.FileID,
		FileURL: p.FileURL,
		Width:   p.Width,
		Height:  p.Height,
	}
}

func (CoreFactory) makeCommands(commands []tb.Command) []legacy.Command {
	comands := []legacy.Command{}
	for _, command := range commands {
		c := legacy.Command{
			Text:        command.Text,
			Description: command.Description,
		}
		comands = append(comands, c)
	}
	return comands
}

func (CoreFactory) makeFile(name string, path string) *legacy.File {
	return &legacy.File{
		Name: name,
		Path: path,
	}
}

func (CoreFactory) makeIBot(m *tb.Message, t *Telebot) legacy.IBot {
	return &TelebotAdapter{m, t}
}
