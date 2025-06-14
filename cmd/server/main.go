package main

import (
	"log"
	"net/http"

	"github.com/etoneja/go-metrics/internal/server"
)

func main() {
	cfg := server.PrepareConfig()

	store := server.NewMemStorage()
	router := server.NewRouter(store)

	log.Println("Server start listening on", cfg.ServerAddress)

	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		log.Fatalf("Сould not start server: %v", err)
	}

}
