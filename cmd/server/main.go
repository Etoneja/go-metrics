package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewMemStorage()
	router := NewRouter(store)

	log.Println("Server start listening on", defaultServerAddress)

	http.ListenAndServe(defaultServerAddress, router)
}
