package gateway

import (
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gouno-agent-demo/config"
)

type Principal struct {
	ID               string
	AllowedProviders map[string]bool
	Limit            int
}
type Authorizer struct {
	keys   []config.GatewayAPIKeyConfig
	mu     sync.Mutex
	visits map[string][]time.Time
}

func NewAuthorizer(keys []config.GatewayAPIKeyConfig) *Authorizer {
	return &Authorizer{keys: keys, visits: make(map[string][]time.Time)}
}

func (a *Authorizer) Authenticate(header string) (Principal, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return Principal{}, false
	}
	for _, item := range a.keys {
		if item.KeyHash == "" {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(item.KeyHash), []byte(parts[1])) == nil {
			providers := make(map[string]bool, len(item.AllowedProviders))
			for _, provider := range item.AllowedProviders {
				providers[provider] = true
			}
			return Principal{ID: item.ID, AllowedProviders: providers, Limit: item.RateLimitPerMinute}, true
		}
	}
	return Principal{}, false
}

func (a *Authorizer) Allow(principal Principal) bool {
	if principal.Limit <= 0 {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now, cutoff := time.Now(), time.Now().Add(-time.Minute)
	requests := a.visits[principal.ID][:0]
	for _, at := range a.visits[principal.ID] {
		if at.After(cutoff) {
			requests = append(requests, at)
		}
	}
	if len(requests) >= principal.Limit {
		a.visits[principal.ID] = requests
		return false
	}
	a.visits[principal.ID] = append(requests, now)
	return true
}
