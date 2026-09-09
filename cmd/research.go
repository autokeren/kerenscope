package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/autokeren/kerenscope/internal/agent"
	"github.com/autokeren/kerenscope/internal/llm"
	"github.com/autokeren/kerenscope/internal/tools"
	"github.com/spf13/cobra"
)

var researchCmd = &cobra.Command{
	Use:   "research <question>",
	Short: "Autonomously investigate a research question using Sectors data",
	Args:  cobra.ExactArgs(1),
	RunE:  runResearch,
}

func init() {
	rootCmd.AddCommand(researchCmd)
}

func runResearch(cmd *cobra.Command, args []string) error {
	question := strings.TrimSpace(args[0])
	client, err := newClient()
	if err != nil {
		return err
	}
	provider, err := llm.ConfigFromEnv()
	if err != nil {
		return err
	}
	registry := tools.DefaultRegistry(client)
	ag := agent.New(provider, registry)

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
	defer cancel()

	fmt.Println(title("KerenScope — Autonomous Research"))
	fmt.Println()
	fmt.Printf("  Question: %s\n\n", question)

	report, err := ag.Research(ctx, question, renderEvent)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(report.Render())

	outDir := "reports"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(outDir, report.Slug()+".md")
	if err := os.WriteFile(path, []byte(report.Render()), 0o644); err != nil {
		return err
	}
	fmt.Printf("Report saved to %s\n", path)
	return nil
}

func renderEvent(ev agent.Event) {
	switch e := ev.(type) {
	case agent.PlanEvent:
		fmt.Printf("  Research objective\n  ─────────────────\n  %s\n\n  Plan\n  ────\n", e.Plan.Objective)
		for _, s := range e.Plan.Steps {
			fmt.Printf("  [%d/%d] %s — %s\n", s.ID, len(e.Plan.Steps), s.Tool, s.Intent)
		}
		fmt.Println()
	case agent.ToolDoneEvent:
		status := "✓"
		if !e.OK {
			status = "✗"
		}
		summary := e.Summary
		if len(summary) > 90 {
			summary = summary[:90] + "…"
		}
		fmt.Printf("  → %-18s %s %s\n", e.Tool, status, summary)
	case agent.DraftEvent:
		fmt.Println("\n  ✓ Analysis drafted")
	case agent.VerifyEvent:
		supported := 0
		for _, c := range e.Verification.Claims {
			if c.Supported {
				supported++
			}
		}
		fmt.Printf("\n  Verification: %d/%d claims supported · confidence %d/100\n", supported, len(e.Verification.Claims), e.Verification.Confidence)
		fmt.Printf("  Numeric check: %d matched, %d unmatched (deterministic)\n", len(e.Verification.Numeric.Matched), len(e.Verification.Numeric.Unmatched))
		for _, l := range e.Verification.Limitations {
			fmt.Printf("  ⚠ %s\n", l)
		}
	}
}
