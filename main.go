package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/microhod/randflix-api/api"
	"github.com/microhod/randflix-api/config"
	"github.com/microhod/randflix-api/storage"
	"github.com/rs/cors"
)

func main() {

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("failed to get config: %s\n", err)
	}
	log.Printf("config: %s\n", cfg)

	store, err := storage.CreateStorage(cfg)
	if err != nil {
		log.Fatalf("failed to create storage: %s\n", err)
	}

	api := api.API{Storage: store}
	defer store.Disconnect()

	r := mux.NewRouter()
	r.HandleFunc("/title/random", api.RandomTitleHandler).
		Methods(http.MethodGet).
		Schemes("http")
	r.HandleFunc("/title", api.TitleHandler).
		Methods(http.MethodPost, http.MethodGet).
		Schemes("http")
	r.HandleFunc("/title/{id}", api.TitleHandler).
		Methods(http.MethodGet, http.MethodPut).
		Schemes("http")

	cors := cors.New(cors.Options{
		AllowedOrigins: cfg.CorsAllowedOrigins,
		AllowedHeaders: cfg.CorsAllowedHeaders,
		AllowedMethods: cfg.CorsAllowedMethods,
	})

	handler := cors.Handler(r)

	log.Printf("(http): starting http server on port %d", cfg.Port)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Port), handler)
}
