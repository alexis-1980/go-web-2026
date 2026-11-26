package server

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"gamedex/internal/catalog"
)

type GamesPageData struct {
	Title  string
	Query  string
	Genre  string
	Genres []string
	Games  []catalog.Game
}

type GameDetailPageData struct {
	Title       string
	Game        catalog.Game
	Tournaments []catalog.Tournament
}

type ErrorPageData struct {
	Title   string
	Code    int
	Message string
}

type Server struct {
	store *catalog.Store
	tmpl  *template.Template
	mux   *http.ServeMux
}

func New(store *catalog.Store, webDir string) (*Server, error) {
	tmpl, err := template.ParseGlob(filepath.Join(webDir, "templates", "*.html"))
	if err != nil {
		return nil, fmt.Errorf("chargement templates impossible : %w", err)
	}

	s := &Server{
		store: store,
		tmpl:  tmpl,
		mux:   http.NewServeMux(),
	}
	s.routes(filepath.Join(webDir, "static"))
	return s, nil
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) renderError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	data := ErrorPageData{
		Title:   fmt.Sprintf("Erreur %d", code),
		Code:    code,
		Message: msg,
	}
	if err := s.tmpl.ExecuteTemplate(w, "erreur.html", data); err != nil {
		http.Error(w, msg, code)
	}
}

func (s *Server) routes(staticDir string) {
	s.mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
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

	s.mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		genre := r.URL.Query().Get("genre")
		games := s.store.SearchGames(q, genre)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := GamesPageData{
			Title:  "Jeux",
			Query:  q,
			Genre:  genre,
			Genres: s.store.Genres(),
			Games:  games,
		}
		if err := s.tmpl.ExecuteTemplate(w, "jeux.html", data); err != nil {
			s.renderError(w, http.StatusInternalServerError, "Erreur d'affichage de la page.")
		}
	})

	s.mux.HandleFunc("POST /recherche", func(w http.ResponseWriter, r *http.Request) {
		q := r.FormValue("q")
		genre := r.FormValue("genre")
		target := "/jeux?q=" + url.QueryEscape(q) + "&genre=" + url.QueryEscape(genre)
		http.Redirect(w, r, target, http.StatusSeeOther)
	})

	s.mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil || id <= 0 {
			s.renderError(w, http.StatusBadRequest, "L'identifiant du jeu doit être un entier positif.")
			return
		}

		game, ok := s.store.GetGame(id)
		if !ok {
			s.renderError(w, http.StatusNotFound, "Aucun jeu trouvé avec cet identifiant.")
			return
		}

		tournaments := s.store.GetTournaments(game.ID)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := GameDetailPageData{
			Title:       game.Name,
			Game:        game,
			Tournaments: tournaments,
		}
		if err := s.tmpl.ExecuteTemplate(w, "jeu_detail.html", data); err != nil {
			s.renderError(w, http.StatusInternalServerError, "Erreur d'affichage de la fiche jeu.")
		}
	})

	fileServer := http.FileServer(http.Dir(staticDir))
	s.mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
}
