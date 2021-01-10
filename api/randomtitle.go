package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/microhod/randflix-api/model/title"
)

const (
	defaultScoreKind = "metacritic"
)

type titleQuery struct {
	service string
	genres  []string
	score   scoreQuery
}

type scoreQuery struct {
	kind string
	min  int
	max  int
}

// RandomTitleHandler handles requests for a random title
func (a *API) RandomTitleHandler(w http.ResponseWriter, req *http.Request) {

	q, err := parseTitleQuery(req.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title, err := a.Storage.RandomTitle(
		title.OnServiceFilter{Service: q.service},
		title.IsGenreFilter{Genres: q.genres},
		title.ScoreBetweenFilter{Kind: q.score.kind, Min: q.score.min, Max: q.score.max},
	)

	if err != nil {
		log.Printf("ERROR: Failed to get random title from storage: %s", err)
		http.Error(w, "Failed to get random title from storage", http.StatusInternalServerError)
		return
	}

	if title == nil {
		http.Error(w, "No matching title found", http.StatusNotFound)
		return
	}

	bytes, err := json.Marshal(title)
	if err != nil {
		log.Printf("ERROR: Could not serialise title: %s", err)
		http.Error(w, "Could not serialise title", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, string(bytes))
	return
}

func parseTitleQuery(query map[string][]string) (*titleQuery, error) {

	tq := &titleQuery{}
	var err error

	// Service
	keys, ok := query["service"]
	if ok && len(keys) > 0 {
		tq.service = keys[0]
	}

	// Genres
	tq.genres, _ = query["genres"]

	// Score
	keys, ok = query["score_kind"]
	if ok && len(keys) > 0 {
		tq.score.kind = keys[0]
	} else {
		tq.score.kind = defaultScoreKind
	}
	keys, ok = query["score_min"]
	if ok && len(keys) > 0 {
		tq.score.min, err = strconv.Atoi(keys[0])
		if err != nil {
			return nil, fmt.Errorf("score_min query parameter must be an integer")
		}
	}
	keys, ok = query["score_max"]
	if ok && len(keys) > 0 {
		tq.score.max, err = strconv.Atoi(keys[0])
		if err != nil {
			return nil, fmt.Errorf("score_max query parameter must be an integer")
		}
	}

	return tq, nil
}