package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/microhod/randflix-api/model/title"
)

// TitleHandler handles requests on the CRUD title endpoint
func (a *API) TitleHandler(w http.ResponseWriter, req *http.Request) {
	switch (req.Method) {
	case http.MethodPost:
		a.createTitle(w, req)
	case http.MethodPut:
		a.updateTitle(w, req)
	case http.MethodGet:
		if mux.Vars(req)["id"] != "" {
			a.getTitle(w, req)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) getTitle(w http.ResponseWriter, req *http.Request) {
	id := (mux.Vars(req))["id"]
	if id == "" {
		http.Error(w, "no title id supplied in url", http.StatusBadRequest)
		return
	}

	title, err := a.Storage.GetTitle(id)
	if err != nil {
		log.Printf("ERROR: failed to get title from storage: %s", err)
		http.Error(w, fmt.Sprintf("failed to get title to storage: %s", err), http.StatusInternalServerError)
		return
	}
	if title == nil {
		http.Error(w, fmt.Sprintf("no title with id: '%s'", id), http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(title)
	if err != nil {
		log.Printf("ERROR: could not serialise title: %s", err)
		http.Error(w, "could not serialise title", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(bytes))
}

func (a *API) createTitle(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	title := parseTitleFromBody(req)
	if title == nil {
		http.Error(w, "could not parse body to title", http.StatusBadRequest)
		return
	}

	t, err := a.Storage.GetTitle(title.ID)
	if err != nil {
		log.Printf("ERROR: failed to get title from storage: %s", err)
		http.Error(w, fmt.Sprintf("failed to get title to storage: %s", err), http.StatusInternalServerError)
		return
	}
	if t != nil {
		http.Error(w, fmt.Sprintf("title with id '%s' already exists", title.ID), http.StatusConflict)
		return
	}


	t, err = a.Storage.AddTitle(title)
	if err != nil {
		log.Printf("ERROR: failed to add title to storage: %s", err)
		http.Error(w, fmt.Sprintf("failed to add title to storage: %s", err), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(t)
	if err != nil {
		log.Printf("ERROR: could not serialise title: %s", err)
		http.Error(w, "could not serialise title", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, string(bytes))
}

func (a *API) updateTitle(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	id := (mux.Vars(req))["id"]
	if id == "" {
		http.Error(w, "no title id supplied in url", http.StatusBadRequest)
		return
	}

	t, err := a.Storage.GetTitle(id)
	if err != nil {
		log.Printf("ERROR: failed to get title from storage: %s", err)
		http.Error(w, fmt.Sprintf("failed to get title to storage: %s", err), http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, fmt.Sprintf("title with id '%s' does not exist", id), http.StatusNotFound)
		return
	}

	title := parseTitleFromBody(req)
	if title == nil {
		http.Error(w, "could not parse body to title", http.StatusBadRequest)
		return
	}
	if id != title.ID {
		http.Error(w, fmt.Sprintf("id mismatch between body (%s) and url (%s)", title.ID, id), http.StatusBadRequest)
		return
	}

	_, err = a.Storage.UpdateTitle(title)
	if err != nil {
		log.Printf("ERROR: failed to update title in storage: %s", err)
		http.Error(w, fmt.Sprintf("failed to update title in storage: %s", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func parseTitleFromBody(req *http.Request) *title.Title {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("ERROR: could not read request body: %s", err)
		return nil
	}

	var title title.Title
	err = json.Unmarshal(body, &title)
	if err != nil {
		log.Printf("ERROR: could not parse body to title: %s", err)
		return nil
	}

	return &title
}