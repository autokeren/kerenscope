package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/autokeren/kerenscope/internal/llm"
	"github.com/autokeren/kerenscope/internal/tools"
	"github.com/autokeren/kerenscope/internal/verify"
)

type PlanStep struct {
	ID     int            `json:"id"`
	Intent string         `json:"intent"`
	Tool   string         `json:"tool"`
	Args   map[string]any `json:"args"`
}

type Plan struct {
	Objective string     `json:"objective"`
	Steps     []PlanStep `json:"steps"`
}

type Claim struct {
	Claim    string `json:"claim"`
	Supported bool   `json:"supported"`
	Evidence  string `json:"evidence"`
}

type Verification struct {
	Claims      []Claim `json:"claims"`
	Confidence  int     `json:"confidence"`
	Limitations []string `json:"limitations"`
	Numeric     verify.NumericCheck `json:"numeric"`
}

type Report struct {
	Question     string
	Plan         Plan
	Draft        string
	Verification Verification
	ToolsUsed    []string
	StartedAt    time.Time
	EndedAt      time.Time
}

type PlanApproval int

const (
	PlanApprove PlanApproval = iota
	PlanRegenerate
	PlanAbort
)

type ApprovePlan func(plan Plan, attempt int) PlanApproval

type Event interface{}

func (a *Agent) completeWithFallback(ctx context.Context, emit func(Event), req llm.Request) (llm.Response, error) {
	resp, err := a.LLM.Complete(ctx, req)
	if err == nil || ctx.Err() != nil {
		return resp, err
	}
	retry := req
	retry.ReasoningEffort = "low"
	if emit != nil {
		emit(ThinkingEvent{Note: "call failed (" + truncate(err.Error(), 60) + ") — retrying at low reasoning effort"})
	}
	return a.LLM.Complete(ctx, retry)
}

type PlanEvent struct{ Plan Plan }
type StepEvent struct {
	Index int
	Total int
	Intent string
	Tool string
	Args json.RawMessage
}
type ToolDoneEvent struct {
	Index int
	OK bool
	Tool string
	Summary string
}
type DraftEvent struct{ Draft string }
type VerifyEvent struct{ Verification Verification }
type ThinkingEvent struct{ Note string }

type evidence struct {
	Tool   string
	Args   map[string]any
	Result string
}

type Agent struct {
	LLM              llm.Provider
	Tools            *tools.Registry
	MaxActTurns      int
	MaxToolCalls     int
	MaxResultChars   int
	EvidenceDigestChars int
	Temperature      float64
}

func New(llmProvider llm.Provider, registry *tools.Registry) *Agent {
	return &Agent{
		LLM:                 llmProvider,
		Tools:               registry,
		MaxActTurns:         20,
		MaxToolCalls:        18,
		MaxResultChars:      8000,
		EvidenceDigestChars: 4000,
		Temperature:         0.2,
	}
}

func (a *Agent) Research(ctx context.Context, question string, emit func(Event), approve ApprovePlan) (*Report, error) {
	if a.LLM == nil {
		return nil, fmt.Errorf("agent: LLM provider is nil")
	}
	if a.Tools == nil {
		return nil, fmt.Errorf("agent: tool registry is nil")
	}
	report := &Report{Question: question, StartedAt: time.Now().UTC()}

	plan, err := a.makePlanApproved(ctx, question, emit, approve)
	if err != nil {
		return nil, fmt.Errorf("planning failed: %w", err)
	}
	report.Plan = plan

	state := newComputeState()
	draft, used, evidenceStore, err := a.execute(ctx, question, plan, emit, state)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	report.Draft = draft
	report.ToolsUsed = used
	if emit != nil {
		emit(DraftEvent{Draft: draft})
	}

	verification, err := a.verify(ctx, question, draft, evidenceStore, state, emit)
	if err != nil {
		verification.Limitations = append(verification.Limitations, "verification step failed: "+err.Error())
	}
	report.Verification = verification
	if emit != nil {
		emit(VerifyEvent{Verification: verification})
	}

	report.EndedAt = time.Now().UTC()
	return report, nil
}

