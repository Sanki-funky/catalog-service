package main

import (
	"log"
	"net/http"
	"time"

	adapter "github.com/Sanky-funky/catalog-service/internal/adapters"
)

func main() {

	server := &http.Server{
		Addr:              ":8080",
		Handler:           adapter.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Println("Server listening on :8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
