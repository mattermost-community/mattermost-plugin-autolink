package main

// Link represents a pattern to autolink
type Link struct {
	Pattern              string
	Template             string
	Scope                []string
	DisableNonWordPrefix bool
	DisableNonWordSuffix bool
}

// Configuration from config.json
type Configuration struct {
	Links []*Link
}