func (a *Agent) makePlanApproved(ctx context.Context, question string, emit func(Event), approve ApprovePlan) (Plan, error) {
	var feedback string
	for attempt := 0; attempt < 3; attempt++ {
		plan, err := a.makePlan(ctx, question, feedback)
		if err != nil {
			return Plan{}, err
		}
		if emit != nil {
			emit(PlanEvent{Plan: plan})
		}
		if approve == nil {
			return plan, nil
		}
		switch approve(plan, attempt) {
		case PlanApprove:
			return plan, nil
		case PlanAbort:
			return Plan{}, fmt.Errorf("research aborted by user before execution")
		default:
			feedback = "The user rejected this plan. Produce a different research approach."
		}
	}
	return Plan{}, fmt.Errorf("plan approval failed after 3 attempts")
}

func (a *Agent) makePlan(ctx context.Context, question string, feedback string) (Plan, error) {
	system := "You are the research planner of KerenScope, an autonomous financial research agent for the Indonesian stock market. Today is " +
		time.Now().UTC().Format("2006-01-02") + ".\n\n" +
		"Given the user's research question, produce a focused research plan of 2-8 steps. " +
		"Each step must use exactly one of the available tools listed below, with concrete arguments. " +
		"Prefer the cheapest sufficient tool per step; use screen_companies for open-ended universe questions ('which stocks are worth looking at') with concrete where/order_by filters, then deep-dive only the top 2-4 candidates with company_report, price_history and foreign_flow. " +
		"company_report for a specific ticker, subsector_report for peer context, and price_history/foreign_flow/broker_summary/insider_filings/news for signal checks. " +
		"Do not include writing or analysis steps — only data-gathering steps. Call the submit_plan tool exactly once.\n\n" +
		"Available tools:\n" + a.toolCatalog()
	if feedback != "" {
		system += "\n\n" + feedback
	}
	plan, err := a.callSubmitPlan(ctx, system, question, nil)
	if err != nil {
		return Plan{}, err
	}
	if validated, ok, verr := a.validatePlan(plan); !ok {
		retry := system + "\n\nYour previous plan was rejected: " + verr + ". Fix it and call submit_plan again."
		plan, err = a.callSubmitPlan(ctx, retry, question, nil)
		if err != nil {
			return Plan{}, err
		}
		validated, ok, verr = a.validatePlan(plan)
		if !ok {
			return Plan{}, fmt.Errorf("invalid research plan: %s", verr)
		}
		return validated, nil
	} else {
		return validated, nil
	}
}

func (a *Agent) callSubmitPlan(ctx context.Context, system, question string, emit func(Event)) (Plan, error) {
	resp, err := a.completeWithFallback(ctx, emit, llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: question},
		},
		Tools: []llm.ToolDef{submitPlanTool()},
		MaxTokens: llm.MaxTokensHint(),
		Temperature: a.Temperature,
	})
	if err != nil {
		return Plan{}, err
	}
	var plan Plan
	for _, call := range resp.ToolCalls {
		if call.Name != "submit_plan" {
			continue
		}
		if err := json.Unmarshal([]byte(trimTrailingJSON(call.Arguments)), &plan); err != nil {
			return Plan{}, fmt.Errorf("submit_plan arguments must be valid JSON: %w", err)
		}
		return plan, nil
	}
	return Plan{}, fmt.Errorf("model did not call submit_plan (finish=%s reply: %s)", resp.FinishReason, truncate(resp.Content, 200))
}

func (a *Agent) validatePlan(plan Plan) (Plan, bool, string) {
	if strings.TrimSpace(plan.Objective) == "" {
		return plan, false, "objective is empty"
	}
	if len(plan.Steps) == 0 {
		return plan, false, "no steps"
	}
	if len(plan.Steps) > 8 {
		plan.Steps = plan.Steps[:8]
	}
	for i := range plan.Steps {
		step := &plan.Steps[i]
		if step.ID == 0 {
			step.ID = i + 1
		}
		if strings.TrimSpace(step.Tool) == "" {
			return plan, false, fmt.Sprintf("step %d has no tool", step.ID)
		}
		if _, ok := a.Tools.Get(step.Tool); !ok {
			return plan, false, fmt.Sprintf("step %d references unknown tool %q", step.ID, step.Tool)
		}
		if step.Args == nil {
			step.Args = map[string]any{}
		}
	}
	return plan, true, ""
}

