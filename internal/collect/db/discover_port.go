package db

import (
	"context"
	"net"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func discoverPorts(ctx context.Context) ([]candidate, []snapshot.Skip) {
	var out []candidate
	seen := make(map[string]bool)

	dialer := net.Dialer{Timeout: 500 * time.Millisecond}
	for port, engine := range defaultPortEngines {
		addr := baseAddr(port)
		if seen[addr] {
			continue
		}
		conn, err := dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			continue
		}
		conn.Close()
		seen[addr] = true
		out = append(out, candidate{
			priority: 1,
			Database: snapshot.Database{
				Engine:  engine,
				Name:    displayName(engine),
				Address: addr,
				Source:  sourcePort,
			},
		})
	}
	return out, nil
}
