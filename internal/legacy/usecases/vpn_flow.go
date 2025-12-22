package usecases

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateVpnFlow(l core.Logger, loc legacy.ILocalizer, api legacy.IVpnAPI, settings legacy.ISettingsProvider) legacy.ITextHandler {
	flow := VpnFlow{l, loc, api, settings, make(map[string]func(*legacy.Message, legacy.IBot) error), make(map[int64]VpnState)}
	flow.callbacks["vpn_create_key"] = flow.create
	flow.callbacks["vpn_manage_key"] = flow.manage
	flow.callbacks["vpn_delete_key"] = flow.delete
	flow.callbacks["vpn_back"] = flow.back
	flow.callbacks["vpn_cancel"] = flow.cancel
	return &flow
}

type VpnFlow struct {
	l        core.Logger
	loc      legacy.ILocalizer
	api      legacy.IVpnAPI
	settings legacy.ISettingsProvider

	callbacks map[string]func(*legacy.Message, legacy.IBot) error
	state     map[int64]VpnState
}

type VpnState struct {
	action string
	source int
}

// GetButtonIds is a core.IButtonHandler protocol implementation
func (flow *VpnFlow) GetButtonIds() []string {
	keys := make([]string, len(flow.callbacks))

	i := 0
	for k := range flow.callbacks {
		keys[i] = k
		i++
	}

	return keys
}

// ButtonPressed is a core.IButtonHandler protocol implementation
func (flow *VpnFlow) ButtonPressed(button *legacy.Button, message *legacy.Message, _ *legacy.User, bot legacy.IBot) error {
	if callback, ok := flow.callbacks[button.ID]; ok {
		return callback(message, bot)
	}
	return fmt.Errorf("not implemented")
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *VpnFlow) HandleText(message *legacy.Message, bot legacy.IBot) error {
	if !message.IsPrivate {
		return fmt.Errorf("not implemented")
	}

	if state, ok := flow.state[message.Chat.ID]; ok {
		return flow.handleAction(state, message, bot)
	}

	if message.Text == "/vpnhelp" {
		if flow.hasAccess(message.Chat.ID) {
			return flow.help(message, bot)
		}
		_, err := bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_mislead"))
		return err
	}

	flow.l.Info(message.Text)

	return fmt.Errorf("not implemented")
}

func (flow *VpnFlow) hasAccess(chatID int64) bool {
	data, err := flow.settings.GetData(chatID, legacy.SPayloadList)
	if err != nil {
		flow.l.Error(err)
		return false
	}

	var settingsV1 struct {
		Payload []string
	}

	err = json.Unmarshal(data, &settingsV1)
	if err != nil {
		flow.l.Error(err)
		// TODO: perform a migration
		return false
	}

	for _, s := range settingsV1.Payload {
		if s == "vpnhelp" {
			return true
		}
	}
	return false
}

func (flow *VpnFlow) help(message *legacy.Message, bot legacy.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_welcome"), flow.getKeyboard(message, keys))
	return err
}

func (flow *VpnFlow) create(message *legacy.Message, bot legacy.IBot) error {
	flow.state[message.Chat.ID] = VpnState{"create", message.ID}
	keyboard := legacy.Keyboard{[]*legacy.Button{{ID: "vpn_back", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_back")}}}

	_, err := bot.Edit(message, flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_create_key_name"), keyboard)
	return err
}

func (flow *VpnFlow) manage(message *legacy.Message, bot legacy.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	text := []string{flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_list_top")}

	for idx, key := range keys {
		text = append(text, flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_list_item", idx+1, key.Title, key.Key))
	}

	text = append(text, flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_list_bottom", len(keys)))

	keyboard := legacy.Keyboard{
		[]*legacy.Button{{ID: "vpn_delete_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_remove_key")}},
		[]*legacy.Button{{ID: "vpn_back", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_back")}},
	}
	_, err = bot.Edit(message, strings.Join(text, "\n"), keyboard)
	return err
}

func (flow *VpnFlow) back(message *legacy.Message, bot legacy.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	delete(flow.state, message.Chat.ID)

	_, err = bot.Edit(message, flow.loc.I18n(message.Sender.LanguageCode, "vpn_welcome"), flow.getKeyboard(message, keys))
	return err
}

func (flow *VpnFlow) delete(message *legacy.Message, bot legacy.IBot) error {
	keys, err := flow.api.GetKeys(message.Chat.ID)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	flow.state[message.Chat.ID] = VpnState{"delete", message.ID}

	text := []string{flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_delete_key_name_top")}

	for _, key := range keys {
		text = append(text, flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_delete_key_name_item", key.Title))
	}

	keyboard := legacy.Keyboard{[]*legacy.Button{
		{ID: "vpn_cancel", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_cancel")},
	}}
	_, err = bot.Edit(message, strings.Join(text, "\n"), keyboard)
	return err
}

func (flow *VpnFlow) cancel(message *legacy.Message, bot legacy.IBot) error {
	return flow.back(message, bot)
}

func (flow *VpnFlow) getKeyboard(message *legacy.Message, keys []*legacy.VpnKey) legacy.Keyboard {
	keyboard := legacy.Keyboard{}

	if len(keys) < 10 {
		keyboard = append(keyboard, []*legacy.Button{{ID: "vpn_create_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_create_key")}})
	}

	if len(keys) > 0 {
		keyboard = append(keyboard, []*legacy.Button{{ID: "vpn_manage_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_manage_key")}})
	}

	return keyboard
}

func (flow *VpnFlow) handleAction(state VpnState, message *legacy.Message, bot legacy.IBot) error {
	if state.action == "create" {
		if len(message.Text) > 64 {
			_, err := bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_enter_create_key_name_too_long"))
			return err
		}
		key, err := flow.api.CreateKey(message.Text, message.Chat.ID, message.Sender)
		if err != nil {
			flow.l.Error(err)
			return err
		}

		delete(flow.state, message.Chat.ID)

		_ = bot.Delete(&legacy.Message{ID: state.source, Chat: message.Chat})

		keyboard := legacy.Keyboard{[]*legacy.Button{{ID: "vpn_manage_key", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_manage_key")}}}
		_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_created", key.Key), keyboard)
		return err
	}

	if state.action == "delete" {
		keys, err := flow.api.GetKeys(message.Chat.ID)
		if err != nil {
			flow.l.Error(err)
			return err
		}

		delete(flow.state, message.Chat.ID)

		_ = bot.Delete(&legacy.Message{ID: state.source, Chat: message.Chat})

		keyboard := legacy.Keyboard{
			[]*legacy.Button{{ID: "vpn_back", Text: flow.loc.I18n(message.Sender.LanguageCode, "vpn_button_back")}},
		}

		for _, k := range keys {
			if k.Title == message.Text {
				err = flow.api.DeleteKey(k)
				if err != nil {
					return err
				}
				_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_deleted", k.Title), keyboard)
				return err
			}
		}
		_, err = bot.SendText(flow.loc.I18n(message.Sender.LanguageCode, "vpn_key_not_found"), keyboard)
		return err
	}

	return fmt.Errorf("unexpected action: %s", state.action)
}