func (a *Agent) execute(ctx context.Context, question string, plan Plan, emit func(Event), state *computeState) (string, []string, []evidence, error) {
	planJSON, _ := json.MarshalIndent(plan, "", "  ")
	system := "You are KerenScope, an autonomous financial research agent for the Indonesian stock market. Today is " +
		time.Now().UTC().Format("2006-01-02") + ".\n\n" +
		"You are executing this approved research plan, step by step, using the available tools:\n" + string(planJSON) + "\n\n" +
		"Rules:\n" +
		"- Call the tools in plan order. Pass concrete arguments. If a tool fails or returns no data, adapt (use a different tool or adjust arguments) instead of giving up.\n" +
		"- When all the data you need is gathered, STOP calling tools and write your final analysis draft in Markdown.\n" +
		"- The draft must: answer the research question directly, cite concrete numbers from the data you saw (state them explicitly), and note data limitations honestly. Write in the same language as the question.\n" +
		"- Tool results may end with [COMPUTED ... METRICS] or [COMPARISON] blocks. Those are deterministic values computed by the KerenScope engine: cite them verbatim and never do arithmetic yourself.\n" +
		"- If the user asks for buy/sell advice (e.g. 'layak dibeli', 'saham apa yang bagus', 'should I buy'), do NOT refuse and do NOT recommend. Reframe: run a transparent criteria-based screen (fundamentals, valuation vs peers, smart-money signals), present the top candidates with evidence, strengths AND risks per candidate, and state that this is analysis, not a buy recommendation.\n" +
		"- You are an information and analysis tool. Never give investment advice or buy/sell recommendations.\n\n" +
		"Available tools:\n" + a.toolCatalog()

	messages := []llm.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: "Research question: " + question},
	}
	var used []string
	var store []evidence
	toolCallCount := 0
	budgetRejections := 0
	dbg := os.Getenv("KEREN_DEBUG") != ""
	for turn := 0; turn < a.MaxActTurns; turn++ {
		if emit != nil {
			emit(ThinkingEvent{Note: "analyzing gathered data"})
		}
		dbgStart := time.Now()
		resp, err := a.completeWithFallback(ctx, emit, llm.Request{
			Messages: messages,
			Tools:    llmTools(a.Tools),
			MaxTokens: llm.MaxTokensHint(),
			Temperature: a.Temperature,
		})
		if os.Getenv("KEREN_DEBUG") != "" {
			fmt.Fprintf(os.Stderr, "[dbg] executor turn %d: %s, tool_calls=%d, content=%d chars, tokens=%d\n", turn, time.Since(dbgStart).Round(time.Second), len(resp.ToolCalls), len(resp.Content), resp.Usage.CompletionTokens)
		}
		if err != nil {
			return "", used, store, err
		}
		if len(resp.ToolCalls) == 0 {
			return resp.Content, used, store, nil
		}
		if toolCallCount+len(resp.ToolCalls) > a.MaxToolCalls {
			budgetRejections++
			messages = append(messages, llm.Message{Role: "assistant", Content: resp.Content, ToolCalls: resp.ToolCalls})
			for _, call := range resp.ToolCalls {
				content := "Skipped: tool budget exceeded. Write the final analysis draft now with the data you already have."
				if budgetRejections >= 2 {
					content = "Skipped: tool budget exceeded. You MUST NOT request any more tools. Your next message must be the complete final analysis draft in Markdown."
				}
				messages = append(messages, llm.Message{Role: "tool", ToolCallID: call.ID, Name: call.Name, Content: content})
			}
			if budgetRejections >= 3 {
				if strings.TrimSpace(resp.Content) != "" {
					return resp.Content, used, store, nil
				}
				return "Analysis incomplete: the tool budget was exhausted before enough data could be gathered for a confident conclusion.", used, store, nil
			}
			continue
		}
		assistant := llm.Message{Role: "assistant", Content: resp.Content, ToolCalls: resp.ToolCalls}
		messages = append(messages, assistant)
		type executed struct {
			call      llm.ToolCall
			args      map[string]any
			result    tools.Result
			resultStr string
		}
		batch := make([]executed, len(resp.ToolCalls))
		sem := make(chan struct{}, 3)
		var wg sync.WaitGroup
		for i, call := range resp.ToolCalls {
			toolCallCount++
			used = append(used, call.Name)
			var args map[string]any
			if err := json.Unmarshal([]byte(trimTrailingJSON(call.Arguments)), &args); err != nil {
				args = map[string]any{}
			}
			wg.Add(1)
			go func(i int, call llm.ToolCall, args map[string]any) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				dbgT0 := time.Now()
				result := a.Tools.Run(ctx, call.Name, args)
				if dbg {
					fmt.Fprintf(os.Stderr, "[dbg] tool %s: %s\n", call.Name, time.Since(dbgT0).Round(time.Millisecond))
				}
				resultJSON, _ := json.Marshal(result)
				resultStr := trimResult(string(resultJSON), a.MaxResultChars)
				if content, replaced := a.postProcess(call.Name, args, result, state); replaced {
					resultStr = content
				} else if content != "" {
					resultStr += content
				}
				batch[i] = executed{call: call, args: args, result: result, resultStr: resultStr}
				if emit != nil {
					emit(ToolDoneEvent{Index: i + 1, OK: result.OK, Tool: call.Name, Summary: resultSummary(call.Name, args, result)})
				}
			}(i, call, args)
		}
		wg.Wait()
		for _, ex := range batch {
			store = append(store, evidence{Tool: ex.call.Name, Args: ex.args, Result: trimResult(mustJSON(ex.result), a.MaxResultChars)})
			messages = append(messages, llm.Message{Role: "tool", ToolCallID: ex.call.ID, Name: ex.call.Name, Content: ex.resultStr})
		}
	}
	return "", used, store, fmt.Errorf("executor reached max turns (%d) without a final draft", a.MaxActTurns)
}

