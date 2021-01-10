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

	cfg := config.GetConfig()

	store, err := storage.CreateStorage(cfg)
	if err != nil {
		log.Panic(fmt.Errorf("Failed to create storage: %s", err))
	}

	a := api.API{Storage: store}

	r := mux.NewRouter()
	r.HandleFunc("/title/random", a.RandomTitleHandler).
		Methods(http.MethodGet).
		Schemes("http")
	r.HandleFunc("/title", a.TitleHandler).
		Methods(http.MethodPost, http.MethodGet).
		Schemes("http")
	r.HandleFunc("/title/{id}", a.TitleHandler).
		Methods(http.MethodGet, http.MethodPut).
		Schemes("http")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedHeaders: []string{"Content-Type"},
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	})

	handler := c.Handler(r)

	log.Print("(http): starting http server...")
	http.ListenAndServe("0.0.0.0:8080", handler)
}
