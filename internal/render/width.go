package render

import (
	"io"
	"os"
	"strconv"

	"golang.org/x/term"
)

const defaultTerminalWidth = 80

// TerminalWidth returns the best-effort terminal width for layout.
func TerminalWidth(w io.Writer) int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil && n > 0 {
			return n
		}
	}

	if f, ok := w.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		if width, _, err := term.GetSize(int(f.Fd())); err == nil && width > 0 {
			return width
		}
	}
	return defaultTerminalWidth
}
