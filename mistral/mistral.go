// Package mistral provides a Mistral AI LLM provider for unillm.
//
// Mistral AI exposes an OpenAI-compatible API.
//
// # Usage
//
//	provider := mistral.New("your-api-key")
//	resp, err := provider.Complete(ctx, &unillm.CompletionRequest{
//		Model:    "mistral-large-latest",
//		Messages: []unillm.Message{{Role: "user", Content: "Hello!"}},
//	})
package mistral

import (
	"context"
	"net/http"

	"github.com/promptrails/unillm"
	"github.com/promptrails/unillm/compat"
)

const defaultBaseURL = "https://api.mistral.ai/v1/chat/completions"

// Provider implements unillm.Provider for Mistral AI's API.
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

func (p *Provider) Complete(ctx context.Context, req *unillm.CompletionRequest) (*unillm.CompletionResponse, error) {
	return p.inner.Complete(ctx, req)
}

func (p *Provider) Stream(ctx context.Context, req *unillm.CompletionRequest) (<-chan unillm.StreamEvent, error) {
	return p.inner.Stream(ctx, req)
}
