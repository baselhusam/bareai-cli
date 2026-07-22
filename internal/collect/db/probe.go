package db

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

func probeDB(ctx context.Context, db *snapshot.Database) {
	if db == nil || db.Address == "" {
		return
	}
	start := time.Now()
	result := snapshot.ProbeResult{}

	conn, err := dialTCP(ctx, db.Address)
	if err != nil {
		result.Error = err.Error()
		result.LatencyMS = time.Since(start).Milliseconds()
		db.Health = &result
		return
	}
	defer conn.Close()
	result.OK = true
	result.LatencyMS = time.Since(start).Milliseconds()

	switch db.Engine {
	case EngineRedis:
		if v, msg := probeRedis(conn); v != "" {
			db.Version = v
			result.Message = msg
		} else if msg != "" {
			result.Message = msg
		}
	case EngineMySQL:
		if v := probeMySQL(conn); v != "" {
			db.Version = v
			result.Message = "mysql " + v
		}
	case EngineMongoDB:
		if v := probeMongoDB(conn); v != "" {
			db.Version = v
			result.Message = "mongodb " + v
		}
	case EnginePostgres:
		if v := probePostgres(conn); v != "" {
			db.Version = v
			result.Message = "postgres " + v
		}
	case EngineElasticsearch:
		if v := probeElasticsearch(ctx, db.Address); v != "" {
			db.Version = v
			result.Message = "elasticsearch " + v
		}
	case EngineQdrant:
		if v := probeQdrant(ctx, db.Address); v != "" {
			db.Version = v
			result.Message = "qdrant " + v
		}
	}

	if result.Message == "" {
		result.Message = "tcp ok"
	}
	db.Health = &result
}

func dialTCP(ctx context.Context, address string) (net.Conn, error) {
	d := net.Dialer{Timeout: 2 * time.Second}
	return d.DialContext(ctx, "tcp", address)
}

func probeRedis(conn net.Conn) (version, message string) {
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := fmt.Fprintf(conn, "PING\r\n"); err != nil {
		return "", ""
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(line, "+PONG") {
		return "", ""
	}
	if _, err := fmt.Fprintf(conn, "INFO server\r\n"); err != nil {
		return "", "pong"
	}
	for {
		l, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		l = strings.TrimSpace(l)
		if l == "" {
			break
		}
		if strings.HasPrefix(l, "redis_version:") {
			return strings.TrimPrefix(l, "redis_version:"), "pong"
		}
	}
	return "", "pong"
}

func probeMySQL(conn net.Conn) string {
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil || n < 5 {
		return ""
	}
	// MySQL handshake: protocol version (1 byte) + null-terminated server version string.
	if buf[0] != 10 {
		return ""
	}
	end := 1
	for end < n && buf[end] != 0 {
		end++
	}
	if end <= 1 {
		return ""
	}
	return string(buf[1:end])
}

func probeMongoDB(conn net.Conn) string {
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	// Minimal isMaster command on admin.$cmd
	payload := []byte(`{"isMaster":1}`)
	header := make([]byte, 16)
	header[0] = 16 + byte(len(payload))
	copy(header[4:12], []byte{0, 0, 0, 0, 0, 0, 0, 0})
	header[12] = 0
	header[13] = 0
	header[14] = 0
	header[15] = 0
	msg := append(header, payload...)
	if _, err := conn.Write(msg); err != nil {
		return ""
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return ""
	}
	body := string(buf)
	if idx := strings.Index(body, `"version"`); idx >= 0 {
		sub := body[idx:]
		if start := strings.Index(sub, `"`); start >= 0 {
			sub = sub[start+1:]
			if end := strings.Index(sub, `"`); end > 0 {
				return sub[:end]
			}
		}
	}
	return ""
}

func probePostgres(conn net.Conn) string {
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	// SSLRequest: length 8, code 80877103
	msg := []byte{0, 0, 0, 8, 4, 210, 22, 47}
	if _, err := conn.Write(msg); err != nil {
		return ""
	}
	buf := make([]byte, 1)
	if _, err := io.ReadFull(conn, buf); err != nil {
		return ""
	}
	// 'N' = no SSL, server will send error with version info
	if buf[0] != 'N' {
		return ""
	}
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	// Format: "N" + message + \0 + severity + \0 + ...
	parts := strings.Split(strings.TrimSuffix(line, "\n"), "\x00")
	for i, p := range parts {
		if p == "server_version" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func probeElasticsearch(ctx context.Context, address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return ""
	}
	url := fmt.Sprintf("http://%s:%s", host, port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var body struct {
		Version struct {
			Number string `json:"number"`
		} `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ""
	}
	return body.Version.Number
}

func probeQdrant(ctx context.Context, address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return ""
	}
	url := fmt.Sprintf("http://%s:%s/", host, port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var body struct {
		Title   string `json:"title"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return ""
	}
	if body.Version != "" {
		return body.Version
	}
	return body.Title
}
