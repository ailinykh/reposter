package usecases_test

import (
	"testing"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_EnableCommands_DoNotCallsSetCommandsMoreThanOneTime(t *testing.T) {
	bot := test_helpers.CreateBot()
	service := usecases.CreateCommandService()

	service.EnableCommands(1, []core.Command{{Text: "c1", Description: "d1"}}, bot)
	assert.Equal(t, []string{"get commands 1", "set commands 1 [{c1 d1}]"}, bot.ActionLog)

	service.EnableCommands(1, []core.Command{{Text: "c1", Description: "d1"}}, bot)
	assert.Equal(t, []string{"get commands 1", "set commands 1 [{c1 d1}]"}, bot.ActionLog)

	service.EnableCommands(1, []core.Command{{Text: "c2", Description: "d2"}}, bot)
	assert.Equal(t, []string{"get commands 1", "set commands 1 [{c1 d1}]", "set commands 1 [{c1 d1} {c2 d2}]"}, bot.ActionLog)
}

func Test_DisableCommands_DoNotCallsSetCommandsMoreThanOneTime(t *testing.T) {
	bot := test_helpers.CreateBot()
	service := usecases.CreateCommandService()

	service.EnableCommands(14, []core.Command{
		{Text: "one", Description: "1"},
		{Text: "two", Description: "2"},
	}, bot)
	assert.Equal(t, []string{"get commands 14", "set commands 14 [{one 1} {two 2}]"}, bot.ActionLog)

	service.DisableCommands(14, []core.Command{{Text: "two", Description: "2"}}, bot)
	assert.Equal(t, []string{"get commands 14", "set commands 14 [{one 1} {two 2}]", "set commands 14 [{one 1}]"}, bot.ActionLog)
}
