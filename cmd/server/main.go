package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewMemStorage()
	http.HandleFunc("/update/", MetricUpdateHandler(store))

	log.Println("Server start listening on", defaultServerAddress)

	http.ListenAndServe(defaultServerAddress, nil)
}
