package fotd

import (
	"fmt"
	"sort"

	"github.com/ailinykh/reposter/v3/internal/repository"
)

func calculateStatistics(rounds []repository.GetRoundsRow, filter func(repository.GetRoundsRow) bool) (map[string]int, []string) {
	entries := map[string]int{}
	for _, r := range rounds {
		if filter(r) {
			// User can change username and invoke /reg command in any time
			// db could have different usernames for a single user in rounds
			// we shoud rely on user_id to avoid duplicated rows for each user
			username := r.Username
			if r.ActualUsername.Valid && len(r.ActualUsername.String) > 0 {
				username = r.ActualUsername.String
			}
			entries[username]++
		}
	}

	players := make([]string, 0, len(entries))
	for username := range entries {
		players = append(players, username)
	}

	sort.Slice(players, func(i, j int) bool {
		if entries[players[i]] == entries[players[j]] {
			return players[i] < players[j]
		}
		return entries[players[i]] > entries[players[j]]
	})

	return entries, players
}

func displayName(p repository.GamePlayer) string {
	if len(p.Username) > 0 {
		return p.Username
	}

	if len(p.LastName) == 0 {
		return p.FirstName
	}

	return p.FirstName + " " + p.LastName
}

func mention(p repository.GamePlayer) string {
	if len(p.Username) > 0 {
		return "@" + p.Username
	}
	return fmt.Sprintf(`<a href="tg://user?id=%d">%s %s</a>`, p.UserID, p.FirstName, p.LastName)
}
