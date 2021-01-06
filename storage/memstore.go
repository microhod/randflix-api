package storage

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"sync"

	"github.com/microhod/randflix-api/model/title"
)

// MemStore is in-memory storage
type MemStore struct {
	lock   sync.RWMutex
	titles map[string]*title.Title
}

type memStoreFilter func(*title.Title) bool

// NewMemStore creates a new empty MemStore
func (*Config) NewMemStore() (Storage, error) {
	s := &MemStore{
		titles: map[string]*title.Title{},
	}

	// Useful for debugging
	// TODO: replace this with a better idea
	s.Populate()

	return s, nil
}

// Populate initializes the storage with some dummy data for testing purposes
func (m *MemStore) Populate() {
	m.titles = map[string]*title.Title{
		"tt0831387": {
			ID:          "tt0831387",
			Name:        "Godzilla",
			Description: "The world is beset by the appearance of monstrous creatures, but one of them may be the only one who can save humanity.",
			Genres: []string{
				"action", "adventure", "sci-fi",
			},
			Scores: map[string]int{
				"metacritic": 62,
			},
			Poster: "https://m.media-amazon.com/images/M/MV5BN2E4ZDgxN2YtZjExMS00MWE5LTg3NjQtNTkxMzJhOTA3MDQ4XkEyXkFqcGdeQXVyMTQxNzMzNDI@.png",
			Directories: map[string]*title.Directory{
				"imdb": {
					URL: "https://imdb.com/title/tt0831387",
				},
			},
			Services: map[string]*title.Service{
				"netflix": {
					URL: "https://netflix.com/watch/11819467",
				},
			},
		},
		"tt4633694": {
			ID:          "tt4633694",
			Name:        "Spider-Man: Into the Spider-Verse",
			Description: "Teen Miles Morales becomes the Spider-Man of his universe, and must join with five spider-powered individuals from other dimensions to stop a threat for all realities.",
			Genres: []string{
				"action", "adventure", "animation",
			},
			Scores: map[string]int{
				"metacritic": 87,
			},
			Poster: "https://m.media-amazon.com/images/M/MV5BMjMwNDkxMTgzOF5BMl5BanBnXkFtZTgwNTkwNTQ3NjM@.png",
			Directories: map[string]*title.Directory{
				"imdb": {
					URL: "https://imdb.com/title/tt4633694",
				},
			},
			Services: map[string]*title.Service{
				"prime": {
					URL: "https://www.amazon.co.uk/gp/video/detail/amzn1.dv.gti.b2b3cd4f-7b33-5301-03d7-10160df79fbd",
				},
			},
		},
		"tt10618286": {
			ID:          "tt10618286",
			Name:        "Mank",
			Description: "1930's Hollywood is reevaluated through the eyes of scathing social critic and alcoholic screenwriter Herman J. Mankiewicz as he races to finish the screenplay of Citizen Kane (1941).",
			Genres: []string{
				"biography", "comedy", "drama",
			},
			Scores: map[string]int{
				"metacritic": 79,
			},
			Directories: map[string]*title.Directory{
				"imdb": {
					URL: "https://imdb.com/title/tt10618286",
				},
			},
		},
	}
}

// AddTitle adds the title to storage
func (m *MemStore) AddTitle(t title.Title) (*title.Title, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.titles[t.ID] != nil {
		return nil, fmt.Errorf("title already exists with id: '%s'", t.ID)
	}

	m.titles[t.ID] = &t
	return m.titles[t.ID], nil
}

// RandomTitle chooese a random title from storage (filtered by the filters)
func (m *MemStore) RandomTitle(filters ...title.Filter) (*title.Title, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	var msFilters []memStoreFilter

	for _, tf := range filters {
		if c, err := parseFilter(tf); err != nil {
			return nil, fmt.Errorf("Error parsing filter: %s", err)
		} else {
			msFilters = append(msFilters, c)
		}
	}

	list := []*title.Title{}

	for _, t := range m.titles {
		if passes(t, msFilters) {
			list = append(list, t)
		}
	}

	if len(list) == 0 {
		return nil, nil
	}

	return list[rand.Intn(len(list))], nil
}

func passes(t *title.Title, filters []memStoreFilter) bool {
	for _, filter := range filters {
		if !filter(t) {
			return false
		}
	}
	return true
}

func parseFilter(tf title.Filter) (memStoreFilter, error) {
	var filter memStoreFilter

	switch tf.(type) {
	case title.OnServiceFilter:
		filter = onService(tf.(title.OnServiceFilter).Service)
	case title.IsGenreFilter:
		filter = isGenre(tf.(title.IsGenreFilter).Genres...)
	case title.ScoreBetweenFilter:
		f := tf.(title.ScoreBetweenFilter)
		filter = scoreBetween(f.Kind, f.Min, f.Max)
	default:
		return nil, fmt.Errorf("Unsupported title filter type: %s", reflect.TypeOf(tf))
	}

	return filter, nil
}

func onService(name string) memStoreFilter {
	if name == "" {
		return truefilter
	}
	return func(t *title.Title) bool {
		return t != nil && t.Services != nil && t.Services[name] != nil && t.Services[name].URL != ""
	}
}

func isGenre(names ...string) memStoreFilter {
	return func(t *title.Title) bool {
		for _, n := range names {
			if !containsCaseInsensitive(t.Genres, n) {
				return false
			}
		}
		return true
	}
}

func scoreBetween(kind string, min int, max int) memStoreFilter {
	if kind == "" {
		return truefilter
	}
	if max == 0 {
		max = math.MaxInt64
	}
	return func(t *title.Title) bool {

		return t.Scores[kind] >= min && t.Scores[kind] <= max
	}
}

func truefilter(*title.Title) bool {
	return true
}

func containsCaseInsensitive(items []string, term string) bool {
	for _, i := range items {
		if strings.ToLower(i) == strings.ToLower(term) {
			return true
		}
	}
	return false
}
