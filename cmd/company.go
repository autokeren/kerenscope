package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/autokeren/kerenscope/internal/sectors"
	"github.com/spf13/cobra"
)

var companyCmd = &cobra.Command{
	Use:   "company <symbol>",
	Short: "Fetch a full company report from Sectors",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
		defer cancel()
		report, err := client.CompanyReport(ctx, args[0])
		if err != nil {
			return err
		}
		if asJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(report)
		}
		renderCompany(report)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(companyCmd)
}

func renderCompany(r *sectors.CompanyReport) {
	var b strings.Builder
	b.WriteString(title("KerenScope — Company Report"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s — %s\n", r.Symbol, r.CompanyName))
	ov := r.Overview
	b.WriteString(fmt.Sprintf("%s · %s · %s\n", ov.Sector, ov.SubSector, ov.ListingBoard))
	b.WriteString(fmt.Sprintf("Market cap: %s (#%d in IDX)\n", formatIDR(ov.MarketCap), ov.MarketCapRank))
	b.WriteString("\n")
	val := r.Valuation
	b.WriteString(subtitle("Price & valuation"))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Last close: %s (%s)\n", formatIDR(int64(val.LastClosePrice)), val.LatestCloseDate))
	if v := r.LatestValuation(); v != nil {
		b.WriteString(fmt.Sprintf("P/B %.2f (peer %.2f) · P/E %.2f (peer %.2f) · PEG %.2f\n", v.PB, v.PBPeer, v.PE, v.PEPeer, v.PEG))
	}
	if val.ForwardPE > 0 {
		b.WriteString(fmt.Sprintf("Forward P/E: %.2f\n", val.ForwardPE))
	}
	b.WriteString("\n")
	b.WriteString(subtitle("Analyst estimates"))
	b.WriteString("\n")
	future := r.Future
	for _, f := range future.CompanyGrowthForecasts {
		b.WriteString(fmt.Sprintf("FY%d→FY%d: revenue %s, EPS %s\n", f.BaseYear, f.EstimateYear, pct(f.RevenueGrowth), pct(f.EPSGrowth)))
	}
	rating := future.AnalystRating
	if rating.NAnalyst > 0 {
		b.WriteString(fmt.Sprintf("Analyst ratings: %d buy / %d hold / %d sell (%d analysts)\n", rating.Buy, rating.Hold, rating.Sell, rating.NAnalyst))
	}
	b.WriteString("\n")
	b.WriteString(disclaimer())
	b.WriteString("\n")
	fmt.Print(b.String())
}

func title(s string) string {
	return "╭─ " + s + " " + strings.Repeat("─", max(0, 52-len(s))) + "╮"
}

func subtitle(s string) string {
	return "── " + s + " ──"
}

func disclaimer() string {
	return "Information & analysis only — not investment advice. Data: Sectors API."
}

func pct(v float64) string {
	return fmt.Sprintf("%+.1f%%", v*100)
}

func formatIDR(v int64) string {
	abs := v
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs >= 1_000_000_000_000_000:
		return fmt.Sprintf("Rp%.1fqt", float64(v)/1e15)
	case abs >= 1_000_000_000_000:
		return fmt.Sprintf("Rp%.1fT", float64(v)/1e12)
	case abs >= 1_000_000_000:
		return fmt.Sprintf("Rp%.1fB", float64(v)/1e9)
	case abs >= 1_000_000:
		return fmt.Sprintf("Rp%.1fM", float64(v)/1e6)
	default:
		return fmt.Sprintf("Rp%d", v)
	}
}
