package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type OpenAICompat struct {
	BaseURL         string
	APIKey          string
	Model           string
	Client          *http.Client
	Effort          string
}

func (p *OpenAICompat) reasoningEffort() string {
	if p.Effort != "" {
		return p.Effort
	}
	if strings.Contains(strings.ToLower(p.Model), "glm") {
		return "high"
	}
	return ""
}

func ConfigFromEnv() (Provider, error) {
	base := strings.TrimRight(os.Getenv("KERENSCOPE_LLM_BASE_URL"), "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	key := os.Getenv("KERENSCOPE_LLM_API_KEY")
	if key == "" {
		key = os.Getenv("OPENAI_API_KEY")
	}
	model := os.Getenv("KERENSCOPE_LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	if key == "" {
		return nil, ErrNoAPIKey
	}
	return &OpenAICompat{BaseURL: base, APIKey: key, Model: model, Client: &http.Client{Timeout: 5 * time.Minute}, Effort: os.Getenv("KERENSCOPE_LLM_REASONING")}, nil
}

type chatRequest struct {
	Model           string    `json:"model"`
	Messages        []Message `json:"messages"`
	Tools           []ToolDef `json:"tools,omitempty"`
	MaxTokens       int       `json:"max_tokens,omitempty"`
	Temperature     float64   `json:"temperature,omitempty"`
	ReasoningEffort string    `json:"reasoning_effort,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage Usage `json:"usage"`
	Model string `json:"model"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

func CallTimeout() time.Duration {
	if v := os.Getenv("KERENSCOPE_LLM_CALL_TIMEOUT_SECS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 4 * time.Minute
}

func (p *OpenAICompat) Complete(ctx context.Context, req Request) (Response, error) {
	callCtx, cancel := context.WithTimeout(ctx, CallTimeout())
	defer cancel()
	ctx = callCtx
	if req.Model == "" {
		req.Model = p.Model
	}
	if req.ReasoningEffort == "" {
		req.ReasoningEffort = p.reasoningEffort()
	}
	body, err := json.Marshal(chatRequest{
		Model:           req.Model,
		Messages:        req.Messages,
		Tools:           req.Tools,
		MaxTokens:       req.MaxTokens,
		Temperature:     req.Temperature,
		ReasoningEffort: req.ReasoningEffort,
	})
	if err != nil {
		return Response{}, err
	}
	target := p.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	resp, err := p.Client.Do(httpReq)
	if err != nil {
		return Response{}, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return Response{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Response{}, fmt.Errorf("llm: HTTP %d: %s", resp.StatusCode, trim(string(data), 500))
	}
	var out chatResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return Response{}, fmt.Errorf("llm: invalid response JSON: %w", err)
	}
	if out.Error != nil && out.Error.Message != "" {
		return Response{}, fmt.Errorf("llm: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return Response{}, fmt.Errorf("llm: no choices in response")
	}
	choice := out.Choices[0]
	return Response{
		Content:      choice.Message.Content,
		ToolCalls:    normalizeToolCalls(choice.Message.ToolCalls),
		FinishReason: choice.FinishReason,
		Model:        out.Model,
		Usage:        out.Usage,
	}, nil
}

func normalizeToolCalls(calls []ToolCall) []ToolCall {
	out := make([]ToolCall, 0, len(calls))
	for _, call := range calls {
		name := strings.TrimSpace(call.Name)
		if name == "" {
			continue
		}
		args := strings.TrimSpace(call.Arguments)
		if args == "" {
			args = "{}"
		}
		if call.ID == "" {
			call.ID = fmt.Sprintf("call_%d", time.Now().UnixNano())
		}
		out = append(out, ToolCall{ID: call.ID, Name: name, Arguments: args})
	}
	return out
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
