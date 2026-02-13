package writer

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Frontmatter returns a YAML frontmatter block for a markdown file.
func Frontmatter(title string, sourceURL string, crawlDate time.Time) string {
	safeTitle := escapeYAML(title)
	return fmt.Sprintf("---\ntitle: %s\nsource_url: %s\ncrawl_date: %s\n---\n\n",
		safeTitle,
		sourceURL,
		crawlDate.Format(time.RFC3339),
	)
}

// WriteMarkdown writes a markdown file to disk, mirroring the URL path structure.
func WriteMarkdown(outputDir string, sourceURL string, title string, markdown string) error {
	path, err := URLToFilePath(outputDir, sourceURL)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", path, err)
	}

	if err := os.WriteFile(path, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

// PageResult holds a processed page for single-file concatenation.
type PageResult struct {
	URL      string
	Title    string
	Markdown string
}

// WriteSingleFile concatenates all pages into a single all-pages.md with a TOC.
func WriteSingleFile(outputDir string, pages []PageResult) error {
	var sb strings.Builder

	// Table of contents
	sb.WriteString("# Documentation Index\n\n")
	for i, p := range pages {
		anchor := slugify(p.Title)
		if anchor == "" {
			anchor = fmt.Sprintf("page-%d", i+1)
		}
		title := p.Title
		if title == "" {
			title = p.URL
		}
		sb.WriteString(fmt.Sprintf("- [%s](#%s)\n", title, anchor))
	}
	sb.WriteString("\n---\n\n")

	// Pages
	for _, p := range pages {
		title := p.Title
		if title == "" {
			title = p.URL
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n", title))
		sb.WriteString(fmt.Sprintf("*Source: %s*\n\n", p.URL))
		sb.WriteString(p.Markdown)
		sb.WriteString("\n\n---\n\n")
	}

	path := filepath.Join(outputDir, "all-pages.md")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// URLToFilePath converts a page URL to a filesystem path under outputDir.
func URLToFilePath(outputDir string, rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL %q: %w", rawURL, err)
	}

	p := strings.TrimPrefix(u.Path, "/")
	if p == "" || strings.HasSuffix(p, "/") {
		p += "index"
	}

	// Strip known extensions
	for _, ext := range []string{".html", ".htm", ".php"} {
		p = strings.TrimSuffix(p, ext)
	}

	p += ".md"

	return filepath.Join(outputDir, p), nil
}

// escapeYAML wraps a string in double quotes if it contains characters that
// could break YAML parsing.
func escapeYAML(s string) string {
	if s == "" {
		return `""`
	}
	needsQuoting := strings.ContainsAny(s, `:#"'{}[]|>&*!%@`+"`") ||
		strings.HasPrefix(s, " ") ||
		strings.HasPrefix(s, "-")
	if needsQuoting {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return s
}

// slugify creates a markdown-compatible anchor from a heading string.
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' {
			return r
		}
		return -1
	}, s)
	s = strings.ReplaceAll(s, " ", "-")
	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
