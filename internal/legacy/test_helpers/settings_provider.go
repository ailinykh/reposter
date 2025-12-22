package test_helpers

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateSettingsProvider() *FakeSettingsProvider {
	return &FakeSettingsProvider{make(map[int64]map[core.SettingKey][]byte), nil}
}

type FakeSettingsProvider struct {
	Data map[core.ChatID]map[core.SettingKey][]byte
	Err  error
}

// GetSettings is a core.ISettingsProvider interface implementation
func (s *FakeSettingsProvider) GetData(chatID core.ChatID, key core.SettingKey) ([]byte, error) {
	if chat, ok := s.Data[chatID]; ok {
		if settings, ok := chat[key]; ok {
			return settings, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

// SetSettings is a core.ISettingsProvider interface implementation
func (s *FakeSettingsProvider) SetData(chatID core.ChatID, key core.SettingKey, data []byte) error {
	if _, ok := s.Data[chatID]; !ok {
		s.Data[chatID] = map[core.SettingKey][]byte{}
	}
	s.Data[chatID][key] = data
	return nil
}
