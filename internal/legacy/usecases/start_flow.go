package usecases

import (
	"encoding/json"
	"strings"
	"sync"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateStartFlow(l core.Logger, loc legacy.ILocalizer, settings legacy.ISettingsProvider, commandService legacy.ICommandService) *StartFlow {
	return &StartFlow{l, loc, settings, commandService, sync.Mutex{}}
}

type StartFlow struct {
	l              core.Logger
	loc            legacy.ILocalizer
	settings       legacy.ISettingsProvider
	commandService legacy.ICommandService
	lock           sync.Mutex
}

func (flow *StartFlow) Start(message *legacy.Message, bot legacy.IBot) error {
	flow.lock.Lock()
	defer flow.lock.Unlock()

	if strings.HasPrefix(message.Text, "/start") {
		if len(message.Text) > 7 {
			payload := message.Text[7:]
			err := flow.handlePayload(payload, message.Chat.ID)
			if err != nil {
				flow.l.Error(err)
				//return err ?
			}

			if payload == "vpnhelp" {
				// do not show default greeting message in case of VPN workflow
				_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "start_vpn"))
				return err
			}
		}

		err := flow.commandService.EnableCommands(message.Chat.ID, []legacy.Command{{Text: "help", Description: "show help message"}}, bot)
		if err != nil {
			flow.l.Error(err)
			// return err ?
		}
		_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "start_welcome") + " " + flow.loc.I18n(message.Sender.LanguageCode, "help"))
		return err
	}

	return nil
}

func (flow *StartFlow) Help(message *legacy.Message, bot legacy.IBot) error {
	_, err := bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "help"))
	return err
}

func (flow *StartFlow) handlePayload(payload string, chatID int64) error {
	data, err := flow.settings.GetData(chatID, legacy.SPayloadList)

	if err != nil {
		flow.l.Error(err)
	}

	var settingsV1 struct {
		Payload []string
	}

	err = json.Unmarshal(data, &settingsV1)
	if err != nil {
		flow.l.Error(err)
		// TODO: perform a migration
	}

	if flow.contains(payload, settingsV1.Payload) {
		return nil
	}

	settingsV1.Payload = append(settingsV1.Payload, payload)
	data, err = json.Marshal(settingsV1)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	return flow.settings.SetData(chatID, legacy.SPayloadList, data)
}

func (flow *StartFlow) contains(payload string, current []string) bool {
	for _, p := range current {
		if p == payload {
			return true
		}
	}
	return false
}
