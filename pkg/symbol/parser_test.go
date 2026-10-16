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
