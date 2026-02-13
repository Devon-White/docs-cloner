package converter

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"

	"github.com/Devon-White/docs-cloner/internal/fetcher"
)

// removeTags are HTML tags that should be stripped entirely during conversion.
var removeTags = []string{
	"nav", "header", "footer", "aside", "script", "style", "noscript", "iframe",
}

var multiBlankLines = regexp.MustCompile(`\n{3,}`)

// ConvertHTML converts an extracted HTML fragment to markdown.
// sourceURL is used to resolve relative links to absolute.
func ConvertHTML(extractedHTML string, sourceURL string) (string, error) {
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)

	for _, tag := range removeTags {
		conv.Register.TagType(tag, converter.TagTypeRemove, converter.PriorityStandard)
	}

	md, err := conv.ConvertString(extractedHTML, converter.WithDomain(domainFromURL(sourceURL)))
	if err != nil {
		return "", fmt.Errorf("html-to-markdown conversion: %w", err)
	}

	return CleanMarkdown(md), nil
}

// FetchRawMD fetches raw markdown from a URL derived from the page URL using
// the given pattern. Supported placeholders: {url}, {path}, {host}.
func FetchRawMD(f *fetcher.Fetcher, ctx context.Context, pageURL string, pattern string) (string, error) {
	mdURL := expandPattern(pattern, pageURL)

	body, err := f.Fetch(ctx, mdURL)
	if err != nil {
		return "", fmt.Errorf("fetching raw markdown from %s: %w", mdURL, err)
	}

	return CleanMarkdown(string(body)), nil
}

// ExtractTitleFromMarkdown extracts the first level-1 heading from markdown.
func ExtractTitleFromMarkdown(md string) string {
	for _, line := range strings.Split(md, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return ""
}

// CleanMarkdown normalizes whitespace in markdown output.
func CleanMarkdown(md string) string {
	// Collapse 3+ blank lines to 2
	md = multiBlankLines.ReplaceAllString(md, "\n\n")

	// Trim trailing whitespace per line
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	md = strings.Join(lines, "\n")

	// Trim leading/trailing blank lines
	md = strings.TrimSpace(md)

	return md
}

func domainFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func expandPattern(pattern, pageURL string) string {
	u, _ := url.Parse(pageURL)
	r := strings.NewReplacer(
		"{url}", pageURL,
		"{path}", u.Path,
		"{host}", u.Host,
	)
	return r.Replace(pattern)
}
