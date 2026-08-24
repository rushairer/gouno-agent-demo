package provider

import (
	"context"
	"encoding/json"
	"fmt"
)

func unsupported(name string) error {
	return fmt.Errorf("provider %s does not support this operation", name)
}

func (p *HTTPProvider) anthropicBody(req Request, stream bool) map[string]any {
	return map[string]any{"model": p.model, "max_tokens": 800, "system": req.Instructions, "messages": []map[string]string{{"role": "user", "content": req.Input}}, "stream": stream}
}

func (p *HTTPProvider) anthropicGenerate(ctx context.Context, req Request) (Result, error) {
	resp, err := p.do(ctx, "/v1/messages", p.anthropicBody(req, false), false)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	var decoded struct {
		Content []struct{ Type, Text string } `json:"content"`
		Usage   struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Result{}, err
	}
	text := ""
	for _, block := range decoded.Content {
		if block.Type == "text" {
			text += block.Text
		}
	}
	return Result{Text: text, InputTokens: decoded.Usage.InputTokens, OutputTokens: decoded.Usage.OutputTokens}, nil
}

func (p *HTTPProvider) anthropicStream(ctx context.Context, req Request, onDelta func(string) error) (Result, error) {
	resp, err := p.do(ctx, "/v1/messages", p.anthropicBody(req, true), true)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	err = readSSE(resp.Body, func(raw json.RawMessage) string {
		var event struct {
			Type  string                      `json:"type"`
			Delta struct{ Type, Text string } `json:"delta"`
		}
		if json.Unmarshal(raw, &event) == nil && event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
			return event.Delta.Text
		}
		return ""
	}, onDelta)
	return Result{}, err
}
