package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MoomooOpen/moomoo-cli/internal/client"
)

func TestWeb3GetBalanceRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertWeb3Request(t, request, http.MethodGet, "/v1/web3/wallet/balance")
		if got := request.URL.Query().Get("address"); got != "0xabc" {
			t.Errorf("address = %q, want %q", got, "0xabc")
		}
		fmt.Fprint(writer, `{"code":0,"message":"","data":{"list":[{"token":"USDC","balance":12.5,"value_usd":12.5}]}}`)
	}))
	defer server.Close()

	balances, err := NewWeb3Service(client.New(server.URL)).GetBalance(context.Background(), "0xabc")
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if len(balances) != 1 || balances[0].Token != "USDC" || balances[0].Balance != 12.5 {
		t.Fatalf("balances = %#v", balances)
	}
}

func TestWeb3GetSwapQuoteRequest(t *testing.T) {
	server := newWeb3JSONServer(t, http.MethodPost, "/v1/web3/swap/quote", map[string]any{
		"from_token": "USDC", "to_token": "ETH", "amount": "25.500000000000000001",
	}, `{"code":0,"message":"","data":{"from_token":"USDC","to_token":"ETH","from_amount":25.5,"to_amount":0.01,"rate":0.000392,"price_impact":0.2}}`)
	defer server.Close()

	quote, err := NewWeb3Service(client.New(server.URL)).GetSwapQuote(context.Background(), "USDC", "ETH", "25.500000000000000001")
	if err != nil {
		t.Fatalf("GetSwapQuote() error = %v", err)
	}
	if quote.FromToken != "USDC" || quote.ToToken != "ETH" || quote.FromAmt != 25.5 {
		t.Fatalf("quote = %#v", quote)
	}
}

func TestWeb3ExecuteSwapRequest(t *testing.T) {
	server := newWeb3JSONServer(t, http.MethodPost, "/v1/web3/swap/execute", map[string]any{
		"from_token": "USDC", "to_token": "ETH", "amount": "25.500000000000000001",
	}, `{"code":0,"message":"","data":{"tx_hash":"0xswap","status":"submitted","from_token":"USDC","to_token":"ETH","from_amount":25.5,"to_amount":0.01}}`)
	defer server.Close()

	result, err := NewWeb3Service(client.New(server.URL)).ExecuteSwap(context.Background(), "USDC", "ETH", "25.500000000000000001")
	if err != nil {
		t.Fatalf("ExecuteSwap() error = %v", err)
	}
	if result.TxHash != "0xswap" || result.Status != "submitted" {
		t.Fatalf("result = %#v", result)
	}
}

func TestWeb3TransferRequest(t *testing.T) {
	server := newWeb3JSONServer(t, http.MethodPost, "/v1/web3/transfer", map[string]any{
		"to": "0xdef", "token": "USDC", "amount": "3.250000000000000001",
	}, `{"code":0,"message":"","data":{"tx_hash":"0xtransfer","status":"submitted"}}`)
	defer server.Close()

	result, err := NewWeb3Service(client.New(server.URL)).Transfer(context.Background(), "0xdef", "USDC", "3.250000000000000001")
	if err != nil {
		t.Fatalf("Transfer() error = %v", err)
	}
	if result.TxHash != "0xtransfer" || result.Status != "submitted" {
		t.Fatalf("result = %#v", result)
	}
}

func TestWeb3MethodsReturnAPIErrors(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		call   func(*Web3Service) error
	}{
		{name: "balance", method: http.MethodGet, path: "/v1/web3/wallet/balance", call: func(service *Web3Service) error {
			_, err := service.GetBalance(context.Background(), "0xabc")
			return err
		}},
		{name: "quote", method: http.MethodPost, path: "/v1/web3/swap/quote", call: func(service *Web3Service) error {
			_, err := service.GetSwapQuote(context.Background(), "USDC", "ETH", "1")
			return err
		}},
		{name: "execute", method: http.MethodPost, path: "/v1/web3/swap/execute", call: func(service *Web3Service) error {
			_, err := service.ExecuteSwap(context.Background(), "USDC", "ETH", "1")
			return err
		}},
		{name: "transfer", method: http.MethodPost, path: "/v1/web3/transfer", call: func(service *Web3Service) error {
			_, err := service.Transfer(context.Background(), "0xdef", "USDC", "1")
			return err
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				assertWeb3Request(t, request, test.method, test.path)
				fmt.Fprint(writer, `{"code":403,"message":"permission denied"}`)
			}))
			defer server.Close()

			err := test.call(NewWeb3Service(client.New(server.URL)))
			if err == nil {
				t.Fatal("expected API error")
			}
		})
	}
}

func newWeb3JSONServer(
	t *testing.T, method string, path string, wantBody map[string]any, response string,
) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertWeb3Request(t, request, method, path)
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		for key, want := range wantBody {
			if got := body[key]; got != want {
				t.Errorf("body[%q] = %#v, want %#v", key, got, want)
			}
		}
		if len(body) != len(wantBody) {
			t.Errorf("body = %#v, want exactly %#v", body, wantBody)
		}
		fmt.Fprint(writer, response)
	}))
}

func assertWeb3Request(t *testing.T, request *http.Request, method string, path string) {
	t.Helper()
	if request.Method != method {
		t.Errorf("method = %s, want %s", request.Method, method)
	}
	if request.URL.Path != path {
		t.Errorf("path = %s, want %s", request.URL.Path, path)
	}
}
