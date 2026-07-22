package mcp

import (
	"context"
	"fmt"
	"os"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/baselhusam/bareai-cli/internal/version"
)

const registeredToolCount = 6

func logStderr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "bareai mcp: "+format+"\n", args...)
}

func logStartup(fifoDir string) {
	logStderr("server ready (version %s, %d tools)", version.Version, registeredToolCount)
	if fifoDir != "" {
		logStderr("waiting for client on fifo %s (e.g. ./scripts/mcp-smoke.sh --attach %s)", fifoDir, fifoDir)
		return
	}
	logStderr("stdio transport · stdout=MCP protocol · stderr=logs · waiting for client on stdin…")
	logStderr("note: another terminal running mcp-smoke without --attach starts its own server")
}

func requestLoggingMiddleware() sdkmcp.Middleware {
	return func(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
		return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
			extra := requestLogDetail(method, req)
			logStderr("← %s%s", method, extra)

			start := time.Now()
			res, err := next(ctx, method, req)
			dur := formatDuration(time.Since(start))

			if err != nil {
				logStderr("→ %s failed (%s): %v", method, dur, err)
			} else {
				logStderr("→ %s ok (%s)", method, dur)
			}
			return res, err
		}
	}
}

func requestLogDetail(method string, req sdkmcp.Request) string {
	switch method {
	case "initialize":
		if ir, ok := req.(*sdkmcp.ServerRequest[*sdkmcp.InitializeParams]); ok && ir.Params != nil && ir.Params.ClientInfo != nil {
			return fmt.Sprintf(" client=%q", ir.Params.ClientInfo.Name)
		}
	case "tools/call":
		if ctr, ok := req.(*sdkmcp.CallToolRequest); ok && ctr.Params != nil {
			return fmt.Sprintf(" tool=%q", ctr.Params.Name)
		}
	}
	return ""
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "<1ms"
	}
	return d.Round(time.Millisecond).String()
}
