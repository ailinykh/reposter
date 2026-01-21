package fotd

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ailinykh/reposter/v3/internal/repository"
	"github.com/ailinykh/reposter/v3/pkg/telegram"
)

type GameRepository interface {
	GetPlayers(ctx context.Context, chatID int64) ([]repository.GamePlayer, error)
	CreatePlayer(ctx context.Context, arg repository.CreatePlayerParams) (repository.GamePlayer, error)
	UpdatePlayer(ctx context.Context, arg repository.UpdatePlayerParams) ([]repository.GamePlayer, error)
	GetRounds(ctx context.Context, chatID int64) ([]repository.GetRoundsRow, error)
	CreateRound(ctx context.Context, arg repository.CreateRoundParams) (repository.GameRound, error)
}

func NewGame(ctx context.Context, logger *slog.Logger, repo GameRepository) *Game {
	return &Game{
		ctx:  ctx,
		l:    logger,
		repo: repo,
	}
}

type Game struct {
	ctx  context.Context
	l    *slog.Logger
	repo GameRepository
}

func (g *Game) Handle(u *telegram.Update, bot *telegram.Bot) error {
	if u.Message == nil || len(u.Message.Commands()) == 0 {
		return nil
	}

	command := u.Message.Commands()[0]
	// command could be bot specific
	before, _, found := strings.Cut(command, "@")
	if found {
		command = before
	}

	if !strings.HasPrefix(command, "/pidor") {
		return nil
	}

	if u.Message.Chat.Private() {
		if u.Message.Chat.Type == "private" {
			_, err := bot.SendMessage(&telegram.SendMessageParams{
				ChatID:    u.Message.Chat.ID,
				Text:      i18n("faggot_not_available_for_private"),
				ParseMode: telegram.ParseModeHTML,
			})
			return err
		}
	}

	g.l.Info("executing", "command", command)

	switch command {
	case "/pidorules":
		return g.rules(u.Message, bot)
	case "/pidoreg":
		return g.reg(u.Message, bot)
	case "/pidorstats":
		return g.stats(strconv.Itoa(time.Now().Year()), u.Message, bot)
	case "/pidorall":
		return g.all(u.Message, bot)
	case "/pidorme":
		return g.me(u.Message, bot)
	default:

		matches := regexp.MustCompile(`^/pidor(\d+)$`).FindAllStringSubmatch(u.Message.Text, -1)
		if len(matches) > 0 && len(matches[0]) > 1 {
			return g.stats(matches[0][1], u.Message, bot)
		} else {
			return g.play(u.Message, bot)
		}
	}
}

func (g *Game) rules(m *telegram.Message, bot *telegram.Bot) error {
	g.l.Debug("game rules requested", "chat_id", m.Chat.ID, "user_id", m.From.ID)
	_, err := bot.SendMessage(&telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      i18n("faggot_rules"),
		ParseMode: telegram.ParseModeHTML,
	})
	return err
}

func (g *Game) reg(m *telegram.Message, bot *telegram.Bot) error {
	g.l.Debug("game registration", "chat_id", m.Chat.ID, "user_id", m.From.ID)
	players, _ := g.repo.GetPlayers(g.ctx, m.Chat.ID)
	for _, p := range players {
		if p.UserID == m.From.ID {
			if p.FirstName != m.From.FirstName || p.LastName != m.From.LastName || p.Username != m.From.Username {
				_, err := g.repo.UpdatePlayer(g.ctx, repository.UpdatePlayerParams{
					UserID:    m.From.ID,
					FirstName: m.From.FirstName,
					LastName:  m.From.LastName,
					Username:  m.From.Username,
				})
				_, err = bot.SendMessage(&telegram.SendMessageParams{
					ChatID:    m.Chat.ID,
					Text:      i18n("faggot_info_updated"),
					ParseMode: telegram.ParseModeHTML,
				})
				return err
			}

			_, err := bot.SendMessage(&telegram.SendMessageParams{
				ChatID:    m.Chat.ID,
				Text:      i18n("faggot_already_in_game"),
				ParseMode: telegram.ParseModeHTML,
			})
			return err
		}
	}

	_, err := g.repo.CreatePlayer(g.ctx, repository.CreatePlayerParams{
		ChatID:    m.Chat.ID,
		UserID:    m.From.ID,
		FirstName: m.From.FirstName,
		LastName:  m.From.LastName,
		Username:  m.From.Username,
	})
	if err != nil {
		return err
	}

	_, err = bot.SendMessage(&telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      i18n("faggot_added_to_game"),
		ParseMode: telegram.ParseModeHTML,
	})
	return err
}

func (g *Game) stats(year string, m *telegram.Message, bot *telegram.Bot) error {
	g.l.Debug("game statistics by year", "chat_id", m.Chat.ID, "user_id", m.From.ID)

	rounds, err := g.repo.GetRounds(g.ctx, m.Chat.ID)
	if err != nil {
		return err
	}

	entries, players := calculateStatistics(rounds, func(gr repository.GetRoundsRow) bool {
		return strconv.Itoa(gr.CreatedAt.Year()) == year
	})

	if len(entries) == 0 {
		_, err := bot.SendMessage(&telegram.SendMessageParams{
			ChatID:    m.Chat.ID,
			Text:      i18n("faggot_stats_empty", year),
			ParseMode: telegram.ParseModeHTML,
		})
		return err
	}

	messages := []string{}
	if year == strconv.Itoa(time.Now().Year()) {
		messages = append(messages, i18n("faggot_stats_top"))
	} else {
		messages = append(messages, i18n("faggot_stats_top_year", year))
	}
	messages = append(messages, "")
	max := min(len(entries), 10) // Show Top 10 players only
	for i, e := range players[:max] {
		message := i18n("faggot_stats_entry", i+1, e, entries[e])
		messages = append(messages, message)
	}
	messages = append(messages, "", i18n("faggot_stats_bottom", len(players)))
	_, err = bot.SendMessage(&telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      strings.Join(messages, "\n"),
		ParseMode: telegram.ParseModeHTML,
	})
	return err
}