func (a *Agent) verify(ctx context.Context, question, draft string, store []evidence, state *computeState, emit func(Event)) (Verification, error) {
	raws := make([]string, len(store))
	for i, ev := range store {
		raws[i] = ev.Result
	}
	facts := verify.CollectFacts(raws...)
	facts = append(facts, state.facts...)
	numeric := verify.CheckNumbers(draft, facts)

	var b strings.Builder
	b.WriteString("Research question: " + question + "\n\nFinal analysis draft:\n" + draft + "\n\n")
	if len(numeric.Matched) > 0 || len(numeric.Unmatched) > 0 {
		b.WriteString(fmt.Sprintf("Deterministic numeric check already performed by the engine: %d draft numbers matched the evidence; %d could not be matched.\n", len(numeric.Matched), len(numeric.Unmatched)))
		b.WriteString("Matched numbers (do NOT re-verify these):\n")
		for _, m := range numeric.Matched {
			b.WriteString(fmt.Sprintf("  %s -> %s\n", m.Number, m.Source))
		}
		b.WriteString("Unmatched numbers (adjudicate these: derived/rounded from evidence = supported with explanation; absent from evidence = unsupported):\n")
		for _, u := range numeric.Unmatched {
			b.WriteString("  " + u + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Evidence gathered (per tool call):\n")
	for i, ev := range store {
		b.WriteString(fmt.Sprintf("\n[%d] tool=%s args=%s\nresult: %s\n", i+1, ev.Tool, mustJSON(ev.Args), trimResult(ev.Result, a.EvidenceDigestChars)))
	}
	system := "You are the verification unit of KerenScope, an autonomous financial research agent for the Indonesian stock market. " +
		"A deterministic numeric check has already matched draft numbers against the evidence. Your job: verify the 10-15 most important claims. " +
		"For semantic claims (trends, rankings, causal statements, conclusions), check them against the evidence. " +
		"Numbers listed as matched are already supported. For unmatched numbers, decide whether they are derived from the evidence (supported) or absent (unsupported). " +
		"Call the submit_verification tool exactly once with: the claims (claim text, supported true/false, specific evidence), " +
		"an overall confidence score 0-100, and honest limitations (missing data, unverified assumptions, stale data). " +
		"Be strict but concise: do not over-reason — verify directly."
	if emit != nil {
		emit(ThinkingEvent{Note: "verifying claims against evidence"})
	}
	resp, err := a.completeWithFallback(ctx, emit, llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: b.String()},
		},
		Tools: []llm.ToolDef{submitVerificationTool()},
		MaxTokens: llm.MaxTokensHint(),
		Temperature: 0.1,
	})
	if err != nil {
		verification := Verification{Numeric: numeric}
		verification.Limitations = []string{"LLM verification call failed: " + err.Error()}
		return verification, nil
	}
	for _, call := range resp.ToolCalls {
		if call.Name != "submit_verification" {
			continue
		}
		var v Verification
		if err := json.Unmarshal([]byte(trimTrailingJSON(call.Arguments)), &v); err != nil {
			return Verification{Numeric: numeric}, fmt.Errorf("submit_verification arguments must be valid JSON: %w", err)
		}
		v.Numeric = numeric
		return v, nil
	}
	return Verification{Numeric: numeric}, fmt.Errorf("model did not call submit_verification (finish=%s reply: %s)", resp.FinishReason, truncate(resp.Content, 200))
}

