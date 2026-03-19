// Package mistral provides a Mistral AI LLM provider for llmrails.
//
// Mistral AI exposes an OpenAI-compatible API.
//
// # Usage
//
//	provider := mistral.New("your-api-key")
//	resp, err := provider.Complete(ctx, &llmrails.CompletionRequest{
//		Model:    "mistral-large-latest",
//		Messages: []llmrails.Message{{Role: "user", Content: "Hello!"}},
//	})
package mistral

import (
	"context"
	"net/http"

	"github.com/promptrails/llmrails"
	"github.com/promptrails/llmrails/compat"
)

const defaultBaseURL = "https://api.mistral.ai/v1/chat/completions"

// Provider implements llmrails.Provider for Mistral AI's API.
type Provider struct{ inner *compat.Provider }

// Option configures the provider.
type Option func(*compat.Config)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option { return func(c *compat.Config) { c.BaseURL = url } }

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *compat.Config) { c.HTTPClient = client }
}

// New creates a new Mistral AI provider.
func New(apiKey string, opts ...Option) *Provider {
	cfg := compat.Config{Name: "mistral", BaseURL: defaultBaseURL, APIKey: apiKey}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Provider{inner: compat.New(cfg)}
}

func (p *Provider) Complete(ctx context.Context, req *llmrails.CompletionRequest) (*llmrails.CompletionResponse, error) {
	return p.inner.Complete(ctx, req)
}

func (p *Provider) Stream(ctx context.Context, req *llmrails.CompletionRequest) (<-chan llmrails.StreamEvent, error) {
	return p.inner.Stream(ctx, req)
}
