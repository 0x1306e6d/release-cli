package main

import (
	"os"

	"github.com/0x1306e6d/release-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
