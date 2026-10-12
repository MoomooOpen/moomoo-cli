package cmd

import (
	"fmt"
	"strings"

	"github.com/MoomooOpen/moomoo-cli/pkg/symbol"
)

const securitySymbolUsage = "security symbol in MARKET.CODE format, for example HK.00700 or US.AAPL"

func normalizeSecuritySymbol(rawSymbol string) (string, error) {
	sym, err := symbol.Parse(rawSymbol)
	if err != nil {
		return "", fmt.Errorf("parsing symbol: %w", err)
	}
	return sym.FutuCode(), nil
}

func normalizeOptionalSecuritySymbol(rawSymbol string) (string, error) {
	if len(strings.TrimSpace(rawSymbol)) == 0 {
		return "", nil
	}
	return normalizeSecuritySymbol(rawSymbol)
}

func normalizeOptionalSecuritySymbolWithMarket(rawSymbol, fallbackMarket string) (string, string, error) {
	if len(strings.TrimSpace(rawSymbol)) == 0 {
		return "", strings.TrimSpace(fallbackMarket), nil
	}
	sym, err := symbol.Parse(rawSymbol)
	if err != nil {
		return "", "", fmt.Errorf("parsing symbol: %w", err)
	}
	return sym.FutuCode(), sym.Market, nil
}
