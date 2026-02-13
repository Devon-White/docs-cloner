package main

import (
	"os"

	"github.com/devon/docs-cloner/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
