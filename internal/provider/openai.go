package provider

import (
	"context"
	"encoding/json"
	"io"
	"strings"
)

func (p *HTTPProvider) Generate(ctx context.Context, req Request) (Result, error) {
	if p.name == "anthropic" {
		return p.anthropicGenerate(ctx, req)
	}
	if p.name != "openai" {
		return Result{}, unsupported(p.name)
	}
	if !p.responsesStreamRequired {
		return p.openAINonStreamingGenerate(ctx, req)
	}
	// Some OpenAI-compatible relays only support streaming Responses requests.
	// Collect the SSE events here so the public synchronous API remains unchanged.
	resp, err := p.do(ctx, "/v1/responses", openAIResponsesBody(p.model, req, true), true)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	return readOpenAIStream(resp.Body, func(string) error { return nil })
}

func (p *HTTPProvider) openAINonStreamingGenerate(ctx context.Context, req Request) (Result, error) {
	resp, err := p.do(ctx, "/v1/responses", openAIResponsesBody(p.model, req, false), false)
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
	resp, err := p.do(ctx, "/v1/responses", openAIResponsesBody(p.model, req, true), true)
	if err != nil {
		return Result{}, err
	}
	defer resp.Body.Close()
	return readOpenAIStream(resp.Body, onDelta)
}

func readOpenAIStream(bodyReader io.Reader, onDelta func(string) error) (Result, error) {
	var result Result
	var output strings.Builder
	err := readSSE(bodyReader, func(raw json.RawMessage) string {
		var event struct {
			Type     string `json:"type"`
			Delta    string `json:"delta"`
			Response struct {
				Usage struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			} `json:"response"`
		}
		if json.Unmarshal(raw, &event) != nil {
			return ""
		}
		if event.Type == "response.completed" {
			result.InputTokens = event.Response.Usage.InputTokens
			result.OutputTokens = event.Response.Usage.OutputTokens
		}
		if event.Type == "response.output_text.delta" {
			return event.Delta
		}
		return ""
	}, func(delta string) error {
		output.WriteString(delta)
		return onDelta(delta)
	})
	result.Text = output.String()
	return result, err
}

// openAIResponsesBody always uses the structured input form. Although the
// Responses API also accepts a string, some OpenAI-compatible relays only
// accept the list form.
func openAIResponsesBody(model string, req Request, stream bool) map[string]any {
	body := map[string]any{
		"model":        model,
		"instructions": req.Instructions,
		"input": []any{
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{"type": "input_text", "text": req.Input},
				},
			},
		},
	}
	if stream {
		body["stream"] = true
	}
	return body
}
