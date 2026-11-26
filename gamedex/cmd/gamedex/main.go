package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"gamedex/internal/api"
	"gamedex/internal/catalog"
	"gamedex/internal/server"
)

const apiBaseURL = "http://localhost:8081/api"

func main() {
	port := flag.String("port", "8080", "port d'écoute du serveur HTTP")
	flag.Parse()

	client := api.NewClient(apiBaseURL, 2*time.Second)

	games, err := client.FetchGames("/games")
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	tournamentsByGame := make(map[int][]catalog.Tournament, len(games))
	for _, g := range games {
		tournaments, err := client.FetchTournaments(g.ID)
		if err != nil {
			fmt.Printf("Erreur tournois pour %s (%d) : %v\n", g.Name, g.Year, err)
			return
		}
		tournamentsByGame[g.ID] = tournaments
	}

	store := catalog.NewStore(games, tournamentsByGame)
	store.PrintSummary()

	srv, err := server.New(store, "web")
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + *port
	log.Printf("Serveur GameDex démarré sur http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
