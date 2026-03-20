# Request Parameters

Complete reference for all `CompletionRequest` fields and their provider support.

## Parameter Reference

### Model & Messages

```go
req := &langrails.CompletionRequest{
    Model:        "gpt-4o",                    // Required: model identifier
    SystemPrompt: "You are a helpful assistant", // Optional: system instruction
    Messages: []langrails.Message{                 // Required: conversation history
        {Role: "user", Content: "Hello!"},
    },
}
```

### Temperature

Controls randomness. Lower = more deterministic, higher = more creative.

```go
temp := 0.7
req.Temperature = &temp  // Range: 0-2 (OpenAI), 0-1 (Anthropic)
```

| Provider | Range | Default |
|----------|-------|---------|
| OpenAI | 0 - 2 | 1 |
| Anthropic | 0 - 1 | 1 |
| Gemini | 0 - 2 | 1 |
| All compat | 0 - 2 | 1 |

### MaxTokens

Maximum number of tokens in the response.

```go
maxTokens := 1000
req.MaxTokens = &maxTokens
```

| Provider | Default | Notes |
|----------|---------|-------|
| OpenAI | Model-dependent | Optional |
| Anthropic | 4096 | **Required** (langrails defaults to 4096) |
| Gemini | Model-dependent | Optional |

### TopP (Nucleus Sampling)

Controls diversity by limiting the cumulative probability of tokens considered.

```go
topP := 0.9
req.TopP = &topP  // Range: 0-1
```

### TopK

Limits the number of tokens considered at each step. Only supported by some providers.

```go
topK := 40
req.TopK = &topK
```

| Provider | Supported |
|----------|-----------|
| Anthropic | Yes |
| Gemini | Yes |
| OpenAI | No (ignored) |
| All compat | No (ignored) |

### FrequencyPenalty

Penalizes tokens based on how often they appear in the output so far. Reduces repetition.

```go
fp := 0.5
req.FrequencyPenalty = &fp  // Range: -2 to 2
```

| Provider | Supported |
|----------|-----------|
| OpenAI | Yes |
| All compat | Yes |
| Anthropic | No (ignored) |
| Gemini | No (ignored) |

### PresencePenalty

Penalizes tokens based on whether they appear at all in the output. Encourages new topics.

```go
pp := 0.6
req.PresencePenalty = &pp  // Range: -2 to 2
```

| Provider | Supported |
|----------|-----------|
| OpenAI | Yes |
| All compat | Yes |
| Anthropic | No (ignored) |
| Gemini | No (ignored) |

### Stop Sequences

Sequences where the model should stop generating.

```go
req.Stop = []string{"\n\n", "END", "```"}
```

| Provider | Field Name | Supported |
|----------|------------|-----------|
| OpenAI | `stop` | Yes (up to 4) |
| Anthropic | `stop_sequences` | Yes |
| Gemini | `stopSequences` | Yes |
| All compat | `stop` | Yes |

### Seed

Enables deterministic output. Same seed + same request = same output (best effort).

```go
seed := 42
req.Seed = &seed
```

| Provider | Supported |
|----------|-----------|
| OpenAI | Yes |
| Some compat | Yes |
| Anthropic | No |
| Gemini | No |

### Structured Output (OutputSchema)

Constrain output to match a JSON schema. See [Structured Output](structured-output.md) for details.

```go
schema := []byte(`{"type":"object","properties":{"name":{"type":"string"}}}`)
req.OutputSchema = &schema
```

### Tools

Define functions the model can call. See [Tool Calling](tool-calling.md) for details.

```go
req.Tools = []langrails.ToolDefinition{
    {
        Name:        "get_weather",
        Description: "Get weather for a city",
        Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
    },
}
```

### Thinking Mode

Enable extended thinking / chain-of-thought reasoning.

```go
// Enable thinking
req.Thinking = true

// With budget (Anthropic only)
budget := 10000
req.Thinking = true
req.ThinkingBudget = &budget
```

**Anthropic** — Extended thinking with configurable budget:
```go
provider := anthropic.New("sk-ant-...")
resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:          "claude-sonnet-4-20250514",
    Messages:       messages,
    Thinking:       true,
    ThinkingBudget: &budget,  // Default: 10000 tokens
})

fmt.Println(resp.Thinking) // Internal reasoning (may be long)
fmt.Println(resp.Content)  // Final answer
```

**OpenAI** — Reasoning effort for o-series models:
```go
provider := openai.New("sk-...")
resp, _ := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:    "o1",
    Messages: messages,
    Thinking: true,
    // ThinkingBudget maps to effort: <=1024 → "low", >=16384 → "high", else "medium"
})
```

| Provider | Method | Response Field |
|----------|--------|----------------|
| Anthropic | `thinking` block with `budget_tokens` | `resp.Thinking` |
| OpenAI | `reasoning.effort` (low/medium/high) | Not returned |
| Gemini | Not supported | — |

## Full Example

```go
temp := 0.8
maxTokens := 2000
topP := 0.95
topK := 50
fp := 0.3
pp := 0.3
seed := 42
budget := 15000

resp, err := provider.Complete(ctx, &langrails.CompletionRequest{
    Model:            "gpt-4o",
    SystemPrompt:     "You are a creative writer.",
    Messages:         []langrails.Message{{Role: "user", Content: "Write a story"}},
    Temperature:      &temp,
    MaxTokens:        &maxTokens,
    TopP:             &topP,
    TopK:             &topK,
    FrequencyPenalty: &fp,
    PresencePenalty:  &pp,
    Stop:             []string{"THE END"},
    Seed:             &seed,
    Thinking:         true,
    ThinkingBudget:   &budget,
})
```

## Provider Feature Matrix

| Parameter | OpenAI | Anthropic | Gemini | DeepSeek | Groq | Fireworks | xAI | OpenRouter | Together | Mistral | Cohere |
|-----------|--------|-----------|--------|----------|------|-----------|-----|------------|----------|---------|--------|
| Temperature | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| MaxTokens | Yes | Yes* | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| TopP | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| TopK | - | Yes | Yes | - | - | - | - | - | - | - | - |
| FrequencyPenalty | Yes | - | - | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| PresencePenalty | Yes | - | - | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Stop | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Seed | Yes | - | - | Yes | - | - | - | Varies | Yes | - | - |
| Tools | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Structured Output | Yes | Yes | Yes | Yes | Yes | Varies | Varies | Varies | Varies | Varies | Varies |
| Streaming | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Thinking | Yes** | Yes | - | Yes*** | - | - | - | - | - | - | - |

\* Anthropic requires max_tokens (defaults to 4096)
\** OpenAI uses reasoning effort for o-series models
\*** DeepSeek R1 supports thinking via OpenAI-compatible API
