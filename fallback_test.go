package llmrails

import (
	"context"
	"errors"
	"testing"
)

func TestFallbackProvider_UsesPrimary(t *testing.T) {
	primary := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return &CompletionResponse{Content: "primary"}, nil
		},
	}
	fallback := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return &CompletionResponse{Content: "fallback"}, nil
		},
	}

	provider := WithFallback(primary, fallback)
	resp, err := provider.Complete(context.Background(), &CompletionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "primary" {
		t.Errorf("expected 'primary', got %q", resp.Content)
	}
}

func TestFallbackProvider_FallsBackOnError(t *testing.T) {
	primary := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return nil, errors.New("primary failed")
		},
	}
	fallback := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return &CompletionResponse{Content: "fallback"}, nil
		},
	}

	provider := WithFallback(primary, fallback)
	resp, err := provider.Complete(context.Background(), &CompletionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "fallback" {
		t.Errorf("expected 'fallback', got %q", resp.Content)
	}
}

func TestFallbackProvider_BothFail(t *testing.T) {
	primary := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return nil, errors.New("primary failed")
		},
	}
	fallback := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return nil, errors.New("fallback failed")
		},
	}

	provider := WithFallback(primary, fallback)
	_, err := provider.Complete(context.Background(), &CompletionRequest{})
	if err == nil {
		t.Fatal("expected error when both providers fail")
	}
	if err.Error() != "fallback failed" {
		t.Errorf("expected fallback error, got: %v", err)
	}
}

func TestFallbackProvider_Chained(t *testing.T) {
	p1 := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return nil, errors.New("p1 failed")
		},
	}
	p2 := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return nil, errors.New("p2 failed")
		},
	}
	p3 := &mockProvider{
		completeFunc: func(_ context.Context, _ *CompletionRequest) (*CompletionResponse, error) {
			return &CompletionResponse{Content: "p3"}, nil
		},
	}

	provider := WithFallback(p1, WithFallback(p2, p3))
	resp, err := provider.Complete(context.Background(), &CompletionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "p3" {
		t.Errorf("expected 'p3', got %q", resp.Content)
	}
}
