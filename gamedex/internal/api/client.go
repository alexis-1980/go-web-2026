package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gamedex/internal/catalog"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) FetchGames(path string) ([]catalog.Game, error) {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("requête %s échouée : %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code HTTP %d sur %s", resp.StatusCode, path)
	}

	var games []catalog.Game
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		return nil, fmt.Errorf("JSON jeux invalide : %w", err)
	}
	return games, nil
}

func (c *Client) FetchTournaments(gameID int) ([]catalog.Tournament, error) {
	path := fmt.Sprintf("/games/%d/tournaments", gameID)
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("requête %s échouée : %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code HTTP %d sur %s", resp.StatusCode, path)
	}

	var tournaments []catalog.Tournament
	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return nil, fmt.Errorf("JSON tournois invalide : %w", err)
	}
	return tournaments, nil
}
