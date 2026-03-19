// Package openrouter provides an OpenRouter LLM provider for llmrails.
//
// OpenRouter is a unified API gateway that routes to multiple LLM providers.
// It exposes an OpenAI-compatible API with additional headers for site
// identification and ranking.
//
// # Usage
//
//	provider := openrouter.New("your-api-key")
//	resp, err := provider.Complete(ctx, &llmrails.CompletionRequest{
//		Model:    "openai/gpt-4o",
//		Messages: []llmrails.Message{{Role: "user", Content: "Hello!"}},
//	})
package openrouter

import (
	"context"
	"net/http"

	"github.com/promptrails/llmrails"
	"github.com/promptrails/llmrails/compat"
)

const defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"

// Provider implements llmrails.Provider for OpenRouter's API.
type Provider struct{ inner *compat.Provider }

// Option configures the provider.
type Option func(*compat.Config)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option { return func(c *compat.Config) { c.BaseURL = url } }

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *compat.Config) { c.HTTPClient = client }
}

// WithSiteInfo sets the HTTP-Referer and X-Title headers for OpenRouter ranking.
func WithSiteInfo(referer, title string) Option {
	return func(c *compat.Config) {
		if c.ExtraHeaders == nil {
			c.ExtraHeaders = make(map[string]string)
		}
		c.ExtraHeaders["HTTP-Referer"] = referer
		c.ExtraHeaders["X-Title"] = title
	}
}

// New creates a new OpenRouter provider.
func New(apiKey string, opts ...Option) *Provider {
	cfg := compat.Config{Name: "openrouter", BaseURL: defaultBaseURL, APIKey: apiKey}
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
