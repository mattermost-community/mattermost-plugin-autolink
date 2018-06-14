package main

// Link represents a pattern to autolink
type Link struct {
	Pattern              string
	Template             string
	DisableNonWordPrefix bool
	DisableNonWordSuffix bool
}

// Configuration from config.json
type Configuration struct {
	Links []*Link
}
