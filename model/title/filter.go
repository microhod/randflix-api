package title

// Filter is a generic filter interface
type Filter interface {}

// OnServiceFilter checks if the title is on the specified service e.g. Netflix
// if service is an empty string, it has no effect
type OnServiceFilter struct {
	Service string
}

// IsGenreFilter checks if the title is of the specified genres (case insensitive)
type IsGenreFilter struct {
	Genres []string
}

// ScoreBetweenFilter checks if the title has the specified score in the specified range
type ScoreBetweenFilter struct {
	Kind string
	Min  int
	Max  int
}
