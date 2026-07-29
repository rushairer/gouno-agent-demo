package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gouno-agent-demo/config"
	"gouno-agent-demo/internal/agent"
	"gouno-agent-demo/internal/gateway"
)

type Handler struct {
	service  *agent.Service
	auth     *gateway.Authorizer
	timeout  time.Duration
	maxInput int
}

func New(service *agent.Service, auth *gateway.Authorizer, cfg config.GatewayConfig) *Handler {
	return &Handler{service: service, auth: auth, timeout: cfg.RequestTimeout, maxInput: cfg.MaxInputChars}
}

func (h *Handler) Ready(c *gin.Context) {
	if !h.service.Ready() {
		h.writeErrorStatus(c, http.StatusServiceUnavailable, "not_ready", "no provider is configured", c.GetString("request_id"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

type messageRequest struct {
	Provider string `json:"provider"`
	Message  string `json:"message"`
	Language string `json:"language,omitempty"`
}
type apiError struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"request_id"`
	} `json:"error"`
}

func (h *Handler) Message(c *gin.Context) {
	_, req, ok := h.authorize(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()
	answer, err := h.service.Answer(ctx, req.Provider, req.Message)
	if err != nil {
		h.writeError(c, err, c.GetString("request_id"))
		return
	}
	c.JSON(http.StatusOK, gin.H{"request_id": c.GetString("request_id"), "answer": answer.Text, "provider": answer.Provider, "citation": gin.H{"id": answer.Citation.ID, "title": answer.Citation.Title}, "usage": gin.H{"input_tokens": answer.InputTokens, "output_tokens": answer.OutputTokens}})
}

func (h *Handler) Stream(c *gin.Context) {
	_, req, ok := h.authorize(c)
	if !ok {
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	h.event(c, "meta", gin.H{"request_id": c.GetString("request_id"), "provider": req.Provider})
	answer, err := h.service.Stream(ctx, req.Provider, req.Message, func(delta string) error { h.event(c, "delta", gin.H{"text": delta}); return nil })
	if err != nil {
		h.event(c, "error", gin.H{"code": errorCode(err)})
		return
	}
	h.event(c, "completed", gin.H{"citation": gin.H{"id": answer.Citation.ID, "title": answer.Citation.Title}})
}

func (h *Handler) authorize(c *gin.Context) (gateway.Principal, messageRequest, bool) {
	principal, authenticated := h.auth.Authenticate(c.GetHeader("Authorization"))
	if !authenticated {
		h.writeErrorStatus(c, http.StatusUnauthorized, "unauthorized", "authentication is required", c.GetString("request_id"))
		return gateway.Principal{}, messageRequest{}, false
	}
	if !h.auth.Allow(principal) {
		h.writeErrorStatus(c, http.StatusTooManyRequests, "rate_limited", "request limit exceeded", c.GetString("request_id"))
		return gateway.Principal{}, messageRequest{}, false
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(h.maxInput*4))
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var req messageRequest
	if err := decoder.Decode(&req); err != nil || decoder.Decode(&struct{}{}) != io.EOF || len([]rune(req.Message)) > h.maxInput {
		h.writeErrorStatus(c, http.StatusBadRequest, "invalid_request", "request body is invalid", c.GetString("request_id"))
		return gateway.Principal{}, messageRequest{}, false
	}
	if !principal.AllowedProviders[req.Provider] {
		h.writeErrorStatus(c, http.StatusForbidden, "provider_forbidden", "provider is not allowed", c.GetString("request_id"))
		return gateway.Principal{}, messageRequest{}, false
	}
	return principal, req, true
}

func (h *Handler) event(c *gin.Context, event string, value any) {
	raw, _ := json.Marshal(value)
	c.SSEvent(event, string(raw))
	c.Writer.Flush()
}
func (h *Handler) writeError(c *gin.Context, err error, id string) {
	status := http.StatusBadGateway
	code := errorCode(err)
	if errors.Is(err, agent.ErrInputRejected) || errors.Is(err, agent.ErrOutOfScope) || errors.Is(err, agent.ErrOutputRejected) {
		status = http.StatusUnprocessableEntity
	}
	h.writeErrorStatus(c, status, code, "request could not be completed", id)
}
func errorCode(err error) string {
	switch {
	case errors.Is(err, agent.ErrInputRejected):
		return "input_rejected"
	case errors.Is(err, agent.ErrOutOfScope):
		return "out_of_scope"
	case errors.Is(err, agent.ErrOutputRejected):
		return "output_rejected"
	default:
		return "upstream_error"
	}
}
func (h *Handler) writeErrorStatus(c *gin.Context, status int, code, message, id string) {
	var response apiError
	response.Error.Code, response.Error.Message, response.Error.RequestID = code, message, id
	c.JSON(status, response)
}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) { c.Set("request_id", fmt.Sprintf("req_%d", time.Now().UnixNano())); c.Next() }
}
func MetadataAudit() gin.HandlerFunc { return func(c *gin.Context) { _ = time.Now(); c.Next() } }
