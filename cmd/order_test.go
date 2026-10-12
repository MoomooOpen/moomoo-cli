package cmd

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/MoomooOpen/moomoo-cli/internal/config"
	"github.com/MoomooOpen/moomoo-cli/internal/service"
)

func TestBuildPlaceOrderRequestDefaultsToMarketOrder(t *testing.T) {
	oldFlags := orderFlags
	oldCfg := cfg
	defer func() {
		orderFlags = oldFlags
		cfg = oldCfg
	}()
	cfg = &config.Config{DefaultAcct: "account-1"}
	orderFlags = struct {
		price        string
		auxPrice     string
		qty          int64
		typ          string
		account      string
		exchange     string
		orderID      string
		market       string
		pageSize     int
		start        int64
		end          int64
		symbol       string
		side         string
		timeInForce  string
		lotType      string
		remark       string
		session      string
		orderClass   string
		multiLegInfo string
	}{typ: defaultOrderType, timeInForce: defaultTimeInForce}

	req, err := buildPlaceOrderRequest("US.AAPL", "10", orderSideBuy)
	if err != nil {
		t.Fatalf("buildPlaceOrderRequest() error = %v", err)
	}
	if req.OrderType != orderTypeMarket || req.Price != "" || req.TimeInForce != defaultTimeInForce {
		t.Fatalf("default request = %#v", req)
	}
	if req.Account != "account-1" || req.Code != "US.AAPL" || req.Side != orderSideBuy || req.Qty != 10 {
		t.Fatalf("request identity = %#v", req)
	}
}

func TestBuildPlaceOrderRequestNormalizesHKSymbol(t *testing.T) {
	oldFlags := orderFlags
	oldCfg := cfg
	defer func() {
		orderFlags = oldFlags
		cfg = oldCfg
	}()
	cfg = &config.Config{DefaultAcct: "account-1"}
	orderFlags = struct {
		price        string
		auxPrice     string
		qty          int64
		typ          string
		account      string
		exchange     string
		orderID      string
		market       string
		pageSize     int
		start        int64
		end          int64
		symbol       string
		side         string
		timeInForce  string
		lotType      string
		remark       string
		session      string
		orderClass   string
		multiLegInfo string
	}{typ: defaultOrderType, timeInForce: defaultTimeInForce}

	req, err := buildPlaceOrderRequest("HK.700", "10", orderSideBuy)
	if err != nil {
		t.Fatalf("buildPlaceOrderRequest() error = %v", err)
	}
	if req.Code != "HK.00700" {
		t.Fatalf("code = %q, want HK.00700", req.Code)
	}
}

func TestBuildHistoryOrdersRequestNormalizesSymbolFilter(t *testing.T) {
	oldFlags := orderFlags
	defer func() { orderFlags = oldFlags }()
	orderFlags.account = "account-1"
	orderFlags.market = "US"
	orderFlags.symbol = "HK.700"

	req, err := buildHistoryOrdersRequest()
	if err != nil {
		t.Fatalf("buildHistoryOrdersRequest() error = %v", err)
	}
	if req.Symbol != "HK.00700" || req.Market != "HK" {
		t.Fatalf("symbol = %q, market = %q; want HK.00700, HK", req.Symbol, req.Market)
	}
}

func TestBuildPlaceOrderRequestSupportsAdvancedFields(t *testing.T) {
	oldFlags := orderFlags
	defer func() { orderFlags = oldFlags }()
	orderFlags.account = "account-1"
	orderFlags.typ = "stop_limit"
	orderFlags.timeInForce = "gtc"
	orderFlags.lotType = "odd"
	orderFlags.remark = "hedge"
	orderFlags.session = "ALL_DAY"
	orderFlags.price = "150.123456789012345678"
	orderFlags.auxPrice = "149.100000000000000001"
	orderFlags.orderClass = orderClassMultiLeg
	orderFlags.multiLegInfo = `{"legs":[{"code":"US.AAPL"}]}`

	req, err := buildPlaceOrderRequest("", "2", "sell_short")
	if err != nil {
		t.Fatalf("buildPlaceOrderRequest() error = %v", err)
	}
	if req.Side != orderSideSellShort || req.OrderType != "STOP_LIMIT" || req.LotType != "ODD" {
		t.Fatalf("normalized request = %#v", req)
	}
	if req.TimeInForce != "GTC" || req.Remark != "hedge" || req.Session != "ALL_DAY" {
		t.Fatalf("optional request fields = %#v", req)
	}
	if req.Price != orderFlags.price || req.AuxPrice != orderFlags.auxPrice {
		t.Fatalf("precise prices = %#v", req)
	}
	if !json.Valid(req.MultiLegInfo) || req.Code != "" {
		t.Fatalf("multi-leg request = %#v", req)
	}
}

