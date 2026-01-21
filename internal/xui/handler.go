package xui

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/helpers"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

type SettingsRepository interface {
	GetSettings(ctx context.Context, arg repository.GetSettingsParams) (repository.ChatSetting, error)
	SetSettings(ctx context.Context, arg repository.SetSettingsParams) (repository.ChatSetting, error)
}

func NewHandler(client *Client, l *slog.Logger, repo SettingsRepository) *Handler {
	return &Handler{
		client: client,
		l:      l,
		repo:   repo,
		state:  helpers.NewSafeMap[int64, string](),
	}
}

type Handler struct {
	client *Client
	l      *slog.Logger
	repo   SettingsRepository
	state  *helpers.SafeMap[int64, string]
}

func (h *Handler) Handle(u *telegram.Update, bot *telegram.Bot) error {
	// should answer
	if u.CallbackQuery != nil && strings.HasPrefix(u.CallbackQuery.Data, "vpn_") {
		return h.handleCallback(u.CallbackQuery, bot)
	}

	// only for private chats
	if u.Message == nil || !u.Message.Chat.Private() {
		return nil
	}

	// check if key name expected
	if state, ok := h.state.Get(u.Message.Chat.ID); ok {
		// consider to use another delimeter maybe
		parts := strings.Split(state, "_")
		if messageID, err := strconv.ParseInt(parts[len(parts)-1], 10, 64); err == nil {
			state = strings.Join(parts[:len(parts)-1], "_")

			switch state {
			case "vpn_enter_new_key_name_expected":
				return h.createKey(messageID, u.Message, bot)
			case "vpn_enter_existing_key_name_expected":
				return h.deleteKey(messageID, u.Message, bot)
			}
		}

		h.l.Error("unexpected vpn state", "state", state)
		h.state.Delete(u.Message.Chat.ID)
		_, err := bot.SendMessage(telegram.SendMessageParams{
			ChatID: u.Message.Chat.ID,
			Text:   i18n("vpn_unexpected_state"),
		})
		return err
	}

	// check commands
	for _, command := range u.Message.Commands() {
		switch command {
		case "/start":
			return h.handlePayload(u.Message, bot)
		case "/vpnhelp":
			if h.checkAccess(u.Message) {
				return h.help(u.Message, bot)
			}
			_, err := bot.SendMessage(telegram.SendMessageParams{
				ChatID: u.Message.Chat.ID,
				Text:   i18n("vpn_mislead"),
			})
			return err
		}
	}
	return nil
}

func (h *Handler) checkAccess(m *telegram.Message) bool {
	data, err := h.repo.GetSettings(context.Background(), repository.GetSettingsParams{
		ChatID: m.Chat.ID,
		Key:    "vpn.enabled",
	})
	if err != nil {
		h.l.Error("failed to get settings", "chat_id", m.Chat.ID, "error", err)
		return false
	}

	var settings struct {
		Enabled bool `json:"enabled"`
	}
	if err = json.Unmarshal(data.Value, &settings); err != nil {
		h.l.Error("failed to unmarshal settings", "chat_id", m.Chat.ID, "error", err)
		return false
	}

	return settings.Enabled
}

func (h *Handler) help(m *telegram.Message, bot *telegram.Bot) error {
	keys, err := h.client.GetKeys(m.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	var isDisabled = true
	_, err = bot.SendMessage(telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      i18n("vpn_welcome"),
		ParseMode: telegram.ParseModeHTML,
		LinkPreviewOptions: &telegram.LinkPreviewOptions{
			IsDisabled: isDisabled,
		},
		ReplyMarkup: telegram.InlineKeyboardMarkup{
			InlineKeyboard: h.makeKeyboard(keys),
		},
	})
	return err
}

