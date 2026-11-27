package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/MoomooOpen/moomoo-cli/internal/client"
	"github.com/MoomooOpen/moomoo-cli/internal/service"
	"github.com/spf13/cobra"
)

func TestQuoteCodeQueryParametersKeepSymbolFlag(t *testing.T) {
	cases := []struct {
		id, input, code string
		args            []string
	}{
		{"quote.crypto_institution_holding_list", "cc.btc", "CC.BTC", nil},
		{"quote.crypto_institution_holding_overview", "cc.btc", "CC.BTC", nil},
		{"quote.institution_transaction_record", "us.aapl", "US.AAPL", []string{"--institution-id", "403413"}},
		{"news.by_security", "us.aapl", "US.AAPL", nil},
	}
	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			testQuoteCodeQueryParameter(t, tc.id, tc.input, tc.code, tc.args)
		})
	}
}

func testQuoteCodeQueryParameter(t *testing.T, id, input, code string, args []string) {
	t.Helper()
	var spec *quoteRESTSpec
	specs := loadQuoteRESTSpecs()
	for i := range specs {
		if specs[i].ID == id {
			spec = &specs[i]
			break
		}
	}
	if spec == nil {
		t.Fatalf("missing spec %s", id)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != spec.Path || r.URL.Query().Get("code") != code || r.URL.Query().Has("symbol") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
		}
		fmt.Fprint(w, `{"ret_code":0,"data":{}}`)
	}))
	defer server.Close()
	previous := quoteSvc
	quoteSvc = service.NewQuoteService(client.New(server.URL, client.WithToken("test")))
	defer func() { quoteSvc = previous }()
	command := newQuoteRESTCommand(spec)
	command.SetArgs(append([]string{"--symbol", input}, args...))
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestQuoteRESTSearchKeywordFlagUsesSymbolQuery(t *testing.T) {
	spec := quoteRESTSearchTestSpec()
	cmd := &cobra.Command{}
	param := spec.Params[0]
	addQuoteRESTFlag(cmd, param, param.Kind)
	if cmd.Flags().Lookup("keyword") == nil {
		t.Fatal("keyword flag is not registered")
	}
	if cmd.Flags().Lookup("symbol") != nil {
		t.Fatal("symbol flag must not be registered for search keyword")
	}
	if usage := cmd.Flags().Lookup("keyword").Usage; strings.Contains(usage, "MARKET.CODE") {
		t.Fatalf("keyword usage must not require symbol format: %q", usage)
	}
	if err := cmd.Flags().Set("keyword", "腾讯"); err != nil {
		t.Fatalf("set keyword flag: %v", err)
	}

	queryParam, keyword, err := resolveQuoteRESTKeyword(cmd, nil, &spec)
	if err != nil {
		t.Fatalf("resolveQuoteRESTKeyword() error = %v", err)
	}
	if queryParam.Name != "symbol" || keyword != "腾讯" {
		t.Fatalf("query param = %q keyword = %q, want symbol and 腾讯", queryParam.Name, keyword)
	}
	query := url.Values{}
	putRESTValue(query, map[string]any{}, false, queryParam, keyword)
	if got := query.Get("symbol"); got != "腾讯" {
		t.Fatalf("symbol query = %q, want 腾讯", got)
	}
	if got := query.Get("keyword"); got != "" {
		t.Fatalf("keyword query = %q, want empty", got)
	}
}

func TestQuoteRESTSearchKeywordDoesNotNormalizePositionalSymbol(t *testing.T) {
	spec := quoteRESTSearchTestSpec()
	cmd := &cobra.Command{}
	addQuoteRESTFlag(cmd, spec.Params[0], spec.Params[0].Kind)

	queryParam, keyword, err := resolveQuoteRESTKeyword(cmd, []string{"AAPL"}, &spec)
	if err != nil {
		t.Fatalf("resolveQuoteRESTKeyword() error = %v", err)
	}
	if queryParam.Name != "symbol" || keyword != "AAPL" {
		t.Fatalf("query param = %q keyword = %q, want symbol and AAPL", queryParam.Name, keyword)
	}
}

func TestQuoteRESTSearchKeywordKeepsDefault(t *testing.T) {
	spec := quoteRESTSearchTestSpec()
	cmd := &cobra.Command{}
	addQuoteRESTFlag(cmd, spec.Params[0], spec.Params[0].Kind)

	_, keyword, err := resolveQuoteRESTKeyword(cmd, nil, &spec)
	if err != nil {
		t.Fatalf("resolveQuoteRESTKeyword() error = %v", err)
	}
	if keyword != "00700" {
		t.Fatalf("keyword default = %q, want 00700", keyword)
	}
}

func quoteRESTSearchTestSpec() quoteRESTSpec {
	return quoteRESTSpec{
		ID:         "quote.stock_feed",
		Cmd:        "stock-feed",
		Group:      quoteGroupID,
		Path:       "/api/v1.0/quote/search-quote",
		Method:     "GET",
		Positional: "keyword",
		Extra:      "symbol",
		Params: []quoteRESTParam{
			{
				Name:     "symbol",
				Kind:     "string",
				Required: true,
				Default:  "00700",
				Source:   "query",
				Flag:     quoteRESTKeywordFlag,
			},
		},
	}
}

func TestSearchAggregateSupportsNews(t *testing.T) {
	var search quoteRESTAggregate
	for _, aggregate := range quoteRESTAggregates {
		if aggregate.Cmd == "search" {
			search = aggregate
			break
		}
	}
	specs := make(map[string]*quoteRESTSpec)
	for i := range quoteRESTSpecs {
		specs[quoteRESTSpecs[i].Cmd] = &quoteRESTSpecs[i]
	}
	command := newQuoteRESTAggregateCommand(search, specs)
	if command == nil {
		t.Fatal("search command not registered")
	}
	variants := make(map[string]*quoteRESTSpec)
	for _, variant := range search.Variants {
		variants[variant.Value] = specs[variant.Cmd]
	}
	selected, err := selectedQuoteRESTSpec(command, search, variants)
	if err != nil || selected.ID != "quote.community_search" {
		t.Fatalf("default = %#v, %v", selected, err)
	}
	if err := command.Flags().Set("type", "news"); err != nil {
		t.Fatal(err)
	}
	if selected, err := selectedQuoteRESTSpec(command, search, variants); err != nil || selected.ID != "quote.news_search" {
		t.Fatalf("news = %#v, %v", selected, err)
	}
}
