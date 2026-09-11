package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/autokeren/kerenscope/internal/config"
	"github.com/autokeren/kerenscope/internal/llm"
	"github.com/autokeren/kerenscope/internal/sectors"
	"golang.org/x/term"
)

const demoLLMBaseURL = "https://kerenscope-llm-demo.pyscalp.workers.dev/v1"

var promptReader *bufio.Reader

func reader() *bufio.Reader {
	if promptReader == nil {
		promptReader = bufio.NewReader(os.Stdin)
	}
	return promptReader
}

func ensureSectorsKey() error {
	if os.Getenv("SECTORS_API_KEY") != "" {
		return nil
	}
	if v, ok := config.Get("SECTORS_API_KEY"); ok {
		os.Setenv("SECTORS_API_KEY", v)
		return nil
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("SECTORS_API_KEY is not set — create one at sectors.app/api (API Key Management), then export it or run keren interactively once")
	}
	fmt.Println()
	fmt.Printf("  %s SECTORS_API_KEY is not set.\n", yellow("⚠"))
	fmt.Printf("    %s create one at sectors.app/api → API Key Management\n", dim("→"))
	for attempt := 0; attempt < 3; attempt++ {
		fmt.Printf("\n  %s ", bold("Paste your Sectors API key (input hidden):"))
		keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return err
		}
		key := strings.TrimSpace(string(keyBytes))
		if key == "" {
			return errors.New("no API key provided — aborting")
		}
		fmt.Printf("  %s validating against the Sectors API…\n", dim("→"))
		client, err := sectors.New(sectors.Options{APIKey: key})
		if err != nil {
			return err
		}
		if _, err := client.Subsectors(context.Background()); err != nil {
			fmt.Printf("  %s key rejected: %s\n", red("✗"), err)
			continue
		}
		fmt.Printf("  %s key valid!\n", green("✓"))
		if saveYes() {
			if err := config.Save(map[string]string{"SECTORS_API_KEY": key}); err != nil {
				fmt.Printf("  %s could not save config (%v) — continuing for this session\n", yellow("⚠"), err)
			} else {
				fmt.Printf("  %s saved to %s — future runs will not ask\n", green("✓"), dim(config.Path()))
			}
		}
		os.Setenv("SECTORS_API_KEY", key)
		return nil
	}
	return errors.New("key validation failed 3 times — aborting")
}

func useDemoLLM() {
	os.Setenv("KERENSCOPE_LLM_API_KEY", "demo")
	os.Setenv("KERENSCOPE_LLM_BASE_URL", demoLLMBaseURL)
	os.Setenv("KERENSCOPE_LLM_MODEL", "@cf/zai-org/glm-5.3-flash")
	os.Setenv("KERENSCOPE_LLM_FALLBACK_MODEL", "@cf/zai-org/glm-5.3")
	os.Setenv("KERENSCOPE_LLM_REASONING", "low")
}

func demoDisabled() bool {
	return os.Getenv("KERENSCOPE_LLM_DEMO") == "0"
}

func ensureLLMConfig() error {
	_, err := llm.ConfigFromEnv()
	if err == nil || !errors.Is(err, llm.ErrNoAPIKey) {
		return err
	}
	if demoDisabled() {
		if !term.IsTerminal(int(os.Stdin.Fd())) {
			return llm.ErrNoAPIKey
		}
		fmt.Println()
		fmt.Printf("  %s No LLM configured and the hosted demo is disabled (KERENSCOPE_LLM_DEMO=0).\n", yellow("⚠"))
		fmt.Printf("    %s research needs your own OpenAI-compatible key\n", dim("→"))
		return llm.ErrNoAPIKey
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		useDemoLLM()
		return nil
	}
	fmt.Println()
	fmt.Printf("  %s No LLM configured.\n", yellow("⚠"))
	fmt.Printf("    %s Press Enter to use the free hosted demo (GLM, rate-limited),\n", dim("→"))
	fmt.Printf("    %s or paste your own OpenAI-compatible API key.\n", dim("→"))
	fmt.Printf("\n  %s ", bold("Enter = hosted demo, or paste a key:"))
	keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return err
	}
	key := strings.TrimSpace(string(keyBytes))
	if key == "" {
		useDemoLLM()
		fmt.Printf("  %s using the hosted demo endpoint — no configuration needed\n", green("✓"))
		if saveYes() {
			if err := config.Save(map[string]string{
				"KERENSCOPE_LLM_API_KEY":  "demo",
				"KERENSCOPE_LLM_BASE_URL": demoLLMBaseURL,
				"KERENSCOPE_LLM_MODEL":    "@cf/zai-org/glm-5.3-flash",
			}); err != nil {
				fmt.Printf("  %s could not save config (%v) — continuing for this session\n", yellow("⚠"), err)
			} else {
				fmt.Printf("  %s saved to %s — future runs will not ask\n", green("✓"), dim(config.Path()))
			}
		}
		return nil
	}
	fmt.Println()
	fmt.Printf("  %s Autonomous research needs an LLM (any OpenAI-compatible provider).\n", yellow("⚠"))
	fmt.Printf("    %s OpenAI, OpenRouter, GLM, Ollama local — anything with a chat-completions endpoint\n", dim("→"))
	for attempt := 0; attempt < 3; attempt++ {
		fmt.Printf("\n  %s ", bold("Base URL [https://api.openai.com/v1]:"))
		base := strings.TrimSpace(readLine())
		if base == "" {
			base = "https://api.openai.com/v1"
		}
		fmt.Printf("  %s ", bold("Model [gpt-4o-mini]:"))
		model := strings.TrimSpace(readLine())
		if model == "" {
			model = "gpt-4o-mini"
		}
		fmt.Printf("  %s pinging %s…\n", dim("→"), model)
		os.Setenv("KERENSCOPE_LLM_API_KEY", key)
		os.Setenv("KERENSCOPE_LLM_BASE_URL", base)
		os.Setenv("KERENSCOPE_LLM_MODEL", model)
		provider, err := llm.ConfigFromEnv()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if _, perr := provider.Complete(ctx, llm.Request{
			Messages:  []llm.Message{{Role: "user", Content: "Reply with one word: ok"}},
			MaxTokens: 16,
		}); perr != nil {
			cancel()
			fmt.Printf("  %s endpoint rejected: %s\n", red("✗"), perr)
			continue
		}
		cancel()
		fmt.Printf("  %s LLM reachable!\n", green("✓"))
		if saveYes() {
			values := map[string]string{
				"KERENSCOPE_LLM_API_KEY":  key,
				"KERENSCOPE_LLM_BASE_URL": base,
				"KERENSCOPE_LLM_MODEL":    model,
			}
			if err := config.Save(values); err != nil {
				fmt.Printf("  %s could not save config (%v) — continuing for this session\n", yellow("⚠"), err)
			} else {
				fmt.Printf("  %s saved to %s — future runs will not ask\n", green("✓"), dim(config.Path()))
			}
		}
		return nil
	}
	return errors.New("LLM configuration failed 3 times — aborting")
}

func readLine() string {
	line, err := reader().ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

func saveYes() bool {
	fmt.Printf("  %s ", bold("Save permanently for future runs? [Y/n]:"))
	line := strings.ToLower(readLine())
	return line == "" || line == "y" || line == "yes"
}
