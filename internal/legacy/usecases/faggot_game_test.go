package usecases_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"github.com/ailinykh/reposter/v2/internal/legacy/test_helpers"
	"github.com/ailinykh/reposter/v2/internal/legacy/usecases"
	"github.com/stretchr/testify/assert"
)

func Test_AllTheCommands_WorksOnlyInGroupChats(t *testing.T) {
	game, bot, _ := makeSUT(map[string]string{"faggot_not_available_for_private": "group only"})
	message := makeGameMessage(1, "Faggot")
	message.IsPrivate = true

	game.Rules(message, bot)
	game.Add(message, bot)
	game.Play(message, bot)
	game.Stats("2020", message, bot)
	game.All(message, bot)
	game.Me(message, bot)

	for _, m := range bot.SentMessages {
		assert.Equal(t, "group only", m)
	}
}
func Test_RulesCommand_DeliversRules(t *testing.T) {
	game, bot, _ := makeSUT(map[string]string{"faggot_rules": "Game rules:"})
	message := makeGameMessage(1, "Faggot")

	game.Rules(message, bot)

	assert.Equal(t, "Game rules:", bot.SentMessages[0])
}

func Test_Add_ChecksAndReplacesPlayerInfoIfNeeded(t *testing.T) {
	game, bot, storage := makeSUT()
	message := makeGameMessage(1, "Faggot")
	player := *message.Sender
	player.Username = "old_username"
	storage.players = []*core.User{&player}

	game.Add(message, bot)

	assert.Equal(t, []*core.User{message.Sender}, storage.players)
	assert.Equal(t, []string{"faggot_info_updated"}, bot.SentMessages)
}

func Test_Add_AppendsPlayerInGameOnlyOnce(t *testing.T) {
	game, bot, storage := makeSUT(map[string]string{
		"faggot_added_to_game":   "Player added",
		"faggot_already_in_game": "Player already in game",
	})
	message := makeGameMessage(1, "Faggot")

	game.Add(message, bot)

	assert.Equal(t, storage.players, []*core.User{message.Sender})
	assert.Equal(t, "Player added", bot.SentMessages[0])

	game.Add(message, bot)

	assert.Equal(t, storage.players, []*core.User{message.Sender})
	assert.Equal(t, "Player already in game", bot.SentMessages[1])
}

func Test_Play_RespondsWithNoPlayers(t *testing.T) {
	message := makeGameMessage(1, "Faggot")
	game, bot, _ := makeSUT(map[string]string{
		"faggot_no_players": "Nobody in game. So you win, %s!",
	}, message)

	err := game.Play(message, bot)

	assert.Nil(t, err)
	assert.Equal(t, "Nobody in game. So you win, Faggot!", bot.SentMessages[0])
}

func Test_Play_RespondsNotEnoughPlayers(t *testing.T) {
	message := makeGameMessage(1, "Faggot")
	game, bot, _ := makeSUT(map[string]string{
		"faggot_not_enough_players": "Not enough players",
	}, message)

	game.Add(message, bot)
	game.Play(message, bot)

	assert.Equal(t, "Not enough players", bot.SentMessages[1])
}

func Test_Play_RespondsWithCurrentGameResult(t *testing.T) {
	m1 := makeGameMessage(1, "")
	m2 := makeGameMessage(2, "")
	game, bot, storage := makeSUT(map[string]string{
		"faggot_game_0_0": "00",
		"faggot_game_1_0": "10",
		"faggot_game_2_0": "20",
		"faggot_game_3_0": "%s",
	}, m1, m2)
	bot.ChatMembers[0] = []string{""}

	game.Add(m1, bot)
	game.Add(m2, bot)
	game.Play(m1, bot)

	winner := storage.rounds[0].Winner
	phrase := fmt.Sprintf(`<a href="tg://user?id=%d">%s %s</a>`, winner.ID, winner.FirstName, winner.LastName)
	assert.Equal(t, "00", bot.SentMessages[2])
	assert.Equal(t, "10", bot.SentMessages[3])
	assert.Equal(t, "20", bot.SentMessages[4])
	assert.Equal(t, phrase, bot.SentMessages[5])
}
func Test_Play_RespondsWinnerAlreadyKnown(t *testing.T) {
	m1 := makeGameMessage(1, "Faggot1")
	m2 := makeGameMessage(2, "Faggot2")
	game, bot, storage := makeSUT(map[string]string{
		"faggot_game_0_0":     "00",
		"faggot_game_1_0":     "10",
		"faggot_game_2_0":     "20",
		"faggot_game_3_0":     "30 %s",
		"faggot_winner_known": "Winner already known %s",
	}, m1)
	bot.ChatMembers[0] = []string{"Faggot1", "Faggot2"}

	game.Add(m1, bot)
	game.Add(m2, bot)
	game.Play(m1, bot)

	winner := storage.rounds[0].Winner.Username
	assert.Equal(t, fmt.Sprintf("30 @%s", winner), bot.SentMessages[5])

	game.Play(m1, bot)

	assert.Equal(t, fmt.Sprintf("Winner already known %s", winner), bot.SentMessages[6])
}

