// Package llmrails provides a unified interface for interacting with multiple
// LLM (Large Language Model) providers through a single, consistent API.
//
// It supports 11+ providers including OpenAI, Anthropic, Google Gemini,
// DeepSeek, Groq, Fireworks, xAI, OpenRouter, Together, Mistral, and Cohere.
//
// # Core Interface
//
// All providers implement the [Provider] interface with two methods:
//
//   - Complete: sends a request and returns the full response
//   - Stream: sends a request and returns a channel of streaming events
//
// # Quick Start
//
//	import (
//		"github.com/promptrails/llmrails/openai"
//	)
//
//	provider := openai.New("sk-...")
//	resp, err := provider.Complete(ctx, &llmrails.CompletionRequest{
//		Model:    "gpt-4o",
//		Messages: []llmrails.Message{{Role: "user", Content: "Hello!"}},
//	})
//
// # Streaming
//
//	events, err := provider.Stream(ctx, &llmrails.CompletionRequest{
//		Model:    "gpt-4o",
//		Messages: []llmrails.Message{{Role: "user", Content: "Hello!"}},
//	})
//	for event := range events {
//		if event.Type == llmrails.EventContent {
//			fmt.Print(event.Content)
//		}
//	}
//
// # Provider Decorators
//
// Providers can be wrapped with decorators for retry and fallback behavior:
//
//	provider := llmrails.WithRetry(openai.New("sk-..."), 3)
//	provider = llmrails.WithFallback(provider, anthropic.New("sk-..."))
package llmrails
