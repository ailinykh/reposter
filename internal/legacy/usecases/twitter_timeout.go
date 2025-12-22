package usecases

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateTwitterFlow is a basic TwitterFlow factory
func CreateTwitterTimeout(l core.Logger, tweetHandler ITweetHandler) *TwitterTimeout {
	return &TwitterTimeout{l, tweetHandler, make(map[legacy.Message]legacy.Message)}
}

// TwitterTimeout is a decorator for TwitterFlow to handle API timeouts gracefully
type TwitterTimeout struct {
	l            core.Logger
	tweetHandler ITweetHandler
	replies      map[legacy.Message]legacy.Message
}

// Process is a ITweetHandler protocol implementation
func (twitterTimeout *TwitterTimeout) Process(tweetID string, message *legacy.Message, bot legacy.IBot) error {
	err := twitterTimeout.tweetHandler.Process(tweetID, message, bot)
	if err != nil {
		if strings.HasPrefix(err.Error(), "Rate limit exceeded") {
			timeout, err := twitterTimeout.parseTimeout(err)
			if err != nil {
				return err
			}

			go func() {
				time.Sleep(time.Duration(timeout) * time.Second)
				_ = twitterTimeout.Process(tweetID, message, bot)
			}()

			minutes := timeout / 60
			seconds := timeout % 60
			reply := fmt.Sprintf("twitter api timeout %d min %d sec", minutes, seconds)
			sent, err := bot.SendText(reply, message)
			if err != nil {
				return err
			}
			twitterTimeout.replies[*message] = *sent
			//TODO: delay original message removing somehow
			return nil
		}
		twitterTimeout.l.Error(err)
	} else if sent, ok := twitterTimeout.replies[*message]; ok {
		_ = bot.Delete(&sent)
		delete(twitterTimeout.replies, *message)
	}
	return err
}

func (twitterTimeout *TwitterTimeout) parseTimeout(err error) (int64, error) {
	r := regexp.MustCompile(`(\-?\d+)$`)
	match := r.FindStringSubmatch(err.Error())
	if len(match) < 2 {
		return 0, fmt.Errorf("rate limit not found")
	}

	limit, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, err
	}

	timeout := limit - time.Now().Unix()
	twitterTimeout.l.Info("twitter api timeout", "seconds", timeout)
	timeout = int64(math.Max(float64(timeout), 2)) // Twitter api timeout might be negative
	return timeout, nil
}
