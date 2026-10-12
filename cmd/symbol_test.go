package cmd

import "testing"

func TestNormalizeSecuritySymbolPadsHKCode(t *testing.T) {
	got, err := normalizeSecuritySymbol("HK.700")
	if err != nil {
		t.Fatalf("normalizeSecuritySymbol() error = %v", err)
	}
	if got != "HK.00700" {
		t.Fatalf("symbol = %q, want HK.00700", got)
	}
}

func TestNormalizeSecuritySymbolRejectsLegacyFormat(t *testing.T) {
	if _, err := normalizeSecuritySymbol("700.HK"); err == nil {
		t.Fatal("expected error for code.market format")
	}
}

func TestNormalizeOptionalSecuritySymbolAllowsEmpty(t *testing.T) {
	got, err := normalizeOptionalSecuritySymbol(" ")
	if err != nil {
		t.Fatalf("normalizeOptionalSecuritySymbol() error = %v", err)
	}
	if got != "" {
		t.Fatalf("symbol = %q, want empty", got)
	}
}

func TestNormalizeOptionalSecuritySymbolWithMarketUsesSymbolMarket(t *testing.T) {
	symbol, market, err := normalizeOptionalSecuritySymbolWithMarket("US.AAPL", "HK")
	if err != nil {
		t.Fatalf("normalizeOptionalSecuritySymbolWithMarket() error = %v", err)
	}
	if symbol != "US.AAPL" || market != "US" {
		t.Fatalf("symbol = %q, market = %q; want US.AAPL, US", symbol, market)
	}
}
