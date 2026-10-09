package cmd

import (
	"strings"
	"testing"
)

func TestParseWeb3Amount(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "integer", value: "2", want: "2"},
		{name: "decimal with whitespace", value: " 0.125 ", want: "0.125"},
		{name: "high precision", value: "0.123456789012345678", want: "0.123456789012345678"},
		{name: "zero", value: "0", wantErr: true},
		{name: "negative", value: "-1", wantErr: true},
		{name: "nan", value: "NaN", wantErr: true},
		{name: "positive infinity", value: "+Inf", wantErr: true},
		{name: "negative infinity", value: "-Inf", wantErr: true},
		{name: "empty", value: "", wantErr: true},
		{name: "not numeric", value: "one", wantErr: true},
		{name: "exponent notation", value: "1e-8", wantErr: true},
		{name: "missing integer part", value: ".5", wantErr: true},
		{name: "missing fraction", value: "1.", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseWeb3Amount(test.value)
			if test.wantErr {
				if err == nil {
					t.Fatalf("parseWeb3Amount(%q) error = nil", test.value)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseWeb3Amount(%q) error = %v", test.value, err)
			}
			if got != test.want {
				t.Fatalf("parseWeb3Amount(%q) = %v, want %v", test.value, got, test.want)
			}
		})
	}
}

func TestWeb3CommandsRejectInvalidAmountBeforeAPICall(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "swap",
			run: func() error {
				return runWeb3Swap(web3SwapCmd, []string{"USDC", "ETH", "NaN"})
			},
		},
		{
			name: "transfer",
			run: func() error {
				return runWeb3Transfer(web3TransferCmd, []string{"0xabc", "USDC", "-1"})
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); err == nil {
				t.Fatal("expected invalid amount error")
			}
		})
	}
}

func TestConfirmWeb3ActionWithReader(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "lowercase y", input: "y\n", want: true},
		{name: "uppercase yes", input: "YES\n", want: true},
		{name: "surrounding whitespace", input: "  yes  \n", want: true},
		{name: "no", input: "n\n", want: false},
		{name: "word beginning with y", input: "yikes\n", want: false},
		{name: "empty", input: "\n", want: false},
		{name: "incomplete input", input: "y", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := confirmWeb3ActionWithReader("Confirm?", strings.NewReader(test.input)); got != test.want {
				t.Fatalf("confirmation = %t, want %t", got, test.want)
			}
		})
	}
}
