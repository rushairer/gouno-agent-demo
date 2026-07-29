package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gouno-agent-demo/config"
)

type HTTPProvider struct {
	name, baseURL, key, model, version string
	responsesStreamRequired            bool
	client                             *http.Client
}

func NewHTTPProvider(name string, cfg config.ProviderConfig, allowedHosts []string, timeout time.Duration) (*HTTPProvider, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	u, err := url.Parse(cfg.BaseURL)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Hostname() != "localhost") {
		return nil, fmt.Errorf("%s: invalid upstream base_url", name)
	}
	allowed := false
	for _, host := range allowedHosts {
		if strings.EqualFold(host, u.Hostname()) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, fmt.Errorf("%s: upstream host is not allowed", name)
	}
	key := os.Getenv(cfg.APIKeyEnv)
	if key == "" {
		return nil, fmt.Errorf("%s: environment variable %s is empty", name, cfg.APIKeyEnv)
	}
	if cfg.Model == "" {
		return nil, fmt.Errorf("%s: model is required", name)
	}
	return &HTTPProvider{name: name, baseURL: strings.TrimRight(cfg.BaseURL, "/"), key: key, model: cfg.Model, version: cfg.APIVersion, responsesStreamRequired: cfg.ResponsesStreamRequired, client: &http.Client{Timeout: timeout, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}, nil
}

func (p *HTTPProvider) Name() string { return p.name }

func (p *HTTPProvider) do(ctx context.Context, path string, body any, stream bool) (*http.Response, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	if p.name == "openai" {
		req.Header.Set("Authorization", "Bearer "+p.key)
	} else {
		req.Header.Set("x-api-key", p.key)
		req.Header.Set("anthropic-version", p.version)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("upstream %s returned %d: %s", p.name, resp.StatusCode, strings.TrimSpace(string(limited)))
	}
	return resp, nil
}

func readSSE(body io.Reader, extract func(json.RawMessage) string, onDelta func(string) error) error {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		delta := extract(json.RawMessage(data))
		if delta != "" {
			if err := onDelta(delta); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
