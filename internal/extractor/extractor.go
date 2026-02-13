package extractor

import (
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// heuristicSelectors is the ordered list of CSS selectors tried when no
// explicit --selector is provided. The first match with meaningful text wins.
var heuristicSelectors = []string{
	"main",
	"article",
	"[role=\"main\"]",
	".content",
	".main-content",
	"#content",
	".markdown-body",
	".documentation-content",
	".docs-content",
	".page-content",
}

// noiseSelectors are elements removed from the content area before extraction.
var noiseSelectors = []string{
	"nav",
	".nav",
	".sidebar",
	".toc",
	".table-of-contents",
	".breadcrumb",
	".breadcrumbs",
	".pagination",
	".edit-page",
	".feedback",
	".header",
	".footer",
	"header",
	"footer",
	"script",
	"style",
	"noscript",
	"iframe",
	".ads",
	".cookie-banner",
}

// Extract parses the HTML body, isolates the main content area, removes noise,
// and returns the cleaned inner HTML and the page title.
func Extract(htmlBody []byte, selector string, sourceURL string) (contentHTML string, title string, err error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBody))
	if err != nil {
		return "", "", err
	}

	// Extract title
	title = strings.TrimSpace(doc.Find("title").First().Text())
	if title == "" {
		title = strings.TrimSpace(doc.Find("h1").First().Text())
	}

	// Select main content area
	var selection *goquery.Selection
	if selector != "" {
		selection = doc.Find(selector)
	} else {
		selection = findMainContent(doc)
	}

	// Remove noise elements
	combined := strings.Join(noiseSelectors, ", ")
	selection.Find(combined).Remove()

	html, err := selection.Html()
	if err != nil {
		return "", title, err
	}

	return html, title, nil
}

// findMainContent tries heuristic selectors in order and returns the first
// match with substantial text content (>50 chars). Falls back to body.
func findMainContent(doc *goquery.Document) *goquery.Selection {
	for _, sel := range heuristicSelectors {
		s := doc.Find(sel).First()
		if s.Length() > 0 && len(strings.TrimSpace(s.Text())) > 50 {
			return s
		}
	}
	return doc.Find("body")
}
