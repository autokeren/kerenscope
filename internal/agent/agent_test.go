package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/autokeren/kerenscope/internal/llm"
	"github.com/autokeren/kerenscope/internal/tools"
)

type scriptedProvider struct {
	responses []llm.Response
	calls     int
	mu        sync.Mutex
}

func (p *scriptedProvider) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.calls >= len(p.responses) {
		return llm.Response{Content: "done"}, nil
	}
	r := p.responses[p.calls]
	p.calls++
	return r, nil
}

func planResponse(objective string, steps ...PlanStep) llm.Response {
	plan := Plan{Objective: objective, Steps: steps}
	args, _ := json.Marshal(plan)
	return llm.Response{ToolCalls: []llm.ToolCall{{ID: "p1", Name: "submit_plan", Arguments: string(args)}}}
}

func toolCallResponse(calls ...llm.ToolCall) llm.Response {
	return llm.Response{ToolCalls: calls}
}

func finalResponse(content string) llm.Response {
	return llm.Response{Content: content}
}

type slowTool struct {
	name string
	delay time.Duration
}

func (t slowTool) Definition() tools.Definition {
	return tools.Definition{Name: t.name, Description: "test tool", Parameters: map[string]any{"type": "object", "properties": map[string]any{}}}
}

func (t slowTool) Run(ctx context.Context, args map[string]any) tools.Result {
	time.Sleep(t.delay)
	return tools.Result{OK: true, Data: map[string]any{"tool": t.name}}
}

func TestExecutorRunsToolCallsInParallel(t *testing.T) {
	registry := tools.NewRegistry().
		Register(slowTool{name: "t1", delay: 150 * time.Millisecond}).
		Register(slowTool{name: "t2", delay: 150 * time.Millisecond}).
		Register(slowTool{name: "t3", delay: 150 * time.Millisecond})
	provider := &scriptedProvider{responses: []llm.Response{
		planResponse("test objective", PlanStep{ID: 1, Intent: "i", Tool: "t1", Args: map[string]any{}}),
		toolCallResponse(
			llm.ToolCall{ID: "c1", Name: "t1", Arguments: "{}"},
			llm.ToolCall{ID: "c2", Name: "t2", Arguments: "{}"},
			llm.ToolCall{ID: "c3", Name: "t3", Arguments: "{}"},
		),
		finalResponse("final draft"),
	}}
	ag := New(provider, registry)
	start := time.Now()
	report, err := ag.Research(context.Background(), "test question", nil, nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed >= 450*time.Millisecond {
		t.Fatalf("3×150ms tools completed in %v — they ran sequentially", elapsed)
	}
	if elapsed < 150*time.Millisecond {
		t.Fatalf("completed in %v — suspiciously fast, tools may not have run", elapsed)
	}
	if report.Draft != "final draft" {
		t.Fatalf("unexpected draft: %q", report.Draft)
	}
	if len(report.ToolsUsed) != 3 {
		t.Fatalf("expected 3 tools used, got %v", report.ToolsUsed)
	}
}

func TestPlanApprovalHook(t *testing.T) {
	registry := tools.NewRegistry().Register(slowTool{name: "t1"})
	provider := &scriptedProvider{responses: []llm.Response{
		planResponse("objective one", PlanStep{ID: 1, Intent: "i", Tool: "t1", Args: map[string]any{}}),
		planResponse("objective two", PlanStep{ID: 1, Intent: "i", Tool: "t1", Args: map[string]any{}}),
		toolCallResponse(llm.ToolCall{ID: "c1", Name: "t1", Arguments: "{}"}),
		finalResponse("approved draft"),
	}}
	ag := New(provider, registry)
	var approvals int
	report, err := ag.Research(context.Background(), "q", nil, func(plan Plan, attempt int) PlanApproval {
		approvals++
		if plan.Objective != "objective two" {
			return PlanRegenerate
		}
		return PlanApprove
	})
	if err != nil {
		t.Fatal(err)
	}
	if approvals != 2 {
		t.Fatalf("expected 2 approval prompts, got %d", approvals)
	}
	if report.Plan.Objective != "objective two" {
		t.Fatalf("expected regenerated plan to be used, got %q", report.Plan.Objective)
	}
}

func TestPlanAbort(t *testing.T) {
	registry := tools.NewRegistry().Register(slowTool{name: "t1"})
	provider := &scriptedProvider{responses: []llm.Response{
		planResponse("objective", PlanStep{ID: 1, Intent: "i", Tool: "t1", Args: map[string]any{}}),
	}}
	ag := New(provider, registry)
	_, err := ag.Research(context.Background(), "q", nil, func(plan Plan, attempt int) PlanApproval {
		return PlanAbort
	})
	if err == nil {
		t.Fatal("expected abort error")
	}
	if fmt.Sprintf("%v", err) == "" {
		t.Fatal("expected non-empty error")
	}
}
