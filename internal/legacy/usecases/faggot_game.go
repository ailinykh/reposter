package usecases

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ailinykh/reposter/v2/internal/core"
	legacy "github.com/ailinykh/reposter/v2/internal/legacy/core"
)

// CreateGameFlow is a simple GameFlow factory
func CreateGameFlow(l core.Logger, t legacy.ILocalizer, s legacy.IGameStorage, r legacy.IRand, settings legacy.ISettingsProvider, commandService legacy.ICommandService) *GameFlow {
	return &GameFlow{l, t, s, r, settings, commandService, sync.Mutex{}}
}

// GameFlow represents faggot game logic
type GameFlow struct {
	l              core.Logger
	t              legacy.ILocalizer
	s              legacy.IGameStorage
	r              legacy.IRand
	settings       legacy.ISettingsProvider
	commandService legacy.ICommandService
	mutex          sync.Mutex
}

// HandleText is a core.ITextHandler protocol implementation
func (flow *GameFlow) HandleText(message *legacy.Message, bot legacy.IBot) error {
	if !strings.HasPrefix(message.Text, "/pidor") {
		return fmt.Errorf("not implemented")
	}

	if strings.HasPrefix(message.Text, "/pidorules") {
		return flow.Rules(message, bot)
	} else if strings.HasPrefix(message.Text, "/pidoreg") {
		return flow.Add(message, bot)
	} else if strings.HasPrefix(message.Text, "/pidorstats") {
		return flow.Stats(strconv.Itoa(time.Now().Year()), message, bot)
	} else if strings.HasPrefix(message.Text, "/pidorall") {
		return flow.All(message, bot)
	} else if strings.HasPrefix(message.Text, "/pidorme") {
		return flow.Me(message, bot)
	}

	r := regexp.MustCompile(`^/pidor(\d+)$`)
	matches := r.FindAllStringSubmatch(message.Text, -1)
	if len(matches) > 0 && len(matches[0]) > 1 {
		return flow.Stats(matches[0][1], message, bot)
	} else {
		return flow.Play(message, bot)
	}
}

// Rules of the game
func (flow *GameFlow) Rules(message *legacy.Message, bot legacy.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_not_available_for_private"))
		return err
	}
	_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_rules"))
	return err
}

// Add a new player to game
func (flow *GameFlow) Add(message *legacy.Message, bot legacy.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_not_available_for_private"))
		return err
	}
	players, _ := flow.s.GetPlayers(message.Chat.ID)
	for _, p := range players {
		if p.ID == message.Sender.ID {
			if p.FirstName != message.Sender.FirstName || p.LastName != message.Sender.LastName || p.Username != message.Sender.Username {
				_ = flow.s.UpdatePlayer(message.Chat.ID, message.Sender)
				_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_info_updated"))
				return err
			}
			_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_already_in_game"))
			return err
		}
	}

	err := flow.s.AddPlayer(message.Chat.ID, message.Sender)
	if err != nil {
		return err
	}

	_, err = bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_added_to_game"))
	return err
}

// Play game
func (flow *GameFlow) Play(message *legacy.Message, bot legacy.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_not_available_for_private"))
		return err
	}
	flow.mutex.Lock()
	defer flow.mutex.Unlock()

	flow.checkSettings(message.Chat.ID, bot)

	flow.l.Info("game started", "chat_id", message.Chat.ID, "user", message.Sender)

	players, _ := flow.s.GetPlayers(message.Chat.ID)
	switch len(players) {
	case 0:
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_no_players", message.Sender.DisplayName()))
		return err
	case 1:
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_not_enough_players"))
		return err
	}

	games, _ := flow.s.GetRounds(message.Chat.ID)
	loc, _ := time.LoadLocation("Europe/Zurich")
	day := time.Now().In(loc).Format("2006-01-02")

	for _, r := range games {
		if r.Day == day {
			_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_winner_known", r.Winner.DisplayName()))
			return err
		}
	}

	winner := players[flow.r.GetRand(len(players))]

	if !bot.IsUserMemberOfChat(winner, message.Chat.ID) {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_winner_left"))
		return err
	}

	flow.l.Info("winner calculated", "date", day, "user", winner)

	if winner.ID == message.Sender.ID {
		if winner.FirstName != message.Sender.FirstName || winner.LastName != message.Sender.LastName || winner.Username != message.Sender.Username {
			err := flow.s.UpdatePlayer(message.Chat.ID, message.Sender)
			if err != nil {
				flow.l.Error(fmt.Errorf("failed to update player: %v", err))
			} else {
				flow.l.Info("player info updated", "user", winner)
			}
		}
	}

	round := &legacy.Round{Day: day, Winner: winner}
	flow.s.AddRound(message.Chat.ID, round)

	for i := 0; i <= 3; i++ {
		templates := []string{}
		for _, key := range flow.t.AllKeys() {
			if strings.HasPrefix(key, fmt.Sprintf("faggot_game_%d", i)) {
				templates = append(templates, key)
			}
		}
		template := templates[flow.r.GetRand(len(templates))]
		phrase := flow.t.I18n(message.Sender.LanguageCode, template)

		if i == 3 {
			// TODO: implementation detail leaked
			if len(winner.Username) == 0 {
				phrase = flow.t.I18n(message.Sender.LanguageCode, template, fmt.Sprintf(`<a href="tg://user?id=%d">%s %s</a>`, winner.ID, winner.FirstName, winner.LastName))
			} else {
				phrase = flow.t.I18n(message.Sender.LanguageCode, template, "@"+winner.Username)
			}
		}

		_, err := bot.SendText(phrase)
		if err != nil {
			flow.l.Error(err)
		}

		r := flow.r.GetRand(3) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}

	return nil
}

