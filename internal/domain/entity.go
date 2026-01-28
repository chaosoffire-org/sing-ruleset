// Package domain contains the core business entities and interfaces.
package domain

// Source represents a rule source with its name, URL, and type.
type Source struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}

// Config represents the application configuration containing version and sources.
type Config struct {
	Version string              `json:"version"`
	Sources map[string][]Source `json:"sources"`
}
