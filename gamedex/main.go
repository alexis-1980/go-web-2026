package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const apiBase = "http://localhost:8081/api"

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

type GamesPageData struct {
	Title string
	Query string
	Games []Game
}

type GameDetailPageData struct {
	Title       string
	Game        Game
	Tournaments []Tournament
}

type ErrorPageData struct {
	Title   string
	Code    int
	Message string
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

func renderError(w http.ResponseWriter, tmpl *template.Template, code int, msg string) {
	w.WriteHeader(code)
	data := ErrorPageData{
		Title:   fmt.Sprintf("Erreur %d", code),
		Code:    code,
		Message: msg,
	}
	if err := tmpl.ExecuteTemplate(w, "erreur.html", data); err != nil {
		http.Error(w, msg, code)
	}
}

func main() {
	port := flag.String("port", "8080", "port d'écoute du serveur HTTP")
	flag.Parse()

	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatal("Erreur chargement templates :", err)
	}

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
		q := r.URL.Query().Get("q")
		list, err := FetchGames(client, "/games")
		if err != nil {
			renderError(w, tmpl, http.StatusBadGateway, "L'API des jeux est indisponible pour le moment.")
			return
		}

		var filtered []Game
		qLower := strings.ToLower(strings.TrimSpace(q))
		for _, g := range list {
			if qLower == "" || strings.Contains(strings.ToLower(g.Name), qLower) {
				filtered = append(filtered, g)
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := GamesPageData{
			Title: "Jeux",
			Query: q,
			Games: filtered,
		}
		if err := tmpl.ExecuteTemplate(w, "jeux.html", data); err != nil {
			renderError(w, tmpl, http.StatusInternalServerError, "Erreur d'affichage de la page.")
		}
	})

	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id <= 0 {
			renderError(w, tmpl, http.StatusBadRequest, "L'identifiant du jeu doit être un entier positif.")
			return
		}

		list, err := FetchGames(client, "/games")
		if err != nil {
			renderError(w, tmpl, http.StatusBadGateway, "L'API des jeux est indisponible pour le moment.")
			return
		}

		var found *Game
		for i := range list {
			if list[i].ID == id {
				found = &list[i]
				break
			}
		}
		if found == nil {
			renderError(w, tmpl, http.StatusNotFound, "Aucun jeu trouvé avec cet identifiant.")
			return
		}

		tournaments, err := FetchTournaments(client, found.ID)
		if err != nil {
			renderError(w, tmpl, http.StatusBadGateway, "Impossible de récupérer les tournois depuis l'API.")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := GameDetailPageData{
			Title:       found.Name,
			Game:        *found,
			Tournaments: tournaments,
		}
		if err := tmpl.ExecuteTemplate(w, "jeu_detail.html", data); err != nil {
			renderError(w, tmpl, http.StatusInternalServerError, "Erreur d'affichage de la fiche jeu.")
		}
	})

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	addr := ":" + *port
	log.Printf("Serveur GameDex démarré sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
