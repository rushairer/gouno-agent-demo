package provider

import "context"

type Request struct {
	Instructions string
	Input        string
}

type Result struct {
	Text         string
	InputTokens  int
	OutputTokens int
}

type Provider interface {
	Name() string
	Generate(context.Context, Request) (Result, error)
	Stream(context.Context, Request, func(string) error) (Result, error)
}
