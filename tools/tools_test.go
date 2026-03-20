package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/promptrails/langrails"
)

type mockProvider struct {
	calls     int
	responses []*langrails.CompletionResponse
}

func (m *mockProvider) Complete(_ context.Context, _ *langrails.CompletionRequest) (*langrails.CompletionResponse, error) {
	idx := m.calls
	m.calls++
	if idx < len(m.responses) {
		return m.responses[idx], nil
	}
	return &langrails.CompletionResponse{Content: "done"}, nil
}

func (m *mockProvider) Stream(_ context.Context, _ *langrails.CompletionRequest) (<-chan langrails.StreamEvent, error) {
	return nil, nil
}

func TestRunLoop_NoTools(t *testing.T) {
	provider := &mockProvider{
		responses: []*langrails.CompletionResponse{
			{Content: "Hello!", Usage: langrails.TokenUsage{TotalTokens: 10}},
		},
	}

	result, err := RunLoop(context.Background(), provider, &langrails.CompletionRequest{
		Model:    "test",
		Messages: []langrails.Message{{Role: "user", Content: "Hi"}},
	}, NewMap(nil))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response.Content != "Hello!" {
		t.Errorf("expected 'Hello!', got %q", result.Response.Content)
	}
	if result.Iterations != 1 {
		t.Errorf("expected 1 iteration, got %d", result.Iterations)
	}
}

func TestRunLoop_SingleToolCall(t *testing.T) {
	provider := &mockProvider{
		responses: []*langrails.CompletionResponse{
			{
				ToolCalls: []langrails.ToolCall{
					{ID: "call_1", Name: "get_weather", Arguments: `{"city":"Istanbul"}`},
				},
				Usage: langrails.TokenUsage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
			},
			{
				Content: "It's 22°C in Istanbul.",
				Usage:   langrails.TokenUsage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
			},
		},
	}

	executor := NewMap(map[string]Func{
		"get_weather": func(_ context.Context, args string) (string, error) {
			var parsed map[string]string
			_ = json.Unmarshal([]byte(args), &parsed)
			if parsed["city"] != "Istanbul" {
				t.Errorf("expected city Istanbul, got %q", parsed["city"])
			}
			return `{"temp": 22, "condition": "sunny"}`, nil
		},
	})

	result, err := RunLoop(context.Background(), provider, &langrails.CompletionRequest{
		Model:    "test",
		Messages: []langrails.Message{{Role: "user", Content: "Weather in Istanbul?"}},
	}, executor)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Response.Content != "It's 22°C in Istanbul." {
		t.Errorf("unexpected content: %q", result.Response.Content)
	}
	if result.Iterations != 2 {
		t.Errorf("expected 2 iterations, got %d", result.Iterations)
	}
	if result.TotalUsage.TotalTokens != 45 {
		t.Errorf("expected 45 total tokens, got %d", result.TotalUsage.TotalTokens)
	}
}

func TestRunLoop_MaxIterations(t *testing.T) {
	// Provider always returns tool calls — should hit max iterations
	provider := &mockProvider{
		responses: make([]*langrails.CompletionResponse, 10),
	}
	for i := range provider.responses {
		provider.responses[i] = &langrails.CompletionResponse{
			ToolCalls: []langrails.ToolCall{
				{ID: "call", Name: "loop_tool", Arguments: "{}"},
			},
		}
	}

	executor := NewMap(map[string]Func{
		"loop_tool": func(_ context.Context, _ string) (string, error) {
			return "ok", nil
		},
	})

	_, err := RunLoop(context.Background(), provider, &langrails.CompletionRequest{
		Model:    "test",
		Messages: []langrails.Message{{Role: "user", Content: "loop"}},
	}, executor, WithMaxIterations(3))

	if err == nil {
		t.Fatal("expected max iterations error")
	}
}

func TestRunLoop_ToolCallHook(t *testing.T) {
	provider := &mockProvider{
		responses: []*langrails.CompletionResponse{
			{ToolCalls: []langrails.ToolCall{{ID: "c1", Name: "test_tool", Arguments: `{}`}}},
			{Content: "done"},
		},
	}

	hookCalled := false
	executor := NewMap(map[string]Func{
		"test_tool": func(_ context.Context, _ string) (string, error) {
			return "result", nil
		},
	})

	_, err := RunLoop(context.Background(), provider, &langrails.CompletionRequest{
		Model:    "test",
		Messages: []langrails.Message{{Role: "user", Content: "test"}},
	}, executor, WithToolCallHook(func(call langrails.ToolCall, result string, err error) {
		hookCalled = true
		if call.Name != "test_tool" {
			t.Errorf("expected tool 'test_tool', got %q", call.Name)
		}
		if result != "result" {
			t.Errorf("expected result 'result', got %q", result)
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hookCalled {
		t.Error("hook was not called")
	}
}

func TestMapExecutor_UnknownTool(t *testing.T) {
	executor := NewMap(map[string]Func{})
	_, err := executor.Execute(context.Background(), "nonexistent", "{}")
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}
