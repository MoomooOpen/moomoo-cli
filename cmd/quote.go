package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/MoomooOpen/moomoo-cli/internal/output"
	"github.com/MoomooOpen/moomoo-cli/internal/service"
	"github.com/MoomooOpen/moomoo-cli/pkg/symbol"
)

// Table header labels for the quote command.
const (
	quoteHeaderCode      = "Code"
	quoteHeaderName      = "Name"
	quoteHeaderLast      = "Last"
	quoteHeaderChange    = "Change"
	quoteHeaderChangePct = "Change%"
	quoteHeaderVolume    = "Volume"
	quoteHeaderTurnover  = "Turnover"
	quoteHeaderHigh      = "High"
	quoteHeaderLow       = "Low"
)

var quoteCmd = &cobra.Command{
	Use:   "quote <symbol...>",
	Short: "Get real-time quote snapshots",
	Long: `Get real-time quote snapshots for one or more securities, including last price, change, change percentage, volume, and turnover.

Supported symbol formats:
  HK.09988     moomoo format (market.code)
  9988         numeric-only symbols default to Hong Kong
  700.HK       code.market
  AAPL         alphabetic-only symbols default to US
  US.AAPL      moomoo format for US stocks`,
	Args: cobra.MinimumNArgs(1),
	RunE: runQuote,
}

func init() {
	rootCmd.AddCommand(quoteCmd)
}

func runQuote(cmd *cobra.Command, args []string) error {
	codes, err := parseAPICodes(args)
	if err != nil {
		return err
	}

	ctx := cmdContext(cmd)
	quotes, err := quoteSvc.GetStockQuote(ctx, codes)
	if err != nil {
		return fmt.Errorf("fetching quotes: %w", err)
	}

	printResult(quotes, func() {
		renderQuoteTable(quotes)
	})
	return nil
}

// parseAPICodes parses args into moomoo API format codes (e.g. "HK.09988").
func parseAPICodes(args []string) ([]string, error) {
	parsed, err := symbol.ParseMulti(args)
	if err != nil {
		return nil, fmt.Errorf("parsing symbols: %w", err)
	}

	codes := make([]string, 0, len(parsed))
	for _, s := range parsed {
		codes = append(codes, s.APICode())
	}
	return codes, nil
}

func renderQuoteTable(quotes []service.StockQuote) {
	tw := output.NewTableWriter(os.Stdout,
		quoteHeaderCode,
		quoteHeaderName,
		quoteHeaderLast,
		quoteHeaderChange,
		quoteHeaderChangePct,
		quoteHeaderHigh,
		quoteHeaderLow,
		quoteHeaderVolume,
		quoteHeaderTurnover,
	)

	for _, q := range quotes {
		change := q.LastPrice - q.PrevClosePrice
		var changePct float64
		if q.PrevClosePrice != 0 {
			changePct = change / q.PrevClosePrice * 100
		}

		tw.AddRow(
			q.Code,
			q.SCName,
			fmt.Sprintf("%.3f", q.LastPrice),
			formatChange(change),
			formatChangePct(changePct),
			fmt.Sprintf("%.3f", q.HighPrice),
			fmt.Sprintf("%.3f", q.LowPrice),
			fmt.Sprintf("%d", q.Volume),
			formatTurnover(q.Turnover),
		)
	}
	tw.Render()
}

func formatChange(v float64) string {
	if v > 0 {
		return fmt.Sprintf("+%.3f", v)
	}
	return fmt.Sprintf("%.3f", v)
}

func formatChangePct(v float64) string {
	if v > 0 {
		return fmt.Sprintf("+%.2f%%", v)
	}
	return fmt.Sprintf("%.2f%%", v)
}

func formatTurnover(v float64) string {
	const (
		billion = 1e8
		tenK    = 1e4
	)
	switch {
	case v >= billion:
		return fmt.Sprintf("%.2fB", v/billion)
	case v >= tenK:
		return fmt.Sprintf("%.2f x10K", v/tenK)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}
