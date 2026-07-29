package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gouno-agent-demo/config"
	"gouno-agent-demo/internal/agent"
	"gouno-agent-demo/internal/billing"
	"gouno-agent-demo/internal/gateway"
	"gouno-agent-demo/internal/provider"
)

type testProvider struct{}

func (testProvider) Name() string { return "openai" }
func (testProvider) Generate(context.Context, provider.Request) (provider.Result, error) {
	return provider.Result{Text: "请重新连接 VPN。", InputTokens: 2, OutputTokens: 3}, nil
}

type recordingUsageRecorder struct{ events []billing.UsageEvent }

func (r *recordingUsageRecorder) Record(_ context.Context, event billing.UsageEvent) error {
	r.events = append(r.events, event)
	return nil
}
func (testProvider) Stream(_ context.Context, _ provider.Request, callback func(string) error) (provider.Result, error) {
	return provider.Result{}, callback("请重新连接 VPN。")
}

func TestMessageEndpointAuthenticationAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash, _ := bcrypt.GenerateFromPassword([]byte("a-long-test-key-1234"), bcrypt.MinCost)
	recorder := &recordingUsageRecorder{}
	handler := New(agent.New(map[string]provider.Provider{"openai": testProvider{}}, "openai"), gateway.NewAuthorizer([]config.GatewayAPIKeyConfig{{ID: "test", KeyHash: string(hash), RateLimitPerMinute: 10}}), config.GatewayConfig{RequestTimeout: time.Second, MaxInputChars: 100, DefaultProvider: "openai", Providers: map[string]config.ProviderConfig{"openai": {Model: "gpt-5-mini"}}}, recorder, nil)
	engine := gin.New()
	engine.Use(RequestID())
	engine.POST("/v1/agent/messages", handler.Message)
	request := httptest.NewRequest(http.MethodPost, "/v1/agent/messages", strings.NewReader(`{"message":"VPN 连不上","language":"en"}`))
	request.Header.Set("Authorization", "Bearer a-long-test-key-1234")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"citation":{"id":"vpn"`) {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
	if len(recorder.events) != 1 || recorder.events[0].AccountID != "test" || recorder.events[0].Provider != "openai" || recorder.events[0].Model != "gpt-5-mini" || recorder.events[0].InputTokens != 2 || recorder.events[0].OutputTokens != 3 {
		t.Fatalf("events=%+v", recorder.events)
	}
	request = httptest.NewRequest(http.MethodPost, "/v1/agent/messages", strings.NewReader(`{"message":"VPN 连不上"}`))
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", response.Code)
	}
	request = httptest.NewRequest(http.MethodPost, "/v1/agent/messages", strings.NewReader(`{"message":"VPN 连不上","language":"ja"}`))
	request.Header.Set("Authorization", "Bearer a-long-test-key-1234")
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"invalid_language"`) {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodPost, "/v1/agent/messages", strings.NewReader(`{"provider":"anthropic","message":"VPN 连不上"}`))
	request.Header.Set("Authorization", "Bearer a-long-test-key-1234")
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"invalid_request"`) {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
	if len(recorder.events) != 1 {
		t.Fatalf("usage must only be recorded for completed calls: %+v", recorder.events)
	}
}
