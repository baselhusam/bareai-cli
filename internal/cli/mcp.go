package cli

import (
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	bareaimcp "github.com/baselhusam/bareai-cli/internal/mcp"
)

var mcpFIFODir string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run MCP server for coding agents (stdio)",
	Long: `Start a Model Context Protocol server on stdin/stdout so coding agents
(Cursor, Claude Desktop, etc.) can inspect this box via bareai tools.

Activity logs (startup, initialize, tool calls) go to stderr; stdout is reserved for MCP protocol traffic.

Use --fifo-dir for local two-terminal testing with scripts/mcp-server-fifo.sh and
scripts/mcp-smoke.sh --attach. Cursor and Claude Desktop use plain stdio (no flag).
See docs/agents.md for setup examples.`,
	Example: `  bareai mcp
  bareai mcp --fifo-dir /tmp/bareai-mcp
  # Cursor MCP config: { "command": "bareai", "args": ["mcp"] }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()
		return bareaimcp.RunStdio(ctx, bareaimcp.RunOptions{FIFODir: mcpFIFODir})
	},
}

func init() {
	mcpCmd.Flags().StringVar(&mcpFIFODir, "fifo-dir", "", "Unix only: read/write MCP via named pipes at DIR/in and DIR/out (for local smoke testing)")
}