func (h *Handler) handlePayload(m *telegram.Message, bot *telegram.Bot) error {
	parts := strings.SplitN(m.Text, " ", 2)
	if len(parts) < 2 {
		return nil
	}

	if parts[1] != "vpnhelp" {
		h.l.Warn("unexpected payload", "payload", parts[1])
		return nil
	}

	h.l.Info("enable vpn access", "chat_id", m.Chat.ID)

	if _, err := h.repo.SetSettings(context.Background(), repository.SetSettingsParams{
		ChatID: m.Chat.ID,
		Key:    "vpn.enabled",
		Value:  json.RawMessage(`{"enabled": true}`),
	}); err != nil {
		return fmt.Errorf("failed to save settings: %s", err)
	}

	return h.help(m, bot)
}

func (h *Handler) createKey(messageID int64, m *telegram.Message, bot *telegram.Bot) error {
	h.l.Info("create new key", "name", m.Text)
	if len(m.Text) > 64 {
		_, err := bot.SendMessage(telegram.SendMessageParams{
			ChatID: m.Chat.ID,
			Text:   i18n("vpn_enter_create_key_name_too_long"),
		})
		return err
	}

	key, err := h.client.CreateKey(m.Text, m.Chat.ID, m.From)
	if err != nil {
		return fmt.Errorf("failed to create new key: %w", err)
	}

	h.state.Delete(m.Chat.ID)

	if _, err := bot.DeleteMessage(telegram.DeleteMessageParams{
		ChatID:    m.Chat.ID,
		MessageID: messageID,
	}); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	buttons := [][]telegram.InlineKeyboardButton{
		{{Text: i18n("vpn_button_manage_key"), CallbackData: "vpn_manage_key"}},
	}
	_, err = bot.SendMessage(telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      i18n("vpn_key_created", key.Key),
		ParseMode: telegram.ParseModeHTML,
		ReplyMarkup: telegram.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	})
	return err
}

func (h *Handler) deleteKey(messageID int64, m *telegram.Message, bot *telegram.Bot) error {
	h.l.Info("delete key", "name", m.Text)
	keys, err := h.client.GetKeys(m.Chat.ID)
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	h.state.Delete(m.Chat.ID)

	if _, err := bot.DeleteMessage(telegram.DeleteMessageParams{
		ChatID:    m.Chat.ID,
		MessageID: messageID,
	}); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	buttons := [][]telegram.InlineKeyboardButton{
		{{Text: i18n("vpn_button_manage_key"), CallbackData: "vpn_manage_key"}},
	}

	for _, k := range keys {
		if k.Title == m.Text {
			if err = h.client.DeleteKey(k); err != nil {
				return fmt.Errorf("failed to delete key: %w", err)
			}
			_, err = bot.SendMessage(telegram.SendMessageParams{
				ChatID:    m.Chat.ID,
				Text:      i18n("vpn_key_deleted", k.Title),
				ParseMode: telegram.ParseModeHTML,
				ReplyMarkup: telegram.InlineKeyboardMarkup{
					InlineKeyboard: buttons,
				},
			})
			return err
		}
	}

	_, err = bot.SendMessage(telegram.SendMessageParams{
		ChatID: m.Chat.ID,
		Text:   i18n("vpn_key_not_found"),
		ReplyMarkup: telegram.InlineKeyboardMarkup{
			InlineKeyboard: buttons,
		},
	})
	return err
}

