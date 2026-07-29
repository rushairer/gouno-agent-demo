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
	"gouno-agent-demo/internal/gateway"
	"gouno-agent-demo/internal/provider"
)

type testProvider struct{}

func (testProvider) Name() string { return "openai" }
func (testProvider) Generate(context.Context, provider.Request) (provider.Result, error) {
	return provider.Result{Text: "请重新连接 VPN。"}, nil
}
func (testProvider) Stream(_ context.Context, _ provider.Request, callback func(string) error) (provider.Result, error) {
	return provider.Result{}, callback("请重新连接 VPN。")
}

func TestMessageEndpointAuthenticationAndSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash, _ := bcrypt.GenerateFromPassword([]byte("a-long-test-key-1234"), bcrypt.MinCost)
	handler := New(agent.New(map[string]provider.Provider{"openai": testProvider{}}), gateway.NewAuthorizer([]config.GatewayAPIKeyConfig{{ID: "test", KeyHash: string(hash), AllowedProviders: []string{"openai"}, RateLimitPerMinute: 10}}), config.GatewayConfig{RequestTimeout: time.Second, MaxInputChars: 100})
	engine := gin.New()
	engine.Use(RequestID())
	engine.POST("/v1/agent/messages", handler.Message)
	request := httptest.NewRequest(http.MethodPost, "/v1/agent/messages", strings.NewReader(`{"provider":"openai","message":"VPN 连不上"}`))
	request.Header.Set("Authorization", "Bearer a-long-test-key-1234")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"citation":{"id":"vpn"`) {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodPost, "/v1/agent/messages", strings.NewReader(`{"provider":"openai","message":"VPN 连不上"}`))
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("code=%d", response.Code)
	}
}
