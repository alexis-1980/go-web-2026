package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const apiBase = "http://localhost:8081/api"

type Game struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Genre string `json:"genre"`
	Year  int    `json:"year"`
}

type Tournament struct {
	ID     int    `json:"id"`
	GameID int    `json:"game_id"`
	Name   string `json:"name"`
	City   string `json:"city"`
}

func newClient() *http.Client {
	return &http.Client{
		Timeout: 2 * time.Second,
	}
}

func FetchGames(client *http.Client, path string) ([]Game, error) {
	resp, err := client.Get(apiBase + path)
	if err != nil {
		return nil, fmt.Errorf("requête %s échouée : %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code HTTP %d sur %s", resp.StatusCode, path)
	}

	var games []Game
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		return nil, fmt.Errorf("JSON jeux invalide : %w", err)
	}
	return games, nil
}

func FetchTournaments(client *http.Client, gameID int) ([]Tournament, error) {
	path := fmt.Sprintf("/games/%d/tournaments", gameID)
	resp, err := client.Get(apiBase + path)
	if err != nil {
		return nil, fmt.Errorf("requête %s échouée : %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("code HTTP %d sur %s", resp.StatusCode, path)
	}

	var tournaments []Tournament
	if err := json.NewDecoder(resp.Body).Decode(&tournaments); err != nil {
		return nil, fmt.Errorf("JSON tournois invalide : %w", err)
	}
	return tournaments, nil
}

func main() {
	port := flag.String("port", "8080", "port d'écoute du serveur HTTP")
	flag.Parse()

	client := newClient()

	games, err := FetchGames(client, "/games")
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	for _, g := range games {
		tournaments, err := FetchTournaments(client, g.ID)
		if err != nil {
			fmt.Printf("- %s (%d) -> erreur tournois : %v\n", g.Name, g.Year, err)
			continue
		}
		fmt.Printf("- %s (%d) : %d tournoi(s)\n", g.Name, g.Year, len(tournaments))
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html>
<html lang="fr">
<head>
  <meta charset="utf-8">
  <title>GameDex</title>
  <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
  <h1>Bienvenue sur GameDex</h1>
  <p><a href="/jeux">Voir les jeux</a></p>
</body>
</html>`)
	})

	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		list, err := FetchGames(client, "/games")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		for _, g := range list {
			fmt.Fprintf(w, "#%d - %s (%d)\n", g.ID, g.Name, g.Year)
		}
	})

	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "ID invalide", http.StatusBadRequest)
			return
		}

		list, err := FetchGames(client, "/games")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		for _, g := range list {
			if g.ID == id {
				tournaments, _ := FetchTournaments(client, g.ID)
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				fmt.Fprintf(w, "%s (%d) - %d tournoi(s)\n", g.Name, g.Year, len(tournaments))
				return
			}
		}

		http.Error(w, "Jeu introuvable", http.StatusNotFound)
	})

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	addr := ":" + *port
	log.Printf("Serveur GameDex démarré sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