func (h *Handler) handleCallback(c *telegram.CallbackQuery, bot *telegram.Bot) error {
	// It's always a real message in this case
	m := c.MaybeInaccessibleMessage
	h.l.Info("got callback", "id", c.ID, "data", c.Data, "message_id", m.ID, "chat_id", m.Chat.ID)

	if err := bot.AnswerCallbackQuery(telegram.AnswerCallbackQueryParams{CallbackQueryID: c.ID}); err != nil {
		h.l.Error("failed to answer callback", "id", c.ID, "message_id", m.ID, "chat_id", m.Chat.ID)
	}

	switch c.Data {
	case "vpn_create_key":
		h.state.Set(m.Chat.ID, fmt.Sprintf("vpn_enter_new_key_name_expected_%d", m.ID))

		buttons := [][]telegram.InlineKeyboardButton{
			{{Text: i18n("vpn_button_back"), CallbackData: "vpn_back"}},
		}
		_, err := bot.EditMessageText(telegram.EditMessageTextParams{
			ChatID:    m.Chat.ID,
			MessageID: m.ID,
			Text:      i18n("vpn_enter_create_key_name"),
			ParseMode: telegram.ParseModeHTML,
			ReplyMarkup: telegram.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},
		})
		return err

	case "vpn_delete_key":
		keys, err := h.client.GetKeys(m.Chat.ID)
		if err != nil {
			return fmt.Errorf("failed to get keys: %w", err)
		}

		h.state.Set(m.Chat.ID, fmt.Sprintf("vpn_enter_existing_key_name_expected_%d", m.ID))

		text := []string{i18n("vpn_enter_delete_key_name_top")}
		for _, key := range keys {
			text = append(text, i18n("vpn_enter_delete_key_name_item", key.Title))
		}

		buttons := [][]telegram.InlineKeyboardButton{
			{{Text: i18n("vpn_button_cancel"), CallbackData: "vpn_back"}},
		}
		_, err = bot.EditMessageText(telegram.EditMessageTextParams{
			ChatID:    m.Chat.ID,
			MessageID: m.ID,
			Text:      strings.Join(text, "\n"),
			ParseMode: telegram.ParseModeHTML,
			ReplyMarkup: telegram.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},
		})
		return err

	case "vpn_manage_key":
		keys, err := h.client.GetKeys(m.Chat.ID)
		if err != nil {
			return fmt.Errorf("failed to get keys: %w", err)
		}

		text := []string{i18n("vpn_key_list_top")}
		for idx, key := range keys {
			text = append(text, i18n("vpn_key_list_item", idx+1, key.Title, key.Key))
		}
		text = append(text, i18n("vpn_key_list_bottom", len(keys)))

		buttons := [][]telegram.InlineKeyboardButton{
			{{Text: i18n("vpn_button_remove_key"), CallbackData: "vpn_delete_key"}},
			{{Text: i18n("vpn_button_back"), CallbackData: "vpn_back"}},
		}
		_, err = bot.EditMessageText(telegram.EditMessageTextParams{
			ChatID:    m.Chat.ID,
			MessageID: m.ID,
			Text:      strings.Join(text, "\n"),
			ParseMode: telegram.ParseModeHTML,
			ReplyMarkup: telegram.InlineKeyboardMarkup{
				InlineKeyboard: buttons,
			},
		})
		return err

	case "vpn_back":
		keys, err := h.client.GetKeys(m.Chat.ID)
		if err != nil {
			return fmt.Errorf("failed to get keys: %w", err)
		}

		h.state.Delete(m.Chat.ID)

		_, err = bot.EditMessageText(telegram.EditMessageTextParams{
			ChatID:    m.Chat.ID,
			MessageID: m.ID,
			Text:      i18n("vpn_welcome"),
			ParseMode: telegram.ParseModeHTML,
			LinkPreviewOptions: &telegram.LinkPreviewOptions{
				IsDisabled: true,
			},
			ReplyMarkup: telegram.InlineKeyboardMarkup{
				InlineKeyboard: h.makeKeyboard(keys),
			},
		})
		return err

	default:
		h.l.Error("ingnoring callback", "data", c.Data)
		return fmt.Errorf("unexpected callback data: %s", c.Data)
	}
}

func (h *Handler) makeKeyboard(keys []*VpnKey) [][]telegram.InlineKeyboardButton {
	buttons := [][]telegram.InlineKeyboardButton{}

	if len(keys) < 10 {
		buttons = append(buttons, []telegram.InlineKeyboardButton{
			{Text: i18n("vpn_button_create_key"), CallbackData: "vpn_create_key"},
		})
	}

	if len(keys) > 0 {
		buttons = append(buttons, []telegram.InlineKeyboardButton{
			{Text: i18n("vpn_button_manage_key"), CallbackData: "vpn_manage_key"},
		})
	}

	return buttons
}
