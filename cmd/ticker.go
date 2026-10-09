package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/MoomooOpen/moomoo-cli/internal/output"
	"github.com/MoomooOpen/moomoo-cli/internal/service"
	"github.com/MoomooOpen/moomoo-cli/pkg/symbol"
)

// Default value for the ticker command count flag.
const defaultTickerCount = 50

// Table header labels for the ticker command.
const (
	tickerHeaderTime      = "Time"
	tickerHeaderPrice     = "Price"
	tickerHeaderVolume    = "Volume"
	tickerHeaderDirection = "Direction"
)

// tickerFlags holds the flag values for the ticker command.
var tickerFlags struct {
	count int
}

var tickerCmd = &cobra.Command{
	Use:   "ticker <symbol>",
	Short: "Get tick-by-tick trades",
	Long:  "Get tick-by-tick trades for a security, including trade price, volume, and direction.",
	Args:  cobra.ExactArgs(1),
	RunE:  runTicker,
}

func init() {
	tickerCmd.Flags().IntVarP(&tickerFlags.count, "count", "c", defaultTickerCount,
		"Number of ticks to return (default: 50)")
	rootCmd.AddCommand(tickerCmd)
}

// runTicker is the top-level handler for the ticker command.
func runTicker(cmd *cobra.Command, args []string) error {
	sym, err := symbol.Parse(args[0])
	if err != nil {
		return fmt.Errorf("parsing symbol: %w", err)
	}

	ctx := cmdContext(cmd)
	ticks, err := quoteSvc.GetTicker(ctx, sym.Full, tickerFlags.count)
	if err != nil {
		return fmt.Errorf("fetching ticker: %w", err)
	}

	printResult(ticks, func() {
		renderTickerTable(ticks)
	})
	return nil
}

// renderTickerTable writes tick-by-tick trade data as a formatted table
// to stdout.
func renderTickerTable(ticks []service.Ticker) {
	tw := output.NewTableWriter(os.Stdout,
		tickerHeaderTime,
		tickerHeaderPrice,
		tickerHeaderVolume,
		tickerHeaderDirection,
	)

	for _, t := range ticks {
		tw.AddRow(
			formatTimestamp(t.Time),
			fmt.Sprintf("%.3f", t.Price),
			fmt.Sprintf("%d", t.Volume),
			t.Direction,
		)
	}
	tw.Render()
}