// All statistics for all time
func (flow *GameFlow) All(message *legacy.Message, bot legacy.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_not_available_for_private"))
		return err
	}

	entries, _ := flow.getStat(message)
	messages := []string{flow.t.I18n(message.Sender.LanguageCode, "faggot_all_top"), ""}
	for i, e := range entries {
		message := flow.t.I18n(message.Sender.LanguageCode, "faggot_all_entry", i+1, e.Player.DisplayName(), e.Score)
		messages = append(messages, message)
	}
	messages = append(messages, "", flow.t.I18n(message.Sender.LanguageCode, "faggot_all_bottom", len(entries)))
	_, err := bot.SendText(strings.Join(messages, "\n"))
	return err
}

// Stats returns current year statistics
func (flow *GameFlow) Stats(year string, message *legacy.Message, bot legacy.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_not_available_for_private"))
		return err
	}

	rounds, _ := flow.s.GetRounds(message.Chat.ID)
	entries := []Stat{}
	players := map[int64]bool{}

	for _, r := range rounds {
		players[r.Winner.ID] = true
		if strings.HasPrefix(r.Day, year) {
			index := Find(entries, r.Winner.ID)
			if index == -1 {
				entries = append(entries, Stat{Player: r.Winner, Score: 1})
			} else {
				entries[index].Score++
			}
		}
	}

	if len(entries) == 0 {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_stats_empty", year))
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			return entries[i].Player.Username < entries[j].Player.Username
		}
		return entries[i].Score > entries[j].Score
	})

	messages := []string{}
	if year == strconv.Itoa(time.Now().Year()) {
		messages = append(messages, flow.t.I18n(message.Sender.LanguageCode, "faggot_stats_top"))
	} else {
		messages = append(messages, flow.t.I18n(message.Sender.LanguageCode, "faggot_stats_top_year", year))
	}
	messages = append(messages, "")
	max := min(len(entries), 10) // Show Top 10 players only
	for i, e := range entries[:max] {
		message := flow.t.I18n(message.Sender.LanguageCode, "faggot_stats_entry", i+1, e.Player.DisplayName(), e.Score)
		messages = append(messages, message)
	}
	messages = append(messages, "", flow.t.I18n(message.Sender.LanguageCode, "faggot_stats_bottom", len(players)))
	_, err := bot.SendText(strings.Join(messages, "\n"))
	return err
}

// Me returns your personal statistics
func (flow *GameFlow) Me(message *legacy.Message, bot legacy.IBot) error {
	if message.IsPrivate {
		_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_not_available_for_private"))
		return err
	}

	entries, _ := flow.getStat(message)
	score := 0
	for _, e := range entries {
		if e.Player.ID == message.Sender.ID {
			score = e.Score
		}
	}
	_, err := bot.SendText(flow.t.I18n(message.Sender.LanguageCode, "faggot_me", message.Sender.DisplayName(), score))
	return err
}

func (flow *GameFlow) getStat(message *legacy.Message) ([]Stat, error) {
	entries := []Stat{}
	rounds, err := flow.s.GetRounds(message.Chat.ID)

	if err != nil {
		return nil, err
	}

	for _, r := range rounds {
		index := Find(entries, r.Winner.ID)
		if index == -1 {
			entries = append(entries, Stat{Player: r.Winner, Score: 1})
		} else {
			entries[index].Score++
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	return entries, nil
}

func (flow *GameFlow) checkSettings(chatID legacy.ChatID, bot legacy.IBot) error {
	data, err := flow.settings.GetData(chatID, legacy.SFaggotGameEnabled)

	if err != nil {
		flow.l.Error(err)
	}

	var settingsV1 struct {
		Enabled bool
	}

	err = json.Unmarshal(data, &settingsV1)
	if err != nil {
		flow.l.Error(err)
		// TODO: perform a migration?
	}

	if settingsV1.Enabled {
		return nil
	}

	settingsV1.Enabled = true
	data, err = json.Marshal(settingsV1)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	err = flow.settings.SetData(chatID, legacy.SFaggotGameEnabled, data)
	if err != nil {
		flow.l.Error(err)
		return err
	}

	commands := []legacy.Command{
		{Text: "pidor", Description: "play the game, see /pidorules first"},
		{Text: "pidorules", Description: "POTD game rules"},
		{Text: "pidoreg", Description: "register for POTD game"},
		{Text: "pidorstats", Description: "POTD game stats for this year"},
		{Text: "pidorall", Description: "POTD game stats for all time"},
		{Text: "pidorme", Description: "POTD personal stats"},
	}

	return flow.commandService.EnableCommands(chatID, commands, bot)
}
