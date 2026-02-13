# docs-cloner

Clone entire documentation sites into AI-friendly markdown via their sitemap.

Built for pulling documentation into context for AI agents when the source repo isn't public.

## Install

```bash
go install github.com/devon/docs-cloner@latest
```

Or build from source:

```bash
git clone https://github.com/devon/docs-cloner.git
cd docs-cloner
go build -o docs-cloner .
```

## Usage

### Basic: clone a docs site to markdown

```bash
docs-cloner --url https://example.com/sitemap.xml -o ./docs
```

This fetches every page in the sitemap, extracts the main content area, converts it to clean markdown with YAML frontmatter, and writes files mirroring the site's URL structure.

### Fetch raw markdown instead of converting HTML

Some documentation sites serve raw `.md` files at alternate URLs. Use `--fetch-md` to skip HTML conversion entirely:

```bash
# Append .md to each page URL (default pattern)
docs-cloner --url https://example.com/sitemap.xml --fetch-md

# Custom pattern with placeholders
docs-cloner --url https://example.com/sitemap.xml --fetch-md "{url}?plain=1"
```

### Produce a single file for LLM context

```bash
docs-cloner --url https://example.com/sitemap.xml --single-file -o ./docs
```

This writes individual files *and* a concatenated `all-pages.md` with a table of contents at the top.

### Custom content selector

If the auto-detection picks up the wrong content area, specify a CSS selector:

```bash
docs-cloner --url https://example.com/sitemap.xml --selector ".docs-content"
```

### Polite crawling

```bash
docs-cloner --url https://example.com/sitemap.xml -c 2 -d 500
```

## Output format

Each page produces a `.md` file with YAML frontmatter:

```markdown
---
title: Page Title
source_url: https://example.com/docs/getting-started
crawl_date: 2026-02-13T15:30:00-05:00
---

Page content in clean markdown...
```

Files are organized to mirror the site structure:

```
output/
  docs/
    getting-started.md
    api/
      authentication.md
      endpoints.md
  blog/
    hello-world.md
```

## CLI Reference

```
Usage:
  docs-cloner [flags]

Flags:
      --url string                 Sitemap URL (required)
  -o, --output string              Output directory (default "./output")
      --fetch-md [pattern]         Fetch raw markdown instead of converting HTML.
                                   Without a value, appends .md to each URL.
                                   With a value, uses it as a URL pattern.
                                   Placeholders: {url}, {path}, {host}
  -c, --concurrency int            Parallel workers (default 5)
  -d, --delay int                  Per-worker delay between requests in ms (default 200)
      --single-file                Also produce a single concatenated all-pages.md
      --selector string            CSS selector for main content (default: auto-detect)
  -v, --verbose                    Log every page
      --user-agent string          Custom User-Agent (default "docs-cloner/1.0")
  -h, --help                       Show help
```

## How it works

1. Fetches and parses the XML sitemap (supports sitemap index files with sub-sitemaps)
2. Fans out page URLs to a configurable worker pool
3. Each worker fetches the page, extracts content using CSS selectors (heuristic cascade or explicit), and converts to markdown
4. Strips navigation, sidebars, footers, and other noise
5. Adds YAML frontmatter with title, source URL, and crawl date
6. Writes `.md` files mirroring the site's URL path structure
7. Optionally concatenates everything into a single file with a TOC

## Content extraction

When no `--selector` is provided, the tool tries these selectors in order and uses the first match with substantial content:

`main` > `article` > `[role="main"]` > `.content` > `.main-content` > `#content` > `.markdown-body` > `.documentation-content` > `.docs-content` > `.page-content`

Noise elements like `nav`, `.sidebar`, `.toc`, `.breadcrumb`, `script`, and `style` are removed before conversion.

## Limitations

- Does not execute JavaScript. Sites that render content client-side will produce empty or incomplete output. Use `--fetch-md` as a workaround for sites that serve raw markdown.
- Respects the sitemap only. Pages not listed in the sitemap won't be cloned.
- No robots.txt checking. Be respectful with concurrency and delay settings.
