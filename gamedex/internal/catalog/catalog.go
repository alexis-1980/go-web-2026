package catalog

import (
	"fmt"
	"strings"
)

type Game struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Genre       string `json:"genre"`
	Platform    string `json:"platform"`
	Description string `json:"description"`
	Year        int    `json:"year"`
}

type Tournament struct {
	ID     int    `json:"id"`
	GameID int    `json:"game_id"`
	Name   string `json:"name"`
	City   string `json:"city"`
	Date   string `json:"date"`
}

type Store struct {
	games       []Game
	byID        map[int]Game
	tournaments map[int][]Tournament
}

func NewStore(games []Game, tournamentsByGame map[int][]Tournament) *Store {
	byID := make(map[int]Game, len(games))
	for _, g := range games {
		byID[g.ID] = g
	}

	tCopy := make(map[int][]Tournament, len(tournamentsByGame))
	for k, v := range tournamentsByGame {
		tCopy[k] = v
	}

	return &Store{
		games:       games,
		byID:        byID,
		tournaments: tCopy,
	}
}

func (s *Store) Genres() []string {
	seen := make(map[string]struct{})
	var genres []string
	for _, g := range s.games {
		if g.Genre == "" {
			continue
		}
		if _, ok := seen[g.Genre]; !ok {
			seen[g.Genre] = struct{}{}
			genres = append(genres, g.Genre)
		}
	}
	return genres
}

func (s *Store) SearchGames(query, genre string) []Game {
	q := strings.ToLower(strings.TrimSpace(query))
	gFilter := strings.TrimSpace(genre)

	var out []Game
	for _, g := range s.games {
		if gFilter != "" && !strings.EqualFold(g.Genre, gFilter) {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(g.Name), q) {
			continue
		}
		out = append(out, g)
	}
	return out
}

func (s *Store) GetGame(id int) (Game, bool) {
	g, ok := s.byID[id]
	return g, ok
}

func (s *Store) GetTournaments(gameID int) []Tournament {
	list := s.tournaments[gameID]
	out := make([]Tournament, len(list))
	copy(out, list)
	return out
}

func (s *Store) PrintSummary() {
	for _, g := range s.games {
		t := s.tournaments[g.ID]
		fmt.Printf("- %s (%d) : %d tournoi(s)\n", g.Name, g.Year, len(t))
	}
}
