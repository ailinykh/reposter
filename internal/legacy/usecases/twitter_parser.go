package usecases

import (
	"fmt"
	"regexp"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateTwitterParser(l core.Logger, tweetHandler ITweetHandler) *TwitterParser {
	return &TwitterParser{l, tweetHandler}
}

type TwitterParser struct {
	l            core.Logger
	tweetHandler ITweetHandler
}

// HandleText is a core.ITextHandler protocol implementation
func (parser *TwitterParser) HandleText(message *legacy.Message, bot legacy.IBot) error {
	r := regexp.MustCompile(`https://(?i:twitter|x)\.com\S+/status/(\d+)`)
	match := r.FindAllStringSubmatch(message.Text, -1)

	if len(match) > 0 {
		parser.l.Info("processing", "url", match[0][0])
	} else {
		return fmt.Errorf("not implemented")
	}

	for _, m := range match {
		err := parser.tweetHandler.Process(m[1], message, bot)
		if err != nil {
			return err
		}
	}

	return nil
}
