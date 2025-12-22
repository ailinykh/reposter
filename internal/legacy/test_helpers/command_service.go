package test_helpers

import (
	"fmt"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateCommandService() *CommandServiceMock {
	return &CommandServiceMock{[]string{}}
}

type CommandServiceMock struct {
	ActionLog []string
}

// EnableCommands is a core.ICommandService interface implementation
func (service *CommandServiceMock) EnableCommands(chatID int64, commands []core.Command, bot core.IBot) error {
	service.ActionLog = append(service.ActionLog, fmt.Sprint("enable commands ", chatID, commands))
	return nil
}

// DisableCommands is a core.ICommandService interface implementation
func (service *CommandServiceMock) DisableCommands(chatID int64, commands []core.Command, bot core.IBot) error {
	service.ActionLog = append(service.ActionLog, fmt.Sprint("disable commands ", chatID, commands))
	return nil
}
