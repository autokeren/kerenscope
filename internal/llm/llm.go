package llm

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
)

type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

type toolCallWire struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolCallFunc `json:"function"`
}

type toolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (tc ToolCall) MarshalJSON() ([]byte, error) {
	return json.Marshal(toolCallWire{ID: tc.ID, Type: "function", Function: toolCallFunc{Name: tc.Name, Arguments: tc.Arguments}})
}

func (tc *ToolCall) UnmarshalJSON(data []byte) error {
	var w toolCallWire
	if err := json.Unmarshal(data, &w); err != nil {
		return err
	}
	tc.ID = w.ID
	tc.Name = w.Function.Name
	tc.Arguments = w.Function.Arguments
	return nil
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

type ToolDef struct {
	Type     string         `json:"type"`
	Function FunctionDef    `json:"function"`
}

type FunctionDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type Request struct {
	Model           string    `json:"model"`
	Messages        []Message `json:"messages"`
	Tools           []ToolDef `json:"tools,omitempty"`
	MaxTokens       int       `json:"max_tokens,omitempty"`
	Temperature     float64   `json:"temperature,omitempty"`
	ReasoningEffort string    `json:"reasoning_effort,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Response struct {
	Content      string     `json:"content"`
	ToolCalls    []ToolCall `json:"tool_calls"`
	FinishReason string     `json:"finish_reason"`
	Model        string     `json:"model"`
	Usage        Usage      `json:"usage"`
}

type Provider interface {
	Complete(ctx context.Context, req Request) (Response, error)
}

var ErrNoAPIKey = errors.New("LLM API key is not set — configure KERENSCOPE_LLM_* environment variables")

const DefaultMaxTokens = 32768

func MaxTokensHint() int {
	if v := os.Getenv("KERENSCOPE_LLM_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return DefaultMaxTokens
}
