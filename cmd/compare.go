package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var compareCmd = &cobra.Command{
	Use:   "compare <symbol> <symbol> [<symbol>...]",
	Short: "Compare companies side by side (fundamentals, valuation, momentum)",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		for i := range args {
			args[i] = strings.ToUpper(strings.TrimSpace(args[i]))
		}
		question := fmt.Sprintf(
			"Bandingkan %s secara menyeluruh: fundamental (profitabilitas, kualitas aset), valuasi, momentum, dan sinyal smart money (foreign flow, broker, insider). "+
				"Sertakan skor komposit dari engine dan jelaskan kelebihan serta risiko masing-masing.",
			strings.Join(args, ", "))
		return runResearchQuestion(cmd.Context(), question)
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)
}