func Test_Play_RespondsWinnerLeftTheChat(t *testing.T) {
	m1 := makeGameMessage(1, "Faggot1")
	m2 := makeGameMessage(2, "Faggot2")
	game, bot, storage := makeSUT(map[string]string{
		"faggot_winner_left": "winner left",
	}, m1)

	storage.players = []*core.User{m1.Sender, m2.Sender}

	game.Play(m1, bot)

	assert.Equal(t, []*core.Round{}, storage.rounds)
	assert.Equal(t, []string{"winner left"}, bot.SentMessages)
}

func Test_Stats_RespondsWithDescendingResultsForCurrentYear(t *testing.T) {
	year := strconv.Itoa(time.Now().Year())
	game, bot, storage := makeSUT(map[string]string{
		"faggot_stats_top":    "top",
		"faggot_stats_entry":  "index:%d,player:%s,scores:%d",
		"faggot_stats_bottom": "total_players:%d",
	})

	expected := []string{
		"top",
		"",
		"index:1,player:Faggot3,scores:3",
		"index:2,player:Faggot1,scores:2",
		"index:3,player:Faggot2,scores:1",
		"",
		"total_players:4",
	}

	m1 := makeGameMessage(1, "Faggot1")
	m2 := makeGameMessage(2, "Faggot2")
	m3 := makeGameMessage(3, "Faggot3")
	m4 := makeGameMessage(4, "Faggot4")

	storage.rounds = []*core.Round{
		{Day: year + "-01-01", Winner: m2.Sender},
		{Day: "2020-01-02", Winner: m3.Sender},
		{Day: "2020-01-03", Winner: m4.Sender},
		{Day: year + "-01-02", Winner: m3.Sender},
		{Day: year + "-01-03", Winner: m3.Sender},
		{Day: year + "-01-04", Winner: m3.Sender},
		{Day: year + "-01-05", Winner: m1.Sender},
		{Day: year + "-01-06", Winner: m1.Sender},
	}

	game.Stats(year, m1, bot)
	assert.Equal(t, expected, strings.Split(bot.SentMessages[0], "\n"))
}

func Test_Stats_RespondsOnlyForTop10Players(t *testing.T) {
	game, bot, storage := makeSUT(map[string]string{
		"faggot_stats_top":    "top",
		"faggot_stats_entry":  "index:%d,player:%s,scores:%d",
		"faggot_stats_bottom": "total_players:%d",
	})

	expected := []string{
		"top",
		"",
		"index:1,player:Faggot01,scores:1",
		"index:2,player:Faggot02,scores:1",
		"index:3,player:Faggot03,scores:1",
		"index:4,player:Faggot04,scores:1",
		"index:5,player:Faggot05,scores:1",
		"index:6,player:Faggot06,scores:1",
		"index:7,player:Faggot07,scores:1",
		"index:8,player:Faggot08,scores:1",
		"index:9,player:Faggot09,scores:1",
		"index:10,player:Faggot10,scores:1",
		"",
		"total_players:99",
	}

	var messages []*core.Message
	var i int64
	for i = 1; i < 100; i++ {
		messages = append(messages, makeGameMessage(i, fmt.Sprintf("Faggot%02d", i)))
	}

	year := strconv.Itoa(time.Now().Year())
	for i, m := range messages {
		day := fmt.Sprintf("%s-%02d-%02d", year, i/30+1, i%30)
		storage.rounds = append(storage.rounds, &core.Round{Day: day, Winner: m.Sender})
	}

	game.Stats(year, messages[0], bot)
	assert.Equal(t, expected, strings.Split(bot.SentMessages[0], "\n"))
}

