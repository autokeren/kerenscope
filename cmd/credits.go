package cmd

import (
	"fmt"

	"github.com/autokeren/kerenscope/internal/sectors"
	"github.com/spf13/cobra"
)

var creditsCmd = &cobra.Command{
	Use:   "credits",
	Short: "Show the local Sectors API credit ledger (estimated spend)",
	RunE: func(cmd *cobra.Command, args []string) error {
		summary := sectors.ReadLedger(sectors.DefaultLedgerPath())
		fmt.Println("KerenScope — Sectors API credit ledger (local estimate)")
		fmt.Println()
		fmt.Printf("  API requests: %d\n", summary.Requests)
		fmt.Printf("  Credits spent: %d\n", summary.Spent)
		fmt.Printf("  Cache hits (free): %d\n", summary.CacheHits)
		if len(summary.ByEndpoint) > 0 {
			fmt.Println("\n  By endpoint:")
			for _, ep := range sortedKeys(summary.ByEndpoint) {
				fmt.Printf("    %3d  %s\n", summary.ByEndpoint[ep], ep)
			}
		}
		fmt.Println("\n  Team grant: 1,000 hackathon credits (expires Sep 30, 2026)")
		fmt.Printf("  Ledger file: %s\n", sectors.DefaultLedgerPath())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(creditsCmd)
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
