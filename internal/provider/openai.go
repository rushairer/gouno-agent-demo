package provider

import (
	"context"
	"encoding/json"
)

func (p *HTTPProvider) Generate(ctx context.Context, req Request) (Result, error) {
	if p.name == "anthropic" {
		return p.anthropicGenerate(ctx, req)
	}
	if p.name != "openai" {
		return Result{}, unsupported(p.name)
	}
	resp, err := p.do(ctx, "/v1/responses", map[string]any{"model": p.model, "instructions": req.Instructions, "input": req.Input}, false)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	var decoded struct {
		OutputText string `json:"output_text"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Result{}, err
	}
	return Result{Text: decoded.OutputText, InputTokens: decoded.Usage.InputTokens, OutputTokens: decoded.Usage.OutputTokens}, nil
}

func (p *HTTPProvider) Stream(ctx context.Context, req Request, onDelta func(string) error) (Result, error) {
	if p.name == "anthropic" {
		return p.anthropicStream(ctx, req, onDelta)
	}
	if p.name != "openai" {
		return Result{}, unsupported(p.name)
	}
	resp, err := p.do(ctx, "/v1/responses", map[string]any{"model": p.model, "instructions": req.Instructions, "input": req.Input, "stream": true}, true)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	err = readSSE(resp.Body, func(raw json.RawMessage) string {
		var event struct {
			Type  string `json:"type"`
			Delta string `json:"delta"`
		}
		if json.Unmarshal(raw, &event) == nil && event.Type == "response.output_text.delta" {
			return event.Delta
		}
		return ""
	}, onDelta)
	return Result{}, err
}
