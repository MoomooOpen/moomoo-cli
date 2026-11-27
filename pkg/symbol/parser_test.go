package symbol

import (
	"strings"
	"testing"
)

func TestParseMarketCodeSymbol(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Symbol
	}{
		{
			name:  "hong kong pads numeric code",
			input: "HK.700",
			want:  &Symbol{Market: MarketHK, Code: "00700", Full: "HK.00700"},
		},
		{
			name:  "lowercase us normalizes",
			input: "us.aapl",
			want:  &Symbol{Market: MarketUS, Code: "AAPL", Full: "US.AAPL"},
		},
		{
			name:  "us class code keeps embedded dot",
			input: "US.BRK.B",
			want:  &Symbol{Market: MarketUS, Code: "BRK.B", Full: "US.BRK.B"},
		},
		{
			name:  "china market keeps six digit code",
			input: "SH.600519",
			want:  &Symbol{Market: MarketSH, Code: "600519", Full: "SH.600519"},
		},
		{
			name:  "crypto market normalizes",
			input: "cc.btc",
			want:  &Symbol{Market: MarketCC, Code: "BTC", Full: "CC.BTC"},
		},
		{
			name:  "japan market",
			input: "JP.7203",
			want:  &Symbol{Market: MarketJP, Code: "7203", Full: "JP.7203"},
		},
		{
			name:  "singapore market",
			input: "SG.1V2",
			want:  &Symbol{Market: MarketSG, Code: "1V2", Full: "SG.1V2"},
		},
		{
			name:  "malaysia market",
			input: "BMS.MYIPO007",
			want:  &Symbol{Market: MarketBMS, Code: "MYIPO007", Full: "BMS.MYIPO007"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Parse(test.input)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", test.input, err)
			}
			if *got != *test.want {
				t.Fatalf("Parse(%q) = %#v, want %#v", test.input, got, test.want)
			}
			if got.FutuCode() != test.want.Full {
				t.Fatalf("FutuCode() = %q, want %q", got.FutuCode(), test.want.Full)
			}
		})
	}
}

func TestParseRejectsNonMarketCodeFormats(t *testing.T) {
	tests := []string{"700.HK", "AAPL", "700", "HK.", ".AAPL", "BTC-USD"}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := Parse(input)
			if err == nil {
				t.Fatalf("Parse(%q) error = nil", input)
			}
			if !strings.Contains(err.Error(), "MARKET.CODE") {
				t.Fatalf("Parse(%q) error = %v, want MARKET.CODE hint", input, err)
			}
		})
	}
}

func TestParseAdditionalMarketPrefixes(t *testing.T) {
	markets := []string{MarketSG, MarketJP, MarketMY, MarketAU, MarketCA, MarketKR, MarketCN, MarketCC, MarketHKCC, MarketBMS}
	for _, market := range markets {
		t.Run(market, func(t *testing.T) {
			got, err := Parse(strings.ToLower(market) + ".example")
			if err != nil || got.Full != market+".EXAMPLE" {
				t.Fatalf("Parse(%s) = %#v, %v", market, got, err)
			}
		})
	}
}
