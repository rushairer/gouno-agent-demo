package provider

import (
	"context"
	"fmt"
	"gouno-agent-demo/config"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testServer(t *testing.T, handler http.Handler) (*httptest.Server, string) {
	t.Helper()
	server := httptest.NewUnstartedServer(handler)
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server.Listener = listener
	server.Start()
	t.Cleanup(server.Close)
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	return server, fmt.Sprintf("http://localhost:%s", port)
}

func TestOpenAIResponsesRequestAndResponse(t *testing.T) {
	t.Setenv("TEST_OPENAI_KEY", "secret")
	_, serverURL := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/responses" || r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("unexpected request %s %s", r.URL.Path, r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output_text":"ok","usage":{"input_tokens":2,"output_tokens":3}}`))
	}))
	p, err := NewHTTPProvider("openai", config.ProviderConfig{Enabled: true, BaseURL: serverURL, APIKeyEnv: "TEST_OPENAI_KEY", Model: "test"}, []string{"localhost"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	result, err := p.Generate(context.Background(), Request{Input: "VPN", Instructions: "safe"})
	if err != nil || result.Text != "ok" || result.OutputTokens != 3 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestAnthropicMessagesRequestAndResponse(t *testing.T) {
	t.Setenv("TEST_ANTHROPIC_KEY", "secret")
	_, serverURL := testServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" || r.Header.Get("x-api-key") != "secret" || r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Fatal("Anthropic headers/path invalid")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":2,"output_tokens":3}}`))
	}))
	p, err := NewHTTPProvider("anthropic", config.ProviderConfig{Enabled: true, BaseURL: serverURL, APIKeyEnv: "TEST_ANTHROPIC_KEY", Model: "test", APIVersion: "2023-06-01"}, []string{"localhost"}, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	result, err := p.Generate(context.Background(), Request{Input: "VPN", Instructions: "safe"})
	if err != nil || result.Text != "ok" || result.InputTokens != 2 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
