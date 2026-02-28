// Package tools provides automatic tool/function calling loop execution.
//
// When an LLM responds with tool calls, this package handles the cycle of:
// 1. Receiving tool calls from the LLM
// 2. Executing them via a ToolExecutor
// 3. Sending results back to the LLM
// 4. Repeating until the LLM gives a final text response
//
// # Usage
//
//	executor := tools.NewMap(map[string]tools.Func{
//		"get_weather": func(ctx context.Context, args string) (string, error) {
//			return `{"temp": 22}`, nil
//		},
//	})
//
//	resp, err := tools.RunLoop(ctx, provider, req, executor)
package tools

import (
	"context"
	"fmt"

	"github.com/promptrails/unillm"
)

// MaxIterations is the default maximum number of tool calling rounds.
// This prevents infinite loops when the model keeps calling tools.
const MaxIterations = 20

// Func is a function that executes a tool call.
// It receives the JSON-encoded arguments and returns a JSON-encoded result.
type Func func(ctx context.Context, arguments string) (string, error)

// Executor executes tool calls by name. Implementations should route
// to the correct tool function based on the tool name.
type Executor interface {
	// Execute runs a tool call and returns the result as a string.
	Execute(ctx context.Context, name string, arguments string) (string, error)
}

// MapExecutor is a simple Executor backed by a map of tool functions.
type MapExecutor struct {
	funcs map[string]Func
}

// NewMap creates an Executor from a map of tool name → function.
func NewMap(funcs map[string]Func) *MapExecutor {
	return &MapExecutor{funcs: funcs}
}

// Execute runs the named tool function. Returns an error if the tool is not found.
func (m *MapExecutor) Execute(ctx context.Context, name string, arguments string) (string, error) {
	fn, ok := m.funcs[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return fn(ctx, arguments)
}

// LoopOption configures the tool loop behavior.
type LoopOption func(*loopConfig)

type loopConfig struct {
	maxIterations int
	onToolCall    func(call unillm.ToolCall, result string, err error)
}

// WithMaxIterations sets the maximum number of tool calling rounds.
func WithMaxIterations(n int) LoopOption {
	return func(c *loopConfig) {
		c.maxIterations = n
	}
}

// WithToolCallHook sets a callback invoked after each tool call execution.
// Useful for logging, tracing, or metrics.
func WithToolCallHook(fn func(call unillm.ToolCall, result string, err error)) LoopOption {
	return func(c *loopConfig) {
		c.onToolCall = fn
	}
}

// RunLoopResult contains the final response and accumulated usage from all iterations.
type RunLoopResult struct {
	// Response is the final completion response (with text content, no more tool calls).
	Response *unillm.CompletionResponse

	// TotalUsage is the accumulated token usage across all iterations.
	TotalUsage unillm.TokenUsage

	// Iterations is the number of LLM calls made (including the initial one).
	Iterations int
}

// RunLoop executes the tool calling loop. It sends the request to the provider,
// and if the response contains tool calls, it executes them and sends the results
// back to the provider. This repeats until the model returns a text response
// or the maximum iterations are reached.
//
// The request's Messages slice is modified in place to append tool call
// and tool result messages.
func RunLoop(
	ctx context.Context,
	provider unillm.Provider,
	req *unillm.CompletionRequest,
	executor Executor,
	opts ...LoopOption,
) (*RunLoopResult, error) {
	cfg := loopConfig{maxIterations: MaxIterations}
	for _, opt := range opts {
		opt(&cfg)
	}

	result := &RunLoopResult{}

	for i := 0; i < cfg.maxIterations; i++ {
		resp, err := provider.Complete(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("tool loop iteration %d: %w", i+1, err)
		}

		result.Iterations++
		result.TotalUsage.PromptTokens += resp.Usage.PromptTokens
		result.TotalUsage.CompletionTokens += resp.Usage.CompletionTokens
		result.TotalUsage.TotalTokens += resp.Usage.TotalTokens

		// No tool calls — we have the final response
		if len(resp.ToolCalls) == 0 {
			result.Response = resp
			return result, nil
		}

		// Append assistant message with tool calls
		req.Messages = append(req.Messages, unillm.Message{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call and append results
		for _, tc := range resp.ToolCalls {
			toolResult, execErr := executor.Execute(ctx, tc.Name, tc.Arguments)

			if cfg.onToolCall != nil {
				cfg.onToolCall(tc, toolResult, execErr)
			}

			if execErr != nil {
				toolResult = fmt.Sprintf(`{"error": "%s"}`, execErr.Error())
			}

			req.Messages = append(req.Messages, unillm.Message{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: tc.ID,
			})
		}
	}

	return nil, fmt.Errorf("tool loop exceeded maximum iterations (%d)", cfg.maxIterations)
}
