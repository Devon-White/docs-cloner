package config

// Config holds all CLI options for a docs-cloner run.
type Config struct {
	SitemapURL  string
	OutputDir   string
	FetchMD     string // URL pattern with {url}/{path}/{host} placeholders; empty = HTML-to-MD mode
	Concurrency int
	DelayMS     int
	SingleFile  bool
	Selector    string // CSS selector for main content; empty = heuristic
	Include     []string // URL must contain at least one of these substrings
	Exclude     []string // URL must not contain any of these substrings
	Clean       bool
	Verbose     bool
	UserAgent   string
}
