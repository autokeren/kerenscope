package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/autokeren/kerenscope/internal/agent"
	"github.com/autokeren/kerenscope/internal/llm"
	"github.com/autokeren/kerenscope/internal/tools"
	"github.com/autokeren/kerenscope/internal/ui"
	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	autoApprove bool
	revealSlow  bool
)

var researchCmd = &cobra.Command{
	Use:     "research <question>",
	Aliases: []string{"r", "riset", "cari"},
	Short:   "Autonomously investigate a research question using Sectors data",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runResearchQuestion(cmd.Context(), strings.TrimSpace(args[0]))
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&autoApprove, "yes", "y", false, "skip the plan approval prompt")
	rootCmd.AddCommand(researchCmd)
}

func runResearchQuestion(ctx context.Context, question string) error {
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

	ctx, cancel := context.WithTimeout(ctx, 12*time.Minute)
	defer cancel()

	fmt.Println(banner())
	fmt.Println("  " + dim("Question"))
	fmt.Println(wrapIndent(question, contentWidth(), "  "))
	fmt.Println()
	beginPhase("planning investigation")

	report, err := ag.Research(ctx, question, renderEvent, planApprovalPrompt())
	if err != nil {
		stopPhase()
		return err
	}

	fmt.Println()
	revealPrint(renderMarkdown(report.Render()), revealSlow)

	outDir := "reports"
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(outDir, report.Slug()+".md")
	if err := os.WriteFile(path, []byte(report.Render()), 0o644); err != nil {
		return err
	}
	fmt.Printf("  %s %s\n", statusIcon(true), "Report saved to "+path)
	htmlPath := filepath.Join(outDir, report.Slug()+".html")
	if htmlBody, err := report.RenderHTML(); err == nil {
		if err := os.WriteFile(htmlPath, htmlBody, 0o644); err == nil {
			fmt.Printf("  %s %s\n", statusIcon(true), "HTML report saved to "+htmlPath)
		}
	}
	return nil
}

func renderMarkdown(md string) string {
	renderer, err := glamour.NewTermRenderer(glamour.WithWordWrap(contentWidth()), glamour.WithAutoStyle())
	if err != nil {
		return md
	}
	out, err := renderer.Render(md)
	if err != nil {
		return md
	}
	return out
}

func planApprovalPrompt() agent.ApprovePlan {
	if autoApprove || !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil
	}
	reader := bufio.NewReader(os.Stdin)
	return func(plan agent.Plan, attempt int) agent.PlanApproval {
		uiMu.Lock()
		defer uiMu.Unlock()
		fmt.Println()
		fmt.Printf("  Plan %d proposed.  %s · %s · %s  ",
			attempt+1,
			bold("[a]pprove"), yellow("[r]egenerate"), red("[q]uit"))
		line, err := reader.ReadString('\n')
		if err != nil {
			return agent.PlanApprove
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "r", "regenerate":
			fmt.Println("  " + dim("→ regenerating plan…"))
			beginPhaseLocked("regenerating plan")
			return agent.PlanRegenerate
		case "q", "quit", "x":
			return agent.PlanAbort
		default:
			return agent.PlanApprove
		}
	}
}

var (
	uiMu       sync.Mutex
	spin       *ui.Spinner
	phaseStart time.Time
)

func beginPhase(label string) {
	uiMu.Lock()
	defer uiMu.Unlock()
	beginPhaseLocked(label)
}

func beginPhaseLocked(label string) {
	stopPhaseLocked()
	phaseStart = time.Now()
	spin = ui.Start(dim(label), os.Stdout, &uiMu)
}

func stopPhase() {
	uiMu.Lock()
	defer uiMu.Unlock()
	stopPhaseLocked()
}

func stopPhaseLocked() {
	if spin != nil {
		spin.Cancel()
		fmt.Print("\r" + strings.Repeat(" ", 110) + "\r")
		spin = nil
	}
}

func phaseElapsed() time.Duration {
	return time.Since(phaseStart).Round(time.Second)
}

func renderEvent(ev agent.Event) {
	uiMu.Lock()
	defer uiMu.Unlock()
	switch e := ev.(type) {
	case agent.PlanEvent:
		stopPhaseLocked()
		fmt.Println(section("Research objective"))
		fmt.Println(wrapIndent(e.Plan.Objective, contentWidth()-4, "    "))
		fmt.Println(section(fmt.Sprintf("Plan · %d steps", len(e.Plan.Steps))))
		for _, s := range e.Plan.Steps {
			fmt.Printf("    %s %s%s\n",
				dim(fmt.Sprintf("%2d.", s.ID)),
				cyan(fmt.Sprintf("%-18s", s.Tool)),
				wrapFirst(s.Intent, contentWidth()-30))
		}
		fmt.Println("    " + dim(fmt.Sprintf("(planned in %s)", phaseElapsed())))
	case agent.ToolDoneEvent:
		stopPhaseLocked()
		summary := e.Summary
		if max := contentWidth() - 30; len(summary) > max && max > 30 {
			summary = summary[:max] + "…"
		}
		fmt.Printf("    %s %s%s\n",
			statusIcon(e.OK),
			cyan(fmt.Sprintf("%-18s", e.Tool)),
			dim(summary))
	case agent.ThinkingEvent:
		beginPhaseLocked(e.Note)
	case agent.DraftEvent:
		stopPhaseLocked()
		fmt.Printf("\n  %s %s %s\n", green("✓"), bold("Analysis drafted"), dim("("+phaseElapsed().String()+")"))
	case agent.VerifyEvent:
		stopPhaseLocked()
		supported := 0
		for _, c := range e.Verification.Claims {
			if c.Supported {
				supported++
			}
		}
		fmt.Println(section("Verification"))
		fmt.Printf("    %s %d/%d claims supported  ·  confidence %s\n",
			statusIcon(supported*2 >= len(e.Verification.Claims)),
			supported, len(e.Verification.Claims),
			bold(fmt.Sprintf("%d/100", e.Verification.Confidence)))
		fmt.Printf("    %s %d numbers matched deterministically · %d adjudicated\n",
			statusIcon(true),
			len(e.Verification.Numeric.Matched), len(e.Verification.Numeric.Unmatched))
		for _, l := range e.Verification.Limitations {
			fmt.Println(wrapIndent("⚠ "+l, contentWidth()-4, "    "))
		}
	}
}

func wrapFirst(s string, width int) string {
	lines := strings.Split(wrap(s, width, ""), "\n")
	for i, line := range lines {
		if i == 0 {
			lines[i] = line
		} else {
			lines[i] = "                        " + line
		}
	}
	return strings.Join(lines, "\n")
}

func revealPrint(s string, slow bool) {
	s = strings.TrimRight(s, "\n")
	if !colorOn {
		fmt.Println(s)
		return
	}
	delay := 25 * time.Millisecond
	if slow {
		delay = 60 * time.Millisecond
	}
	for _, line := range strings.Split(s, "\n") {
		fmt.Println(line)
		time.Sleep(delay)
	}
}
