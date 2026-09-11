package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestToolCallWireFormatParses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"choices":[{"finish_reason":"tool_calls","message":{"content":"","tool_calls":[{"id":"call_1","type":"function","function":{"name":"submit_plan","arguments":"{\"objective\":\"o\"}"}}]}}]}`))
	}))
	defer srv.Close()
	p := &OpenAICompat{BaseURL: srv.URL, APIKey: "k", Model: "m", Client: srv.Client()}
	resp, err := p.Complete(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "submit_plan" || resp.ToolCalls[0].ID != "call_1" {
		t.Fatalf("bad tool call: %+v", resp.ToolCalls[0])
	}
	if resp.ToolCalls[0].Arguments != `{"objective":"o"}` {
		t.Fatalf("bad arguments: %s", resp.ToolCalls[0].Arguments)
	}
	if resp.FinishReason != "tool_calls" {
		t.Fatalf("expected finish_reason tool_calls, got %s", resp.FinishReason)
	}
}

func TestToolCallsMarshalBackAsNestedWireFormat(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		msgs := body["messages"].([]any)
		last := msgs[len(msgs)-1].(map[string]any)
		data, _ := json.Marshal(last["tool_calls"])
		got = string(data)
		w.Write([]byte(`{"choices":[{"message":{"content":"done"}}]}`))
	}))
	defer srv.Close()
	p := &OpenAICompat{BaseURL: srv.URL, APIKey: "k", Model: "m", Client: srv.Client()}
	_, err := p.Complete(context.Background(), Request{Messages: []Message{{Role: "assistant", ToolCalls: []ToolCall{{ID: "call_9", Name: "t", Arguments: "{}"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"name":"t"`) || !strings.Contains(got, `"arguments":"{}"`) || !strings.Contains(got, `"function"`) {
		t.Fatalf("assistant tool_calls must marshal as nested wire format, got: %s", got)
	}
	if !strings.Contains(got, `"id":"call_9"`) {
		t.Fatalf("expected call id in wire format, got: %s", got)
	}
}

func TestFallsBackToSecondaryModel(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(raw))
		json.Unmarshal(raw, &body)
		if body["model"] == "primary-model" {
			w.WriteHeader(http.StatusRequestTimeout)
			w.Write([]byte(`{"error":{"message":"request timeout"}}`))
			return
		}
		w.Write([]byte(`{"choices":[{"message":{"content":"flash saved the day"}}]}`))
	}))
	defer srv.Close()
	p := &OpenAICompat{BaseURL: srv.URL, APIKey: "k", Model: "primary-model", FallbackModel: "flash-model", Client: srv.Client()}
	resp, err := p.Complete(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "flash saved the day" {
		t.Fatalf("expected fallback response, got %q", resp.Content)
	}
	if len(bodies) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(bodies))
	}
	if !strings.Contains(bodies[0], `"primary-model"`) {
		t.Fatal("first attempt must use the primary model")
	}
	if !strings.Contains(bodies[1], `"flash-model"`) {
		t.Fatal("second attempt must switch to the fallback model")
	}
}

func TestNoFallbackLoopWhenSameModel(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()
	p := &OpenAICompat{BaseURL: srv.URL, APIKey: "k", Model: "m", FallbackModel: "m", Client: srv.Client()}
	_, err := p.Complete(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}})
	if err == nil {
		t.Fatal("expected error to propagate")
	}
	if calls != 1 {
		t.Fatalf("must not retry when fallback equals primary model, got %d calls", calls)
	}
}

func TestDemoHeaderAlwaysPresent(t *testing.T) {
	var gotDemoHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotDemoHeader = r.Header.Get("X-KerenScope-Demo")
		w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()
	p := &OpenAICompat{BaseURL: srv.URL, APIKey: "k", Model: "m", Client: srv.Client()}
	if _, err := p.Complete(context.Background(), Request{Messages: []Message{{Role: "user", Content: "hi"}}}); err != nil {
		t.Fatal(err)
	}
	if gotDemoHeader != "1" {
		t.Fatalf("expected X-KerenScope-Demo header, got %q", gotDemoHeader)
	}
}
