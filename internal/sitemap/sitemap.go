package sitemap

import "encoding/xml"

// URLSet represents a standard sitemap <urlset>.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []URL    `xml:"url"`
}

// URL is a single <url> entry in a sitemap.
type URL struct {
	Loc string `xml:"loc"`
}

// SitemapIndex represents a <sitemapindex> that links to sub-sitemaps.
type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

// Sitemap is a single <sitemap> entry in a sitemap index.
type Sitemap struct {
	Loc string `xml:"loc"`
}

// ParseResult holds the output of parsing a sitemap document.
type ParseResult struct {
	PageURLs    []string // page URLs from a <urlset>
	SubSitemaps []string // sub-sitemap URLs from a <sitemapindex>
}

// Parse parses raw XML bytes as either a sitemap index or a urlset.
// It returns page URLs and/or sub-sitemap URLs depending on the document type.
func Parse(data []byte) (*ParseResult, error) {
	// Try sitemap index first
	var idx SitemapIndex
	if err := xml.Unmarshal(data, &idx); err == nil && idx.XMLName.Local == "sitemapindex" {
		result := &ParseResult{}
		for _, s := range idx.Sitemaps {
			if s.Loc != "" {
				result.SubSitemaps = append(result.SubSitemaps, s.Loc)
			}
		}
		return result, nil
	}

	// Otherwise parse as urlset
	var urlset URLSet
	if err := xml.Unmarshal(data, &urlset); err != nil {
		return nil, err
	}

	result := &ParseResult{}
	for _, u := range urlset.URLs {
		if u.Loc != "" {
			result.PageURLs = append(result.PageURLs, u.Loc)
		}
	}
	return result, nil
}
