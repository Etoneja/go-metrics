package main

import (
	"log"
	"net/http"
)

func main() {
	store := NewMemStorage()
	http.HandleFunc("/update/", MetricUpdateHandler(store))

	log.Println("Server start listening on", defaultServerAddress)

	err := http.ListenAndServe(defaultServerAddress, nil)
	if err != nil {
		log.Fatalf("Сould not start server: %v", err)
	}

}
