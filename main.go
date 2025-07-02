package main

import (
	"os"

	"github.com/iammatthew2/zoink/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
