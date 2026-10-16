package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/MoomooOpen/moomoo-cli/internal/config"
	"github.com/MoomooOpen/moomoo-cli/internal/reporting"
)

func TestReportAndCheckUpdateDoesNotReportCommandExecution(t *testing.T) {
	oldReporter := requestReporter
	oldCfg := cfg
	oldOutputFormat := outputFormat
	t.Cleanup(func() {
		requestReporter = oldReporter
		cfg = oldCfg
		outputFormat = oldOutputFormat
	})

	received := make(chan struct{}, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	requestReporter = reporting.NewReporter(reporting.Options{Endpoint: srv.URL, Token: "token-1"})
	cfg = &config.Config{AutoUpdate: false}
	outputFormat = formatJSON

	reportAndCheckUpdate(&cobra.Command{Use: "quote"})

	select {
	case <-received:
		t.Fatal("command execution must not produce a report event")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestNormalizeOutputFormat(t *testing.T) {
	got, err := normalizeOutputFormat("JSON")
	if err != nil {
		t.Fatalf("normalizeOutputFormat() error = %v", err)
	}
	if got != formatJSON {
		t.Fatalf("format = %q, want %q", got, formatJSON)
	}

	if _, err := normalizeOutputFormat("josn"); err == nil {
		t.Fatal("expected invalid output format error")
	}
}
