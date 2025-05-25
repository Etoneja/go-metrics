package main

import (
	"log"
	"net/http"
)

func main() {
	cfg := prepareConfig()

	store := NewMemStorage()
	router := NewRouter(store)

	log.Println("Server start listening on", cfg.ServerAddress)
	
	err := http.ListenAndServe(cfg.ServerAddress, router)
	if err != nil {
		log.Fatalf("Сould not start server: %v", err)
	}

}
