package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewMemStorage()
	router := NewRouter(store)

	log.Println("Server start listening on", defaultServerAddress)

	err := http.ListenAndServe(defaultServerAddress, router)
	if err != nil {
		log.Fatalf("Сould not start server: %v", err)
	}

}
