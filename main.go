package main

import (
	"os"

	"github.com/mvillene/zoink/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
