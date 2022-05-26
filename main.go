package main

import (
	"os"

	"github.com/cosmosquad-labs/blockparser/cmd"
)

func main() {
	if err := cmd.NewBlockParserCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
