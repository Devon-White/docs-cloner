package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/devon/docs-cloner/internal/config"
	"github.com/devon/docs-cloner/internal/pipeline"
	"github.com/spf13/cobra"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "docs-cloner",
	Short: "Clone documentation sites into AI-friendly markdown",
	Long: `docs-cloner fetches a documentation site via its XML sitemap and converts
each page to clean markdown suitable for use with AI systems.

It supports two modes:
  - HTML-to-Markdown (default): fetches each page's HTML, extracts the main
    content area, and converts it to clean markdown.
  - Raw Markdown (--fetch-md): fetches markdown directly from an alternate URL
    pattern, useful for sites that serve raw .md files.`,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVar(&cfg.SitemapURL, "url", "", "sitemap URL (required)")
	rootCmd.Flags().StringVarP(&cfg.OutputDir, "output", "o", "./output", "output directory")
	rootCmd.Flags().StringVar(&cfg.FetchMD, "fetch-md", "", "URL pattern for raw markdown (use {url}, {path}, {host} as placeholders; omit value to default to {url}.md)")
	rootCmd.Flags().Lookup("fetch-md").NoOptDefVal = "{url}.md"
	rootCmd.Flags().IntVarP(&cfg.Concurrency, "concurrency", "c", 5, "number of parallel workers")
	rootCmd.Flags().IntVarP(&cfg.DelayMS, "delay", "d", 200, "delay between requests per worker (ms)")
	rootCmd.Flags().BoolVar(&cfg.SingleFile, "single-file", false, "also produce a single concatenated all-pages.md")
	rootCmd.Flags().StringVar(&cfg.Selector, "selector", "", "CSS selector for main content area (default: auto-detect)")
	rootCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "verbose logging")
	rootCmd.Flags().StringVar(&cfg.UserAgent, "user-agent", "docs-cloner/1.0", "custom User-Agent string")

	rootCmd.MarkFlagRequired("url")
}

func run(cmd *cobra.Command, args []string) error {
	if cfg.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}
	if cfg.DelayMS < 0 {
		return fmt.Errorf("delay must be non-negative")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	return pipeline.Run(ctx, &cfg)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
