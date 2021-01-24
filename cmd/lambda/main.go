package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/microhod/randflix-api/cmd"

	"github.com/akrylysov/algnhsa"
)

func main() {
	cfg, store, api, err := cmd.InitAPI()
	if err != nil {
		log.Fatal(err)
	}
	defer store.Disconnect()

	r := mux.NewRouter()
	r = api.ConfigureRandomRoutes(r)

	var handler http.Handler
	handler = api.ConfigureRandomHandler(cfg, r)

	log.Print("(http): starting http server")
	// start lambda
	algnhsa.ListenAndServe(handler, nil)
}
