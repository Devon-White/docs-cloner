package fetcher

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Fetcher wraps an HTTP client with rate-limiting, User-Agent, and gzip support.
type Fetcher struct {
	client    *http.Client
	userAgent string
	delay     time.Duration
}

// New creates a Fetcher with the given User-Agent and per-call delay.
func New(userAgent string, delayMS int) *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: userAgent,
		delay:     time.Duration(delayMS) * time.Millisecond,
	}
}

// Fetch retrieves the body of the given URL. It automatically decompresses
// gzip responses and URLs ending in .gz.
func (f *Fetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	time.Sleep(f.delay)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	var reader io.Reader = resp.Body

	// Decompress if gzip content-encoding or .gz URL
	if resp.Header.Get("Content-Encoding") == "gzip" || strings.HasSuffix(url, ".gz") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("decompressing gzip response from %s: %w", url, err)
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("reading body from %s: %w", url, err)
	}

	return body, nil
}
