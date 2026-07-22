package mcp

import (
	"context"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/baselhusam/bareai-cli/internal/version"
)

// RunOptions configures how the MCP server connects to its client.
type RunOptions struct {
	// FIFODir enables local two-terminal testing via named pipes at
	// FIFODir/in (client→server) and FIFODir/out (server→client).
	FIFODir string
}

// RunStdio starts the bareai MCP server on stdin/stdout or optional FIFO pair.
func RunStdio(ctx context.Context, opts RunOptions) error {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "bareai",
		Version: version.Version,
	}, nil)
	RegisterTools(server)
	server.AddReceivingMiddleware(requestLoggingMiddleware())
	logStartup(opts.FIFODir)

	var transport sdkmcp.Transport
	if opts.FIFODir != "" {
		t, err := fifoTransport(opts.FIFODir)
		if err != nil {
			return err
		}
		transport = t
	} else {
		transport = &sdkmcp.StdioTransport{}
	}

	if err := server.Run(ctx, transport); err != nil {
		return fmt.Errorf("mcp server: %w", err)
	}
	return nil
}
