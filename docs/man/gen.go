//go:build ignore

package main

import (
	"log"
	"os"
	"time"

	"github.com/spf13/cobra/doc"

	"github.com/baselhusam/bareai-cli/internal/cli"
)

func main() {
	header := &doc.GenManHeader{
		Title:   "bareai",
		Section: "1",
		Date:    &time.Time{},
	}
	if err := os.MkdirAll("docs/man/man1", 0o755); err != nil {
		log.Fatal(err)
	}
	if err := doc.GenManTree(cli.RootCommand(), header, "docs/man/man1"); err != nil {
		log.Fatal(err)
	}
}
