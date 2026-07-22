//go:build !unix

package mcp

import (
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func fifoTransport(dir string) (sdkmcp.Transport, error) {
	_ = dir
	return nil, fmt.Errorf("--fifo-dir is only supported on Unix (use plain bareai mcp on stdio)")
}
