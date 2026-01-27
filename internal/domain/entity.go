package domain

type Source struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

type Config struct {
	Version string              `json:"version"`
	Sources map[string][]Source `json:"sources"`
}