func llmTools(registry *tools.Registry) []llm.ToolDef {
	defs := registry.Definitions()
	out := make([]llm.ToolDef, 0, len(defs))
	for _, d := range defs {
		out = append(out, llm.ToolDef{Type: "function", Function: llm.FunctionDef{Name: d.Name, Description: d.Description, Parameters: d.Parameters}})
	}
	return out
}

func (a *Agent) toolCatalog() string {
	var b strings.Builder
	for _, def := range a.Tools.Definitions() {
		params, _ := json.Marshal(def.Parameters)
		b.WriteString(fmt.Sprintf("- %s: %s\n  parameters: %s\n", def.Name, def.Description, string(params)))
	}
	return b.String()
}

func submitPlanTool() llm.ToolDef {
	return llm.ToolDef{
		Type: "function",
		Function: llm.FunctionDef{
			Name:        "submit_plan",
			Description: "Submit the research plan for the question",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"objective": map[string]any{"type": "string", "description": "One sentence research objective"},
					"steps": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"id":     map[string]any{"type": "integer"},
								"intent": map[string]any{"type": "string", "description": "What this step investigates"},
								"tool":   map[string]any{"type": "string", "description": "Tool name from the catalog"},
								"args":   map[string]any{"type": "object", "description": "Arguments for the tool"},
							},
							"required": []any{"id", "intent", "tool", "args"},
						},
					},
				},
				"required": []any{"objective", "steps"},
			},
		},
	}
}

func submitVerificationTool() llm.ToolDef {
	return llm.ToolDef{
		Type: "function",
		Function: llm.FunctionDef{
			Name:        "submit_verification",
			Description: "Submit claim-by-claim verification of the analysis draft against the evidence",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"claims": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"claim":    map[string]any{"type": "string"},
								"supported": map[string]any{"type": "boolean"},
								"evidence": map[string]any{"type": "string"},
							},
							"required": []any{"claim", "supported", "evidence"},
						},
					},
					"confidence":  map[string]any{"type": "integer", "description": "0-100 overall confidence"},
					"limitations": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				},
				"required": []any{"claims", "confidence"},
			},
		},
	}
}

func trimResult(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "... [truncated]"
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func mustJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "?"
	}
	return string(data)
}

func trimTrailingJSON(s string) string {
	s = strings.TrimSpace(s)
	for len(s) > 0 {
		var probe any
		if err := json.Unmarshal([]byte(s), &probe); err == nil {
			return s
		}
		if s[len(s)-1] == '}' || s[len(s)-1] == ']' || s[len(s)-1] == ',' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}
