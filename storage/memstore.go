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

	return s, nil
}

// AddTitle adds the title to storage
func (m *MemStore) AddTitle(t *title.Title) (*title.Title, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.titles[t.ID] != nil {
		return nil, fmt.Errorf("title already exists with id: '%s'", t.ID)
	}

	m.titles[t.ID] = t
	return m.titles[t.ID], nil
}

// UpdateTitle replaces the title in storage
func (m *MemStore) UpdateTitle(t *title.Title) (*title.Title, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.titles[t.ID] == nil {
		return nil, fmt.Errorf("title does not exist with id: '%s'", t.ID)
	}

	m.titles[t.ID] = t
	return m.titles[t.ID], nil
}

// GetTitle retrieves a title from storage by id
func (m *MemStore) GetTitle(id string) (*title.Title, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.titles[id], nil
}

// RandomTitle chooese a random title from storage (filtered by the filters)
func (m *MemStore) RandomTitle(filters ...title.Filter) (*title.Title, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	var msFilters []memStoreFilter

	for _, tf := range filters {
		if f, err := parseFilter(tf); err != nil {
			return nil, fmt.Errorf("Error parsing filter: %s", err)
		} else {
			msFilters = append(msFilters, f)
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
