package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/baselhusam/bareai-cli/internal/snapshot"
)

// NewClient returns an HTTP client bounded by ctx.
func NewClient(ctx context.Context) *http.Client {
	return &http.Client{
		Timeout: timeoutFromContext(ctx),
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func timeoutFromContext(ctx context.Context) time.Duration {
	if deadline, ok := ctx.Deadline(); ok {
		if d := time.Until(deadline); d > 0 {
			return d
		}
	}
	return 10 * time.Second
}

func get(ctx context.Context, client *http.Client, url string) (*http.Response, []byte, time.Duration, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, 0, err
	}
	resp, err := client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return nil, nil, latency, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp, nil, latency, err
	}
	return resp, body, latency, nil
}

func postJSON(ctx context.Context, client *http.Client, url, payload string) (*http.Response, []byte, time.Duration, error) {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return nil, nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return nil, nil, latency, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return resp, nil, latency, err
	}
	return resp, body, latency, nil
}

func failResult(latency time.Duration, status int, err error) snapshot.ProbeResult {
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	return snapshot.ProbeResult{
		OK:        false,
		LatencyMS: latency.Milliseconds(),
		Status:    status,
		Error:     msg,
	}
}

func okResult(latency time.Duration, status int, message string) snapshot.ProbeResult {
	return snapshot.ProbeResult{
		OK:        true,
		LatencyMS: latency.Milliseconds(),
		Status:    status,
		Message:   message,
	}
}

func joinURL(base, path string) string {
	return strings.TrimRight(base, "/") + path
}

func defaultModel(runtime string, models []snapshot.LLMModel) string {
	if len(models) > 0 && models[0].ID != "" {
		return models[0].ID
	}
	switch runtime {
	case RuntimeOllama:
		return "llama3.2"
	case RuntimeVLLM, RuntimeSGLang, RuntimeOpenAICompat:
		return "default"
	default:
		return ""
	}
}

func escapeJSON(s string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	return replacer.Replace(s)
}

func readError(status int, body []byte, latency time.Duration) snapshot.ProbeResult {
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}
	return snapshot.ProbeResult{
		OK:        false,
		LatencyMS: latency.Milliseconds(),
		Status:    status,
		Error:     msg,
	}
}
