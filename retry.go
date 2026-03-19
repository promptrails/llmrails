package llmrails

import (
	"context"
	"errors"
	"math"
	"time"
)

// RetryProvider wraps a Provider with automatic retry logic using
// exponential backoff. Only retryable errors (rate limits, server errors)
// are retried.
type RetryProvider struct {
	inner      Provider
	maxRetries int
	baseDelay  time.Duration
}

// RetryOption configures the retry behavior.
type RetryOption func(*RetryProvider)

// WithBaseDelay sets the base delay for exponential backoff.
// Default is 1 second. The actual delay doubles with each retry:
// 1s, 2s, 4s, 8s, etc.
func WithBaseDelay(d time.Duration) RetryOption {
	return func(r *RetryProvider) {
		r.baseDelay = d
	}
}

// WithRetry wraps a provider with retry logic. maxRetries is the maximum
// number of retry attempts (not including the initial attempt).
//
// Example:
//
//	provider := llmrails.WithRetry(openai.New("sk-..."), 3)
func WithRetry(provider Provider, maxRetries int, opts ...RetryOption) *RetryProvider {
	r := &RetryProvider{
		inner:      provider,
		maxRetries: maxRetries,
		baseDelay:  time.Second,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Complete sends a completion request with automatic retries.
func (r *RetryProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		resp, err := r.inner.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if !isRetryable(err) {
			return nil, err
		}

		if attempt < r.maxRetries {
			if err := r.sleep(ctx, attempt); err != nil {
				return nil, lastErr
			}
		}
	}

	return nil, lastErr
}

// Stream sends a streaming request with automatic retries.
// Note: only the initial connection is retried, not mid-stream failures.
func (r *RetryProvider) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamEvent, error) {
	var lastErr error

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		ch, err := r.inner.Stream(ctx, req)
		if err == nil {
			return ch, nil
		}

		lastErr = err

		if !isRetryable(err) {
			return nil, err
		}

		if attempt < r.maxRetries {
			if err := r.sleep(ctx, attempt); err != nil {
				return nil, lastErr
			}
		}
	}

	return nil, lastErr
}

func (r *RetryProvider) sleep(ctx context.Context, attempt int) error {
	delay := r.baseDelay * time.Duration(math.Pow(2, float64(attempt)))
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.IsRetryable()
	}
	return false
}