func (g *Game) all(m *telegram.Message, bot *telegram.Bot) error {
	g.l.Debug("game statistics", "chat_id", m.Chat.ID, "user_id", m.From.ID)

	rounds, err := g.repo.GetRounds(g.ctx, m.Chat.ID)
	if err != nil {
		return err
	}

	entries, players := calculateStatistics(rounds, func(gr repository.GetRoundsRow) bool { return true })

	messages := []string{i18n("faggot_all_top"), ""}
	for i, p := range players {
		message := i18n("faggot_all_entry", i+1, p, entries[p])
		messages = append(messages, message)
	}
	messages = append(messages, "", i18n("faggot_all_bottom", len(players)))
	_, err = bot.SendMessage(&telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      strings.Join(messages, "\n"),
		ParseMode: telegram.ParseModeHTML,
	})
	return err
}

func (g *Game) me(m *telegram.Message, bot *telegram.Bot) error {
	g.l.Debug("game statistics for person", "chat_id", m.Chat.ID, "user_id", m.From.ID)

	rounds, err := g.repo.GetRounds(g.ctx, m.Chat.ID)
	if err != nil {
		return err
	}

	entries, _ := calculateStatistics(rounds, func(gr repository.GetRoundsRow) bool {
		return gr.UserID == m.From.ID
	})

	_, err = bot.SendMessage(&telegram.SendMessageParams{
		ChatID:    m.Chat.ID,
		Text:      i18n("faggot_me", m.From.DisplayName(), entries[m.From.DisplayName()]),
		ParseMode: telegram.ParseModeHTML,
	})
	return err
}

var mutex sync.Mutex

func (g *Game) play(m *telegram.Message, bot *telegram.Bot) error {
	g.l.Info("game started", "chat_id", m.Chat.ID, "user_id", m.From.ID)
	mutex.Lock()
	defer mutex.Unlock()

	// TODO: chat settigs and bot menu

	players, _ := g.repo.GetPlayers(g.ctx, m.Chat.ID)
	switch len(players) {
	case 0:
		_, err := bot.SendMessage(&telegram.SendMessageParams{
			ChatID:    m.Chat.ID,
			Text:      i18n("faggot_no_players", m.From.DisplayName()),
			ParseMode: telegram.ParseModeHTML,
		})
		return err
	case 1:
		_, err := bot.SendMessage(&telegram.SendMessageParams{
			ChatID:    m.Chat.ID,
			Text:      i18n("faggot_not_enough_players"),
			ParseMode: telegram.ParseModeHTML,
		})
		return err
	}

	rounds, _ := g.repo.GetRounds(g.ctx, m.Chat.ID)
	loc, _ := time.LoadLocation("Europe/Zurich")
	today := time.Now().In(loc).Truncate(24 * time.Hour)
	for _, r := range rounds {
		if r.CreatedAt.Truncate(24 * time.Hour).Equal(today) {
			// already have `displayName(champion)` in `Username` field
			_, err := bot.SendMessage(&telegram.SendMessageParams{
				ChatID:    m.Chat.ID,
				Text:      i18n("faggot_champion_known", r.Username),
				ParseMode: telegram.ParseModeHTML,
			})
			return err
		}
	}

	champion := players[rand.IntN(len(players))]

	if !bot.IsUserMemberOfChat(&telegram.GetChatMemberParams{ChatID: m.Chat.ID, UserID: champion.UserID}) {
		_, err := bot.SendMessage(&telegram.SendMessageParams{
			ChatID:    m.Chat.ID,
			Text:      i18n("faggot_champion_left"),
			ParseMode: telegram.ParseModeHTML,
		})
		return err
	}

	g.l.Info("champion calculated", "date", today, "username", displayName(champion))

	g.repo.CreateRound(g.ctx, repository.CreateRoundParams{
		ChatID:   m.Chat.ID,
		UserID:   champion.UserID,
		Username: displayName(champion),
	})

	for i := 0; i <= 3; i++ {
		templates := []string{}
		for _, key := range allKeys() {
			if strings.HasPrefix(key, fmt.Sprintf("faggot_game_%d", i)) {
				templates = append(templates, key)
			}
		}
		template := templates[rand.IntN(len(templates))]
		phrase := i18n(template)

		if i == 3 {
			phrase = i18n(template, mention(champion))
		}

		_, err := bot.SendMessage(&telegram.SendMessageParams{
			ChatID:    m.Chat.ID,
			Text:      phrase,
			ParseMode: telegram.ParseModeHTML,
		})
		if err != nil {
			g.l.Error("failed to send message", "error", err)
		}

		r := rand.IntN(3) + 1
		time.Sleep(time.Duration(r) * time.Second)
	}
	return nil
}
