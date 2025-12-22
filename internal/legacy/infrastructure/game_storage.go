package infrastructure

import (
	"github.com/ailinykh/reposter/v2/internal/legacy/core"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CreateGameStorage is a default GameStorage factory
func CreateGameStorage(dbFile string) *GameStorage {
	conn, err := gorm.Open(sqlite.Open(dbFile+"?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(err)
	}

	if conn.Migrator().HasTable(&Player{}) && conn.Migrator().HasColumn(&Player{}, "chat_id") {
		conn.Migrator().RenameColumn(&Player{}, "chat_id", "game_id")
		conn.Migrator().RenameTable("faggot_entries", "faggot_rounds")
		conn.Migrator().RenameColumn(&Round{}, "chat_id", "game_id")
	} else {
		conn.AutoMigrate(&Player{}, &Round{})
	}
	return &GameStorage{conn}
}

// GameStorage implements core.IGameStorage interface
type GameStorage struct {
	conn *gorm.DB
}

// GetPlayers is a core.IGameStorage interface implementation
func (s *GameStorage) GetPlayers(gameID int64) ([]*core.User, error) {
	var dbPlayers []Player
	var corePlayers []*core.User
	s.conn.Where("game_id = ?", gameID).Find(&dbPlayers)
	for _, p := range dbPlayers {
		user := makeUser(p)
		corePlayers = append(corePlayers, user)
	}
	return corePlayers, nil
}

// GetRounds is a core.IGameStorage interface implementation
func (s *GameStorage) GetRounds(gameID int64) ([]*core.Round, error) {
	players, err := s.GetPlayers(gameID)
	if err != nil {
		return nil, err
	}
	var dbRounds []Round
	var coreRounds []*core.Round
	s.conn.Where("game_id = ?", gameID).Find(&dbRounds)
	for _, r := range dbRounds {
		for _, p := range players {
			if p.ID == r.UserID {
				coreRounds = append(coreRounds, &core.Round{Day: r.Day, Winner: p})
				break
			}
		}
	}
	return coreRounds, nil
}

// AddPlayer is a core.IGameStorage interface implementation
func (s *GameStorage) AddPlayer(gameID int64, user *core.User) error {
	player := Player{
		GameID:       gameID,
		UserID:       user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Username:     user.Username,
		LanguageCode: user.LanguageCode,
	}
	s.conn.Create(&player)
	return nil
}

// AddPlayer is a core.IGameStorage interface implementation
func (s *GameStorage) UpdatePlayer(gameID int64, user *core.User) error {
	player := Player{
		GameID: gameID,
		UserID: user.ID,
	}
	s.conn.Model(&player).Updates(Player{FirstName: user.FirstName, LastName: user.LastName, Username: user.Username})
	return nil
}

// AddRound is a core.IGameStorage interface implementation
func (s *GameStorage) AddRound(gameID int64, round *core.Round) error {
	dbRound := Round{
		GameID:   gameID,
		UserID:   round.Winner.ID,
		Day:      round.Day,
		Username: round.Winner.Username,
	}
	s.conn.Create(&dbRound)
	return nil
}

func makeUser(player Player) *core.User {
	return &core.User{
		ID:           player.UserID,
		FirstName:    player.FirstName,
		LastName:     player.LastName,
		Username:     player.Username,
		LanguageCode: player.LanguageCode,
	}
}
