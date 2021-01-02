package title

// Title describes an object representing a piece of entertainment e.g. Movie or a TV Show
type Title struct {
	Name        string 				  `json:"name"`
	Description string 				  `json:"description"`
	Genres      []string			  `json:"genres"`
	Scores      map[string]int  	  `json:"scores"`
	Poster	    string				  `json:"poster"`
	Directories map[string]*Directory `json:"directories"`
	Services    map[string]*Service   `json:"services"`
}

// Directory is a reference to a title in an external store such as IMDB
type Directory struct {
	URL string						 `json:"url"`
	AdditionalInfo map[string]string `json:"additionalInfo"`
}

// Service is a reference to a title in an external streming service such as Netflix
type Service struct {
	URL string						 `json:"url"`
	AdditionalInfo map[string]string `json:"additionalInfo"`
}
