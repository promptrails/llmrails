// Package together provides a Together AI LLM provider for unillm.
//
// Together AI exposes an OpenAI-compatible API.
//
// # Usage
//
//	provider := together.New("your-api-key")
//	resp, err := provider.Complete(ctx, &unillm.CompletionRequest{
//		Model:    "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
//		Messages: []unillm.Message{{Role: "user", Content: "Hello!"}},
//	})
package together

import (
	"context"
	"net/http"

	"github.com/promptrails/unillm"
	"github.com/promptrails/unillm/compat"
)

const defaultBaseURL = "https://api.together.xyz/v1/chat/completions"

// Provider implements unillm.Provider for Together AI's API.
type Provider struct{ inner *compat.Provider }

// Option configures the provider.
type Option func(*compat.Config)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option { return func(c *compat.Config) { c.BaseURL = url } }

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *compat.Config) { c.HTTPClient = client }
}

// New creates a new Together AI provider.
func New(apiKey string, opts ...Option) *Provider {
	cfg := compat.Config{Name: "together", BaseURL: defaultBaseURL, APIKey: apiKey}
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
