// Package billing defines the boundary between completed model calls and future
// usage-ledger, quota, and payment implementations.
package billing

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// UsageEvent contains billing-safe metadata for one successfully completed model call.
// It deliberately excludes the request and response bodies.
type UsageEvent struct {
	RequestID    string
	AccountID    string
	Provider     string
	Model        string
	InputTokens  int
	OutputTokens int
	CompletedAt  time.Time
}

// UsageRecorder records completed usage. Replace LoggingUsageRecorder with a durable,
// idempotent ledger before charging customers or enforcing paid quotas. RequestID is
// only a server-generated trace ID; a future client Idempotency-Key must be persisted
// and used as the account-scoped ledger idempotency key.
type UsageRecorder interface {
	Record(context.Context, UsageEvent) error
}

type LoggingUsageRecorder struct{ logger *zap.Logger }

func NewLoggingUsageRecorder(logger *zap.Logger) *LoggingUsageRecorder {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &LoggingUsageRecorder{logger: logger}
}

func (r *LoggingUsageRecorder) Record(_ context.Context, event UsageEvent) error {
	r.logger.Info("usage recorded; replace logging recorder with a durable billing ledger before charging customers",
		zap.String("request_id", event.RequestID),
		zap.String("account_id", event.AccountID),
		zap.String("provider", event.Provider),
		zap.String("model", event.Model),
		zap.Int("input_tokens", event.InputTokens),
		zap.Int("output_tokens", event.OutputTokens),
		zap.Time("completed_at", event.CompletedAt),
	)
	return nil
}
