package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/autokeren/kerenscope/internal/llm"
	"github.com/autokeren/kerenscope/internal/sectors"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose the setup: Sectors API key, LLM provider, cache and ledger",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(banner())
		fmt.Println(section("Doctor"))
		fmt.Printf("    %s\n", dim("version "+cmd.Root().Version))
		fmt.Println()

		sectorsOK := checkSectors(cmd)
		fmt.Println()
		llmOK := checkLLMProvider(cmd)
		fmt.Println()
		checkLocal()

		fmt.Println()
		fmt.Println(section("Status"))
		if !sectorsOK {
			fmt.Printf("    %s core commands blocked — fix the Sectors key first\n", red("✗"))
			os.Exit(1)
		}
		fmt.Printf("    %s company/credits ready\n", green("✓"))
		if llmOK {
			fmt.Printf("    %s research/compare ready\n", green("✓"))
		} else {
			fmt.Printf("    %s research/compare needs a working LLM key (see above)\n", yellow("⚠"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func checkSectors(cmd *cobra.Command) bool {
	if os.Getenv("SECTORS_API_KEY") == "" {
		fmt.Printf("    %s SECTORS_API_KEY not set\n", red("✗"))
		fmt.Printf("      %s create one at sectors.app/api (API Key Management), then:\n", dim("→"))
		fmt.Printf("      %s\n", bold(`export SECTORS_API_KEY="..."`))
		return false
	}
	client, err := sectors.New(sectors.Options{})
	if err != nil {
		fmt.Printf("    %s %v\n", red("✗"), err)
		return false
	}
	fmt.Printf("    %s SECTORS_API_KEY set — pinging the API…\n", statusIcon(true))
	start := time.Now()
	if _, err := client.Subsectors(cmd.Context()); err != nil {
		fmt.Printf("    %s Sectors API unreachable: %s\n", red("✗"), err)
		fmt.Printf("      %s check the key or api.sectors.app status\n", dim("→"))
		return false
	}
	fmt.Printf("    %s Sectors API reachable %s\n", green("✓"), dim(fmt.Sprintf("(%s — first call costs 1 credit, then it is cached)", time.Since(start).Round(time.Millisecond))))
	return true
}

func checkLLMProvider(cmd *cobra.Command) bool {
	provider, err := llm.ConfigFromEnv()
	if err != nil {
		if os.Getenv("KERENSCOPE_LLM_DEMO") == "0" {
			fmt.Printf("    %s LLM key not set and hosted demo disabled (KERENSCOPE_LLM_DEMO=0)\n", red("✗"))
			fmt.Printf("      %s research/compare need any OpenAI-compatible provider, e.g.:\n", dim("→"))
			fmt.Printf("      %s\n", bold(`export KERENSCOPE_LLM_API_KEY="sk-..."`))
			return false
		}
		fmt.Printf("    %s no LLM key set — pinging the free hosted demo endpoint…\n", statusIcon(true))
		demo := &llm.OpenAICompat{
			BaseURL:       demoLLMBaseURL,
			APIKey:        "demo",
			Model:         "@cf/zai-org/glm-5.3-flash",
			FallbackModel: "@cf/zai-org/glm-5.3",
			Effort:        "low",
			Client:        &http.Client{Timeout: 30 * time.Second},
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
		defer cancel()
		start := time.Now()
		if _, derr := demo.Complete(ctx, llm.Request{Messages: []llm.Message{{Role: "user", Content: "Reply with one word: ok"}}, MaxTokens: 16}); derr != nil {
			fmt.Printf("    %s hosted demo unreachable: %s\n", red("✗"), derr)
			fmt.Printf("      %s set your own key via KERENSCOPE_LLM_API_KEY\n", dim("→"))
			return false
		}
		fmt.Printf("    %s hosted demo reachable %s\n", green("✓"), dim(fmt.Sprintf("(%s) — research works out of the box", time.Since(start).Round(time.Millisecond))))
		fmt.Printf("      %s bring your own key any time for full speed and unlimited runs\n", dim("→"))
		return true
	}
	model := os.Getenv("KERENSCOPE_LLM_MODEL")
	if model == "" {
		model = "gpt-4o-mini (default)"
	}
	base := os.Getenv("KERENSCOPE_LLM_BASE_URL")
	if base == "" {
		base = "https://api.openai.com/v1 (default)"
	}
	fmt.Printf("    %s LLM config found — pinging %s…\n", statusIcon(true), cyan(model))
	fmt.Printf("      %s\n", dim(base))
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()
	start := time.Now()
	resp, err := provider.Complete(ctx, llm.Request{
		Messages: []llm.Message{{Role: "user", Content: "Reply with one word: ok"}},
		MaxTokens: 16,
	})
	if err != nil {
		fmt.Printf("    %s LLM endpoint unreachable: %s\n", red("✗"), err)
		fmt.Printf("      %s check the key, base URL and model name\n", dim("→"))
		return false
	}
	replied := resp.Model
	if replied == "" {
		replied = model
	}
	fmt.Printf("    %s LLM reachable %s\n", green("✓"), dim(fmt.Sprintf("(%s replied in %s)", replied, time.Since(start).Round(time.Millisecond))))
	if fb := os.Getenv("KERENSCOPE_LLM_FALLBACK_MODEL"); fb != "" {
		fmt.Printf("      %s fallback model: %s\n", dim("→"), fb)
	}
	return true
}

func checkLocal() {
	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache", "kerenscope")
	entries := 0
	if files, err := os.ReadDir(cacheDir); err == nil {
		entries = len(files)
	}
	fmt.Printf("    %s disk cache: %s %s\n", green("✓"), cacheDir, dim(fmt.Sprintf("(%d entries)", entries)))
	summary := sectors.ReadLedger(sectors.DefaultLedgerPath())
	fmt.Printf("    %s credit ledger: %s %s\n", green("✓"), sectors.DefaultLedgerPath(), dim(fmt.Sprintf("(%d requests, %d credits spent)", summary.Requests, summary.Spent)))
	fmt.Printf("    %s terminal: %s\n", green("✓"), dim(fmt.Sprintf("colors %s, width %d", map[bool]string{true: "on", false: "off"}[colorOn], termWidth())))
}
