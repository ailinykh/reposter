package helpers

import (
	"encoding/json"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateBoolSettingProvider(settingsProvider core.ISettingsProvider) core.IBoolSettingProvider {
	return &BoolSettingProvider{settingsProvider}
}

type BoolSettingProvider struct {
	settingsProvider core.ISettingsProvider
}

func (provider *BoolSettingProvider) GetBool(chatID core.ChatID, key core.SettingKey) bool {
	data, _ := provider.settingsProvider.GetData(chatID, key)

	var settings struct {
		Enabled bool
	}

	err := json.Unmarshal(data, &settings)
	if err != nil {
		return false
	}

	return settings.Enabled
}

func (provider *BoolSettingProvider) SetBool(chatID core.ChatID, key core.SettingKey, value bool) error {
	var settings struct {
		Enabled bool
	}
	settings.Enabled = value
	data, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	return provider.settingsProvider.SetData(chatID, key, data)
}