func TestConfirmActionAcceptsExplicitYes(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin; _ = reader.Close() }()
	os.Stdin = reader
	if _, err := writer.WriteString("yes\n"); err != nil {
		t.Fatalf("write confirmation: %v", err)
	}
	_ = writer.Close()
	if !confirmAction("Confirm test?") {
		t.Fatal("confirmAction() rejected explicit yes")
	}
}

func TestBuildPlaceOrderRequestRejectsInvalidMultiLegJSON(t *testing.T) {
	oldFlags := orderFlags
	defer func() { orderFlags = oldFlags }()
	orderFlags.account = "account-1"
	orderFlags.typ = defaultOrderType
	orderFlags.timeInForce = defaultTimeInForce
	orderFlags.orderClass = orderClassMultiLeg
	orderFlags.multiLegInfo = `{"legs":`

	_, err := buildPlaceOrderRequest("", "2", orderSideBuyBack)
	if err == nil || !strings.Contains(err.Error(), "JSON object") {
		t.Fatalf("error = %v, want invalid JSON error", err)
	}
}

func TestSingleOrderColumnsOmitsEmptyValues(t *testing.T) {
	order := &service.Order{
		OrderID: "FH1D27CBC2D97F2000",
		Type:    "MODIFIED",
		Price:   301,
		Qty:     1,
	}

	headers, row := singleOrderColumns(order)

	wantHeaders := []string{orderHeaderID, orderHeaderType, orderHeaderPrice, orderHeaderQty}
	wantRow := []string{"FH1D27CBC2D97F2000", "MODIFIED(Modified)", "301.000", "1"}
	if !reflect.DeepEqual(headers, wantHeaders) {
		t.Fatalf("headers = %#v, want %#v", headers, wantHeaders)
	}
	if !reflect.DeepEqual(row, wantRow) {
		t.Fatalf("row = %#v, want %#v", row, wantRow)
	}
}

func TestSingleOrderColumnsKeepsKnownOrderFields(t *testing.T) {
	order := &service.Order{
		OrderID: "FH1D27CBC2D97F2000",
		Symbol:  "US.AAPL",
		Side:    "BUY",
		Type:    "LIMIT",
		Price:   300,
		Qty:     1,
	}

	headers, row := singleOrderColumns(order)

	wantHeaders := []string{orderHeaderID, orderHeaderSymbol, orderHeaderSide, orderHeaderType, orderHeaderPrice, orderHeaderQty}
	wantRow := []string{"FH1D27CBC2D97F2000", "US.AAPL", "BUY(Buy)", "LIMIT(Limit)", "300.000", "1"}
	if !reflect.DeepEqual(headers, wantHeaders) {
		t.Fatalf("headers = %#v, want %#v", headers, wantHeaders)
	}
	if !reflect.DeepEqual(row, wantRow) {
		t.Fatalf("row = %#v, want %#v", row, wantRow)
	}
}

func TestOrderToRowFormatsMicrosecondTimestamp(t *testing.T) {
	const timestampSeconds = int64(1709251200)
	const timestampMicroseconds = int64(1709251200000000)

	order := &service.Order{
		OrderID:    "FH1D27CBC2D97F2000",
		Symbol:     "US.AAPL",
		Side:       "BUY",
		Type:       "LIMIT",
		Price:      300,
		Qty:        1,
		CreateTime: timestampMicroseconds,
	}

	row := orderToRow(order)

	if row[len(row)-1] != formatTimestamp(timestampSeconds) {
		t.Fatalf("formatted time = %q", row[len(row)-1])
	}
}

func TestBuildOrderDetailsRequest(t *testing.T) {
	oldFlags := orderFlags
	oldCfg := cfg
	defer func() {
		orderFlags = oldFlags
		cfg = oldCfg
	}()
	cfg = nil
	orderFlags.account = "123"
	orderFlags.exchange = "us"

	req, err := buildOrderDetailsRequest([]string{"ORD001", "ORD002"})
	if err != nil {
		t.Fatalf("buildOrderDetailsRequest() error = %v", err)
	}
	if req.Account != "123" || req.Exchange != "US" {
		t.Fatalf("request identity = %#v", req)
	}
	wantOrderIDs := []string{"ORD001", "ORD002"}
	if !reflect.DeepEqual(req.OrderIDs, wantOrderIDs) {
		t.Fatalf("order IDs = %#v, want %#v", req.OrderIDs, wantOrderIDs)
	}
}

func TestFormatDetailValueFormatsNestedValues(t *testing.T) {
	value := map[string]any{"name": "AAPL", "enabled": true}

	got := formatDetailValue("metadata", value)

	if !strings.Contains(got, `"name":"AAPL"`) || !strings.Contains(got, `"enabled":true`) {
		t.Fatalf("formatted detail value = %q", got)
	}
}

func TestFormatDetailValueFormatsExtraTimeFields(t *testing.T) {
	got := formatDetailValue("updated_time", float64(1709251200000000))

	if got != formatTimestamp(1709251200) {
		t.Fatalf("formatted time = %q", got)
	}
}
