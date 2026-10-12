package symbol

import (
	"fmt"
	"strings"
	"unicode"
)

// Market identifier constants.
const (
	MarketHK  = "HK"
	MarketUS  = "US"
	MarketSH  = "SH"
	MarketSZ  = "SZ"
	MarketBJ  = "BJ"
	MarketBMD = "BMD"
)

// knownMarkets contains market prefixes recognised by the moomoo API.
var knownMarkets = map[string]bool{
	MarketHK:  true,
	MarketUS:  true,
	MarketSH:  true,
	MarketSZ:  true,
	MarketBJ:  true,
	MarketBMD: true,
}

// HK stock code width for zero-padding.
const hkCodeWidth = 5

// Symbol represents a parsed financial instrument identifier.
type Symbol struct {
	Code   string // e.g. "00700"
	Market string // e.g. "HK"
	Full   string // e.g. "HK.00700"
}

// Parse parses a symbol input string into a Symbol struct.
//
// Supported formats:
//   - "HK.700" -> pads to "HK.00700"
//   - "US.AAPL" -> as-is
func Parse(input string) (*Symbol, error) {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return nil, fmt.Errorf("empty symbol input")
	}

	code, market, err := splitSymbol(input)
	if err != nil {
		return nil, err
	}
	market = strings.ToUpper(market)
	code = normalizeCode(code, market)

	return buildSymbol(code, market), nil
}

// ParseMulti parses multiple symbol input strings.
func ParseMulti(inputs []string) ([]*Symbol, error) {
	if len(inputs) == 0 {
		return nil, fmt.Errorf("no symbol inputs provided")
	}

	symbols := make([]*Symbol, 0, len(inputs))
	for _, input := range inputs {
		sym, err := Parse(input)
		if err != nil {
			return nil, fmt.Errorf("parsing %q: %w", input, err)
		}
		symbols = append(symbols, sym)
	}
	return symbols, nil
}

// splitSymbol splits a moomoo-format symbol into (code, market).
func splitSymbol(input string) (string, string, error) {
	idx := strings.LastIndex(input, ".")
	if idx <= 0 || idx == len(input)-1 {
		return "", "", fmt.Errorf("symbol must use MARKET.CODE format, for example HK.00700 or US.AAPL")
	}
	left, right := input[:idx], input[idx+1:]
	market := strings.ToUpper(left)
	if !knownMarkets[market] {
		return "", "", fmt.Errorf("unsupported symbol market %q; use MARKET.CODE format", left)
	}
	return right, market, nil
}

// isAllDigits returns true if s consists entirely of ASCII digits.
func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// normalizeCode applies market-specific normalization.
// HK stock codes are zero-padded to 5 digits.
func normalizeCode(code, market string) string {
	if market == MarketHK && isAllDigits(code) {
		return padLeft(code, '0', hkCodeWidth)
	}
	return strings.ToUpper(code)
}

// padLeft pads s with the given character on the left until it reaches width.
func padLeft(s string, pad rune, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(string(pad), width-len(s)) + s
}

// buildSymbol constructs a Symbol from code and market.
func buildSymbol(code, market string) *Symbol {
	return &Symbol{
		Code:   code,
		Market: market,
		Full:   market + "." + code,
	}
}

// FutuCode returns the symbol in moomoo API format: "MARKET.CODE".
func (s *Symbol) FutuCode() string {
	return s.Full
}

// APICode returns the symbol in moomoo API format: "MARKET.CODE".
func (s *Symbol) APICode() string {
	return s.FutuCode()
}
