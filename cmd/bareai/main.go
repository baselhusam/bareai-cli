package main

import (
	"os"

	"github.com/baselhusam/bareai-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
