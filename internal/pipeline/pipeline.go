package pipeline

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Devon-White/docs-cloner/internal/config"
	"github.com/Devon-White/docs-cloner/internal/converter"
	"github.com/Devon-White/docs-cloner/internal/extractor"
	"github.com/Devon-White/docs-cloner/internal/fetcher"
	"github.com/Devon-White/docs-cloner/internal/sitemap"
	"github.com/Devon-White/docs-cloner/internal/writer"
)

type pageResult struct {
	URL      string
	Title    string
	Markdown string
	Err      error
}

// Run executes the full docs-cloner pipeline: fetch sitemap, process pages
// concurrently, and write markdown files to disk.
func Run(ctx context.Context, cfg *config.Config) error {
	f := fetcher.New(cfg.UserAgent, cfg.DelayMS)

	// Fetch and resolve sitemap (including sitemap index recursion)
	log.Printf("Fetching sitemap: %s", cfg.SitemapURL)
	urls, err := fetchSitemapURLs(ctx, f, cfg.SitemapURL)
	if err != nil {
		return fmt.Errorf("sitemap: %w", err)
	}
	log.Printf("Found %d URLs in sitemap", len(urls))

	// Filter URLs by include/exclude patterns
	if len(cfg.Include) > 0 || len(cfg.Exclude) > 0 {
		filtered := urls[:0]
		for _, u := range urls {
			if matchesFilter(u, cfg.Include, cfg.Exclude) {
				filtered = append(filtered, u)
			}
		}
		log.Printf("Filtered to %d URLs (from %d)", len(filtered), len(urls))
		urls = filtered

		if len(urls) == 0 && runtime.GOOS == "windows" {
			log.Println("Hint: Git Bash rewrites args starting with \"/\" into Windows paths.")
			log.Println("      Use --include docs/en/ (no leading /) or set MSYS_NO_PATHCONV=1")
		}
	}

	if len(urls) == 0 {
		log.Println("No URLs found after filtering. Nothing to do.")
		return nil
	}

	// Clean output directory if requested
	if cfg.Clean {
		log.Printf("Cleaning output directory: %s", cfg.OutputDir)
		if err := os.RemoveAll(cfg.OutputDir); err != nil {
			return fmt.Errorf("cleaning output directory: %w", err)
		}
	}

	// Fan-out: send URLs to workers
	urlCh := make(chan string, len(urls))
	for _, u := range urls {
		urlCh <- u
	}
	close(urlCh)

	// Workers produce results
	resultCh := make(chan pageResult, cfg.Concurrency*2)

	var wg sync.WaitGroup
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for pageURL := range urlCh {
				select {
				case <-ctx.Done():
					return
				default:
				}
				result := processPage(ctx, f, cfg, pageURL)
				resultCh <- result
			}
		}(i)
	}

	// Close results when all workers finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results and write to disk
	var results []writer.PageResult
	var errCount int
	total := len(urls)
	done := 0

	for result := range resultCh {
		done++
		if result.Err != nil {
			errCount++
			log.Printf("[%d/%d] ERROR %s: %v", done, total, result.URL, result.Err)
			continue
		}

		if err := writer.WriteMarkdown(cfg.OutputDir, result.URL, result.Title, result.Markdown); err != nil {
			errCount++
			log.Printf("[%d/%d] WRITE ERROR %s: %v", done, total, result.URL, err)
			continue
		}

		if cfg.Verbose {
			log.Printf("[%d/%d] OK %s", done, total, result.URL)
		}

		results = append(results, writer.PageResult{
			URL:      result.URL,
			Title:    result.Title,
			Markdown: result.Markdown,
		})
	}

	// Single-file output
	if cfg.SingleFile && len(results) > 0 {
		log.Printf("Writing single file with %d pages...", len(results))
		if err := writer.WriteSingleFile(cfg.OutputDir, results); err != nil {
			return fmt.Errorf("single file: %w", err)
		}
	}

	log.Printf("Done. %d pages written, %d errors.", len(results), errCount)
	if len(results) == 0 && errCount > 0 {
		return fmt.Errorf("all %d pages failed", errCount)
	}
	return nil
}

// fetchSitemapURLs recursively fetches sitemap URLs, resolving sitemap indexes.
func fetchSitemapURLs(ctx context.Context, f *fetcher.Fetcher, sitemapURL string) ([]string, error) {
	body, err := f.Fetch(ctx, sitemapURL)
	if err != nil {
		return nil, err
	}

	result, err := sitemap.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", sitemapURL, err)
	}

	urls := result.PageURLs

	// Recurse into sub-sitemaps
	for _, subURL := range result.SubSitemaps {
		subURLs, err := fetchSitemapURLs(ctx, f, subURL)
		if err != nil {
			log.Printf("WARNING: sub-sitemap %s failed: %v", subURL, err)
			continue
		}
		urls = append(urls, subURLs...)
	}

	return urls, nil
}

// processPage fetches and converts a single page to markdown.
func processPage(ctx context.Context, f *fetcher.Fetcher, cfg *config.Config, pageURL string) pageResult {
	var markdown string
	var title string

	if cfg.FetchMD != "" {
		md, err := converter.FetchRawMD(f, ctx, pageURL, cfg.FetchMD)
		if err != nil {
			return pageResult{URL: pageURL, Err: err}
		}
		markdown = md
		title = converter.ExtractTitleFromMarkdown(md)
	} else {
		body, err := f.Fetch(ctx, pageURL)
		if err != nil {
			return pageResult{URL: pageURL, Err: err}
		}

		html, pageTitle, err := extractor.Extract(body, cfg.Selector, pageURL)
		if err != nil {
			return pageResult{URL: pageURL, Err: fmt.Errorf("extraction: %w", err)}
		}

		md, err := converter.ConvertHTML(html, pageURL)
		if err != nil {
			return pageResult{URL: pageURL, Err: fmt.Errorf("conversion: %w", err)}
		}

		markdown = md
		title = pageTitle
	}

	markdown = converter.CleanMarkdown(markdown)

	// Add frontmatter
	markdown = writer.Frontmatter(title, pageURL, time.Now()) + markdown

	return pageResult{
		URL:      pageURL,
		Title:    title,
		Markdown: markdown,
	}
}

// matchesFilter returns true if the URL passes include/exclude filters.
// If include is non-empty, the URL must contain at least one include substring.
// If exclude is non-empty, the URL must not contain any exclude substring.
func matchesFilter(url string, include, exclude []string) bool {
	if len(include) > 0 {
		matched := false
		for _, pattern := range include {
			if strings.Contains(url, pattern) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	for _, pattern := range exclude {
		if strings.Contains(url, pattern) {
			return false
		}
	}
	return true
}

