package main

import "github.com/ailinykh/reposter/v3/pkg/telegram"

type UpdateHandler interface {
	Handle(*telegram.Update, *telegram.Bot) error
}
