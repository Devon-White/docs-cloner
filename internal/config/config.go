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
	Verbose     bool
	UserAgent   string
}
