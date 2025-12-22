package usecases

import (
	"sort"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
)

func CreateCommandService() core.ICommandService {
	return &CommandService{make(map[int64][]core.Command)}
}

type CommandService struct {
	cache map[int64][]core.Command
}

type ByText []core.Command

func (t ByText) Len() int           { return len(t) }
func (t ByText) Less(i, j int) bool { return t[i].Text < t[j].Text }
func (t ByText) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

// EnableCommands is a core.ICommandService interface implementation
func (service *CommandService) EnableCommands(chatID int64, commands []core.Command, bot core.IBot) error {
	var existing []core.Command
	var err error
	if found, ok := service.cache[chatID]; ok {
		existing = found
	} else {
		existing, err = bot.GetCommands(chatID)
		if err != nil {
			return nil
		}
	}

	new := []core.Command{}
	for _, c := range commands {
		if service.contains(c, existing) {
			continue
		}
		new = append(new, c)
	}

	if len(new) == 0 {
		return nil
	}

	new = append(new, existing...)
	service.cache[chatID] = new
	sort.Sort(ByText(new))
	return bot.SetCommands(chatID, new)
}

// DisableCommands is a core.ICommandService interface implementation
func (service *CommandService) DisableCommands(chatID int64, commands []core.Command, bot core.IBot) error {
	var existing []core.Command
	var err error
	if found, ok := service.cache[chatID]; ok {
		existing = found
	} else {
		existing, err = bot.GetCommands(chatID)
		if err != nil {
			return nil
		}
	}

	actual := []core.Command{}
	for _, c := range existing {
		if service.contains(c, commands) {
			continue
		}
		actual = append(actual, c)
	}

	service.cache[chatID] = actual
	return bot.SetCommands(chatID, actual)
}

func (CommandService) contains(command core.Command, commands []core.Command) bool {
	for _, c := range commands {
		if c.Text == command.Text {
			return true
		}
	}
	return false
}
