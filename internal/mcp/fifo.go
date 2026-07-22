//go:build unix

package mcp

import (
	"fmt"
	"os"
	"path/filepath"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func fifoTransport(dir string) (sdkmcp.Transport, error) {
	inPath := filepath.Join(dir, "in")
	outPath := filepath.Join(dir, "out")
	for _, path := range []string{inPath, outPath} {
		st, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("fifo %s: %w (create with: mkfifo %q %q)", path, err, inPath, outPath)
		}
		if st.Mode()&os.ModeNamedPipe == 0 {
			return nil, fmt.Errorf("fifo %s: not a named pipe", path)
		}
	}

	// O_RDWR opens without blocking until a peer connects.
	in, err := os.OpenFile(inPath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open fifo in: %w", err)
	}
	out, err := os.OpenFile(outPath, os.O_RDWR, 0)
	if err != nil {
		in.Close()
		return nil, fmt.Errorf("open fifo out: %w", err)
	}

	logStderr("fifo transport on %s (client writes %s, reads %s)", dir, inPath, outPath)
	return &sdkmcp.IOTransport{Reader: in, Writer: out}, nil
}
