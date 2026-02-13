package main

import (
	"os"

	"github.com/Devon-White/docs-cloner/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
