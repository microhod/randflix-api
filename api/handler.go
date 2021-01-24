package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/microhod/randflix-api/config"
	"github.com/rs/cors"
)

// ConfigureRandomRoutes adds routes for the random api to the router
func (a API) ConfigureRandomRoutes(r *mux.Router) *mux.Router {
	r.HandleFunc("/title/random", a.RandomTitleHandler).
		Methods(http.MethodGet).
		Schemes("http")

	return r
}

// ConfigureRandomHandler adds middleware for the random api to the handler
func (a API) ConfigureRandomHandler(c *config.Config, h http.Handler) http.Handler {
	cors := cors.New(cors.Options{
		AllowedOrigins: c.AllowedOrigins,
		AllowedHeaders: c.AllowedHeaders,
		AllowedMethods: []string{http.MethodGet, http.MethodOptions},
	})

	return cors.Handler(h)
}

// ConfigureTitleRoutes adds routes for the title api to the router
func (a API) ConfigureTitleRoutes(r *mux.Router) *mux.Router {
	r.HandleFunc("/title", a.TitleHandler).
		Methods(http.MethodPost, http.MethodGet).
		Schemes("http")
	r.HandleFunc("/title/{id}", a.TitleHandler).
		Methods(http.MethodGet, http.MethodPut).
		Schemes("http")

	return r
}

// ConfigureTitleHandler adds middleware for the title api to the handler
func (a API) ConfigureTitleHandler(c *config.Config, h http.Handler) http.Handler {
	return h
}
