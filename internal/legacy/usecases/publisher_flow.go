package usecases

import (
	"fmt"
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreatePublisherFlow is a basic PublisherFlow factory
func CreatePublisherFlow(chatID legacy.ChatID, username string, l core.Logger) *PublisherFlow {
	publisher := PublisherFlow{
		l:           l,
		chatID:      chatID,
		username:    username,
		imageChan:   make(chan imgSource),
		requestChan: make(chan msgSource),
	}

	go publisher.runLoop()
	return &publisher
}

// PublisherFlow represents last sent image keeper logic
type PublisherFlow struct {
	l core.Logger

	chatID      int64
	username    string
	imageChan   chan imgSource
	requestChan chan msgSource
}

type imgSource struct {
	imageID string
	bot     legacy.IBot
}

type msgSource struct {
	message legacy.Message
	bot     legacy.IBot
}

// HandleImage is a core.IImageHandler protocol implementation
func (p *PublisherFlow) HandleImage(image *legacy.Image, message *legacy.Message, bot legacy.IBot) error {
	if message.Chat.ID == p.chatID && message.Sender.Username == p.username {
		p.imageChan <- imgSource{image.ID, bot}
	}

	return nil
}

func (p *PublisherFlow) HandleRequest(message *legacy.Message, bot legacy.IBot) error {
	if message.Chat.ID == p.chatID {
		p.requestChan <- msgSource{*message, bot}
	}

	return nil
}

func (p *PublisherFlow) runLoop() {
	photos := []string{}
	queue := []string{}

	disposal := func(m legacy.Message, bot legacy.IBot, timeout int) {
		time.Sleep(time.Duration(timeout) * time.Second)
		p.l.Info("disposing message", "message_id", m.ID, "chat_id", m.Chat.ID)
		err := bot.Delete(&m)
		if err != nil {
			p.l.Error(err)
		}
	}

	for {
		select {
		case is := <-p.imageChan:
			p.l.Info("got photo", "iamge_id", is.imageID)
			queue = append(queue, is.imageID)

		case <-time.After(1 * time.Second):
			if len(queue) > 0 {
				p.l.Info(fmt.Sprintf("had %d actual photo(s)", len(queue)))
				photos = queue
				queue = []string{}
			}

		case ms := <-p.requestChan:
			go disposal(ms.message, ms.bot, 0)
			switch count := len(photos); count {
			case 0:
				_, err := ms.bot.SendText("I have nothing for you comrade major")
				if err != nil {
					p.l.Error(err)
				}
			case 1:
				p.l.Info("have one actual photo")
				sent, err := ms.bot.SendImage(&legacy.Image{ID: photos[0]}, "")
				if err != nil {
					p.l.Error(err)
				} else {
					go disposal(*sent, ms.bot, 30)
				}
			default:
				p.l.Info("have many actual photos", "count", count)
				album := []*legacy.Image{}
				for _, p := range photos {
					album = append(album, &legacy.Image{ID: p})
				}
				sent, err := ms.bot.SendAlbum(album)
				if err != nil {
					p.l.Error(err)
				} else {
					for _, m := range sent {
						go disposal(*m, ms.bot, 30)
					}
				}
			}
		}
	}
}
