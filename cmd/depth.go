package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/MoomooOpen/moomoo-cli/internal/output"
	"github.com/MoomooOpen/moomoo-cli/internal/service"
	"github.com/MoomooOpen/moomoo-cli/pkg/symbol"
)

// Table header labels for the depth command.
const (
	depthHeaderAskPrice = "AskPrice"
	depthHeaderAskVol   = "AskVol"
	depthHeaderBidPrice = "BidPrice"
	depthHeaderBidVol   = "BidVol"
)

// Placeholder for missing depth levels when asks and bids differ in length.
const depthEmptyPlaceholder = "-"

var depthCmd = &cobra.Command{
	Use:   "depth <symbol>",
	Short: "Get order book depth",
	Long:  "Get order book depth for a security, showing ask and bid levels side by side.",
	Args:  cobra.ExactArgs(1),
	RunE:  runDepth,
}

func init() {
	rootCmd.AddCommand(depthCmd)
}

// runDepth is the top-level handler for the depth command.
func runDepth(cmd *cobra.Command, args []string) error {
	sym, err := symbol.Parse(args[0])
	if err != nil {
		return fmt.Errorf("parsing symbol: %w", err)
	}

	ctx := cmdContext(cmd)
	depth, err := quoteSvc.GetDepth(ctx, sym.Full)
	if err != nil {
		return fmt.Errorf("fetching depth: %w", err)
	}

	printResult(depth, func() {
		renderDepthTable(depth)
	})
	return nil
}

// renderDepthTable writes order book data as a formatted table to stdout,
// aligning asks and bids side-by-side.
func renderDepthTable(depth *service.Depth) {
	tw := output.NewTableWriter(os.Stdout,
		depthHeaderAskPrice,
		depthHeaderAskVol,
		depthHeaderBidPrice,
		depthHeaderBidVol,
	)

	rowCount := max(len(depth.Asks), len(depth.Bids))
	for i := 0; i < rowCount; i++ {
		askPrice, askVol := formatDepthEntry(depth.Asks, i)
		bidPrice, bidVol := formatDepthEntry(depth.Bids, i)
		tw.AddRow(askPrice, askVol, bidPrice, bidVol)
	}
	tw.Render()
}

// formatDepthEntry returns formatted price and volume strings for the entry
// at index i, or placeholder strings if the index is out of range.
func formatDepthEntry(entries []service.DepthEntry, i int) (string, string) {
	if i >= len(entries) {
		return depthEmptyPlaceholder, depthEmptyPlaceholder
	}
	return fmt.Sprintf("%.3f", entries[i].Price), fmt.Sprintf("%d", entries[i].Volume)
}
