package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestSimCommandRegistration(t *testing.T) {
	for _, args := range [][]string{
		{"sim", "account", "list"},
		{"sim", "account", "funds"},
		{"sim", "position"},
		{"sim", "order", "buy"},
		{"sim", "order", "sell"},
		{"sim", "order", "place"},
		{"sim", "order", "list"},
		{"sim", "order", "history"},
		{"sim", "order", "modify"},
		{"sim", "order", "cancel"},
		{"sim", "order", "max-qty"},
	} {
		if _, _, err := rootCmd.Find(args); err != nil {
			t.Fatalf("command %v not registered: %v", args, err)
		}
	}
}

func TestFlatSimTradeCommandsAreNotRegistered(t *testing.T) {
	for _, name := range []string{
		"sim-trade-account-list",
		"sim-trade-cash-info",
		"sim-trade-order-list",
		"sim-trade-history-order-list",
		"sim-trade-max-buy-sell",
		"sim-trade-input-order",
		"sim-trade-modify-order",
		"sim-trade-cancel-order",
		"sim-trade-position-list",
	} {
		if _, _, err := rootCmd.Find([]string{name}); err == nil {
			t.Fatalf("flat simulated-trading command %q is still registered", name)
		}
	}
}

func TestSimMarketValue(t *testing.T) {
	tests := []struct {
		value string
		want  int
	}{
		{value: "", want: simMarketHongKong},
		{value: "HK", want: simMarketHongKong},
		{value: "us", want: simMarketUnitedStates},
		{value: "US_OPTION", want: simMarketUSOption},
		{value: "hkcc", want: simMarketHongKongConnect},
	}

	for _, test := range tests {
		got, err := simMarketValue(test.value)
		if err != nil {
			t.Fatalf("simMarketValue(%q) error = %v", test.value, err)
		}
		if got != test.want {
			t.Fatalf("simMarketValue(%q) = %d, want %d", test.value, got, test.want)
		}
	}
}

func TestSimOrderTypeAndSideValues(t *testing.T) {
	orderType, err := simOrderTypeValue("limit")
	if err != nil {
		t.Fatalf("simOrderTypeValue() error = %v", err)
	}
	if orderType != simOrderTypeLimit {
		t.Fatalf("order type = %d, want %d", orderType, simOrderTypeLimit)
	}

	side, err := simOrderSideValue("sell")
	if err != nil {
		t.Fatalf("simOrderSideValue() error = %v", err)
	}
	if side != simOrderSideSellValue {
		t.Fatalf("order side = %d, want %d", side, simOrderSideSellValue)
	}
}

func TestBuildSimPlaceOrderBody(t *testing.T) {
	oldFlags := simFlags
	defer func() { simFlags = oldFlags }()
	simFlags.market = "US"
	simFlags.typ = "limit"
	simFlags.price = "180.5"
	simFlags.text = "paper test"
	simFlags.session = "RTH"

	body, err := buildSimPlaceOrderBody("US.AAPL", "2", orderSideBuy)
	if err != nil {
		t.Fatalf("buildSimPlaceOrderBody() error = %v", err)
	}
	if body[simFieldMarket] != simMarketUnitedStates || body[simFieldSymbol] != "US.AAPL" {
		t.Fatalf("market/symbol body = %#v", body)
	}
	if body[simFieldOrderType] != simOrderTypeLimit || body[simFieldOrderSide] != simOrderSideBuyValue {
		t.Fatalf("order enum body = %#v", body)
	}
	if body[simFieldQuantity] != "2" || body[simFieldPrice] != "180.5" || body[simFieldText] != "paper test" {
		t.Fatalf("order value body = %#v", body)
	}
	if body[simFieldTradeTimeType] != simTradeTimeTypeRegular {
		t.Fatalf("session body = %#v", body)
	}
}

func TestBuildSimPlaceOrderBodyRejectsInvalidQuantity(t *testing.T) {
	oldFlags := simFlags
	defer func() { simFlags = oldFlags }()
	simFlags.market = defaultSimMarket
	simFlags.typ = defaultSimOrderType

	if _, err := buildSimPlaceOrderBody("US.AAPL", "0", orderSideBuy); err == nil {
		t.Fatal("expected invalid quantity error")
	}
}

func TestSimModifyFlagsAreRequired(t *testing.T) {
	for _, name := range []string{"qty", "price"} {
		flag := simOrderModifyCmd.Flags().Lookup(name)
		if flag == nil {
			t.Fatalf("flag %s is not registered", name)
		}
		if len(flag.Annotations[cobra.BashCompOneRequiredFlag]) == 0 {
			t.Fatalf("flag %s is not marked required", name)
		}
	}
}

func TestApplySimModifyBodyRequiresQuantityAndPrice(t *testing.T) {
	tests := []struct {
		name     string
		quantity string
		price    string
		wantErr  string
		wantBody map[string]any
	}{
		{
			name:    "missing quantity",
			price:   "181",
			wantErr: simModifyQtyRequiredError,
		},
		{
			name:     "missing price",
			quantity: "3",
			wantErr:  simModifyPriceRequiredError,
		},
		{
			name:     "complete modify body",
			quantity: "3",
			price:    "181",
			wantBody: map[string]any{
				simFieldNewQuantity: "3",
				simFieldNewPrice:    "181",
			},
		},
	}

	oldFlags := simFlags
	defer func() { simFlags = oldFlags }()
	for _, test := range tests {
		simFlags.quantity = test.quantity
		simFlags.price = test.price
		body := map[string]any{}
		err := applySimModifyBody(body)
		if test.wantErr != "" {
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("%s: error = %v, want %q", test.name, err, test.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: applySimModifyBody() error = %v", test.name, err)
		}
		for key, want := range test.wantBody {
			if got := body[key]; got != want {
				t.Fatalf("%s: body[%s] = %#v, want %#v", test.name, key, got, want)
			}
		}
	}
}

func TestParsePositiveSimQuantityRejectsNaN(t *testing.T) {
	if _, err := parsePositiveSimQuantity("NaN"); err == nil {
		t.Fatal("expected invalid NaN quantity error")
	}
}