func Test_All_RespondsWithDescendingResultsForAllTime(t *testing.T) {
	game, bot, storage := makeSUT(map[string]string{
		"faggot_all_top":    "top",
		"faggot_all_entry":  "index:%d,player:%s,scores:%d",
		"faggot_all_bottom": "total_players:%d",
	})

	expected := []string{
		"top",
		"",
		"index:1,player:Faggot3,scores:4",
		"index:2,player:Faggot1,scores:2",
		"index:3,player:Faggot2,scores:1",
		"",
		"total_players:3",
	}

	m1 := makeGameMessage(1, "Faggot1")
	m2 := makeGameMessage(2, "Faggot2")
	m3 := makeGameMessage(3, "Faggot3")

	storage.rounds = []*core.Round{
		{Day: "2021-01-01", Winner: m2.Sender},
		{Day: "2020-01-02", Winner: m3.Sender},
		{Day: "2020-01-02", Winner: m3.Sender},
		{Day: "2021-01-03", Winner: m3.Sender},
		{Day: "2021-01-04", Winner: m3.Sender},
		{Day: "2021-01-05", Winner: m1.Sender},
		{Day: "2021-01-06", Winner: m1.Sender},
	}

	game.All(m1, bot)
	assert.Equal(t, expected, strings.Split(bot.SentMessages[0], "\n"))
}

func Test_Me_RespondsWithPersonalStat(t *testing.T) {
	game, bot, storage := makeSUT(map[string]string{
		"faggot_me": "username:%s,scores:%d",
	})

	m1 := makeGameMessage(1, "Faggot1")
	m2 := makeGameMessage(2, "Faggot2")

	storage.rounds = []*core.Round{
		{Day: "2021-01-01", Winner: m2.Sender},
		{Day: "2021-01-05", Winner: m1.Sender},
		{Day: "2021-01-06", Winner: m1.Sender},
	}

	game.Me(m1, bot)
	assert.Equal(t, fmt.Sprintf("username:%s,scores:%d", m1.Sender.Username, 2), bot.SentMessages[0])

	game.Me(m2, bot)
	assert.Equal(t, fmt.Sprintf("username:%s,scores:%d", m2.Sender.Username, 1), bot.SentMessages[1])
}

// Helpers

func makeGameMessage(id int64, username string) *core.Message {
	player := &core.User{
		ID:        id,
		FirstName: "FirstName" + fmt.Sprint(id),
		LastName:  "LastName" + fmt.Sprint(id),
		Username:  username,
	}
	return &core.Message{ID: 0, Chat: &core.Chat{ID: 0}, Sender: player}
}

func makeSUT(args ...interface{}) (*usecases.GameFlow, *test_helpers.FakeBot, *GameStorageMock) {
	dict := map[string]string{}
	storage := &GameStorageMock{players: []*core.User{}, rounds: []*core.Round{}}
	bot := test_helpers.CreateBot()
	l := &test_helpers.FakeLogger{}
	s := test_helpers.CreateSettingsProvider()

	for _, arg := range args {
		switch opt := arg.(type) {
		case map[string]string:
			dict = opt
		case *core.Message:
			s.SetData(opt.Chat.ID, "key", []byte{})
		}
	}

	t := test_helpers.CreateLocalizer(dict)
	c := test_helpers.CreateCommandService()
	r := &RandMock{}
	game := usecases.CreateGameFlow(l, t, storage, r, s, c)
	return game, bot, storage
}

// GameStorageMock

type GameStorageMock struct {
	players []*core.User
	rounds  []*core.Round
}

func (s *GameStorageMock) AddPlayer(gameID int64, player *core.User) error {
	s.players = append(s.players, player)
	return nil
}

func (s *GameStorageMock) UpdatePlayer(gameID int64, user *core.User) error {
	for _, p := range s.players {
		if p.ID == user.ID {
			p.FirstName = user.FirstName
			p.LastName = user.LastName
			p.Username = user.Username
		}
	}
	return nil
}

func (s *GameStorageMock) GetPlayers(gameID int64) ([]*core.User, error) {
	return s.players, nil
}

func (s *GameStorageMock) AddRound(gameID int64, round *core.Round) error {
	s.rounds = append(s.rounds, round)
	return nil
}

func (s *GameStorageMock) GetRounds(gameID int64) ([]*core.Round, error) {
	return s.rounds, nil
}

// IRandMock

type RandMock struct{}

func (RandMock) GetRand(int) int {
	return 0
}
