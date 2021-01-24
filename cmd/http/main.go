package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/microhod/randflix-api/cmd"
)

func main() {

	cfg, store, api, err := cmd.InitAPI()
	if err != nil {
		log.Fatal(err)
	}
	defer store.Disconnect()

	r := mux.NewRouter()
	r = api.ConfigureRandomRoutes(r)
	r = api.ConfigureTitleRoutes(r)

	var handler http.Handler
	handler = api.ConfigureRandomHandler(cfg, r)
	handler = api.ConfigureTitleHandler(cfg, r)

	log.Printf("(http): starting http server on port %d", cfg.Port)
	http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Port), handler)
}
