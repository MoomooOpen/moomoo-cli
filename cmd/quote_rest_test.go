package cmd

import (
	"net/url"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

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
		ID:         "quote.news_search",
		Cmd:        "news-search",
		Group:      quoteGroupID,
		Path:       "/api/v1.0/quote/find-news",
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
