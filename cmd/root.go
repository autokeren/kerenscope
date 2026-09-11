package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/autokeren/kerenscope/internal/sectors"
	"github.com/spf13/cobra"
)

var version = "0.4.2-dev"

var (
	noCache   bool
	asJSON    bool
	revealSlow bool
	noBrowser bool
)

var rootCmd = &cobra.Command{
	Use:   "keren [question]",
	Short: "KerenScope — autonomous financial research agent for the Indonesian market, powered by Sectors",
	Long: `KerenScope is an autonomous financial research agent.

Give it a research question about the Indonesian stock market and it
investigates on its own: planning, querying Sectors data, cross-checking
evidence, and writing an evidence-backed report.

Information and analysis only — not investment advice.`,
	Version: version,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			question := strings.Join(args, " ")
			if err := runResearchQuestion(cmd.Context(), question); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			return
		}
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noCache, "no-cache", false, "bypass the disk cache (debugging; spends credits)")
	rootCmd.PersistentFlags().BoolVar(&revealSlow, "reveal-slow", false, "dramatic line-by-line reveal of the final report (for recording)")
	rootCmd.PersistentFlags().BoolVar(&noBrowser, "no-browser", false, "do not open the HTML report in a browser automatically")
	rootCmd.PersistentFlags().BoolVar(&asJSON, "json", false, "print raw JSON instead of a rendered view")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func newClient() (*sectors.Client, error) {
	if err := ensureSectorsKey(); err != nil {
		return nil, err
	}
	return sectors.New(sectors.Options{NoCache: noCache})
}
