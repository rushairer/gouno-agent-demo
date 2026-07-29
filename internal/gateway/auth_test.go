package gateway

import (
	"golang.org/x/crypto/bcrypt"
	"gouno-agent-demo/config"
	"testing"
)

func TestAuthenticateAuthorizeAndRateLimit(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("a-long-test-key-1234"), bcrypt.MinCost)
	auth := NewAuthorizer([]config.GatewayAPIKeyConfig{{ID: "demo", KeyHash: string(hash), RateLimitPerMinute: 1}})
	p, ok := auth.Authenticate("Bearer a-long-test-key-1234")
	if !ok || p.ID != "demo" {
		t.Fatal("expected authenticated OpenAI principal")
	}
	if !auth.Allow(p) || auth.Allow(p) {
		t.Fatal("expected per-principal rate limit")
	}
	if _, ok := auth.Authenticate("Bearer wrong"); ok {
		t.Fatal("wrong key authenticated")
	}
}
