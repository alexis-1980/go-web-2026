package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func main() {
	port := flag.String("port", "8080", "port d'écoute du serveur HTTP")
	flag.Parse()

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
  <p>Catalogue de jeux vidéo et de leurs tournois.</p>
</body>
</html>`)
	})

	mux.HandleFunc("GET /jeux", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Liste des jeux GameDex")
	})

	mux.HandleFunc("GET /jeux/{id}", func(w http.ResponseWriter, r *http.Request) {
		rawID := r.PathValue("id")
		id, err := strconv.Atoi(rawID)
		if err != nil {
			http.Error(w, "ID invalide : un entier est attendu", http.StatusBadRequest) // 400
			return
		}

		if id < 1 || id > 10 {
			http.Error(w, "Jeu introuvable", http.StatusNotFound) // 404
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Détail du jeu #%d\n", id)
	})

	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))

	addr := ":" + *port
	log.Printf("Serveur GameDex démarré sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
