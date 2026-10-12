package cmd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/MoomooOpen/moomoo-cli/internal/reporting"
)

const (
	commandReportAuthHeader          = "Bearer token-1"
	commandReportCommandUse          = "quote"
	commandReportToken               = "token-1"
	expectedCommandReportCLISource   = int32(6)
	expectedCommandReportSkillSource = int32(8)
	expectedCommandReportVersion     = "1.2.3"
)

func TestReportCommandExecutionDefaultsToCLISource(t *testing.T) {
	event := runCommandReportTest(t, "", "", expectedCommandReportCLISource, agentNameCLI)
	assertCommandReportEvent(t, event, expectedCommandReportCLISource, agentNameCLI)
}

func TestReportCommandExecutionUsesSkillSource(t *testing.T) {
	event := runCommandReportTest(t, reportSourceModeSkill, agentNameCodex, expectedCommandReportSkillSource, agentNameCodex)
	assertCommandReportEvent(t, event, expectedCommandReportSkillSource, agentNameCodex)
}

func runCommandReportTest(t *testing.T, sourceMode, reportAgent string, wantSource int32, wantAgent string) reporting.Event {
	t.Helper()
	restoreCommandReportGlobals(t)
	received := make(chan reporting.Event, 1)
	srv := newCommandReportServer(t, received, wantSource, wantAgent)
	defer srv.Close()

	configureCommandReport(sourceMode, reportAgent, srv.URL)
	reportCommandExecution(&cobra.Command{Use: commandReportCommandUse}, commandOutcomeSuccess, nil)
	closeCommandReport(t)
	return receiveCommandReport(t, received)
}

func restoreCommandReportGlobals(t *testing.T) {
	t.Helper()
	oldReporter := requestReporter
	oldSource := reportSourceMode
	oldAgent := agentName
	oldStart := cmdStartTime
	oldVersion := Version
	oldActiveCommand := activeCommand
	t.Cleanup(func() {
		requestReporter = oldReporter
		reportSourceMode = oldSource
		agentName = oldAgent
		cmdStartTime = oldStart
		Version = oldVersion
		activeCommand = oldActiveCommand
	})
}

func newCommandReportServer(t *testing.T, received chan<- reporting.Event, wantSource int32, wantAgent string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != commandReportAuthHeader {
			t.Fatalf("Authorization = %q, want %s", got, commandReportAuthHeader)
		}
		var payload struct {
			Event reporting.Event `json:"event"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode report request: %v", err)
		}
		if payload.Event.Source != wantSource || payload.Event.AgentName != wantAgent {
			t.Fatalf("source metadata = %+v, want source %d agent %s", payload.Event, wantSource, wantAgent)
		}
		received <- payload.Event
		w.WriteHeader(http.StatusOK)
	}))
}

func configureCommandReport(sourceMode, reportAgent, endpoint string) {
	requestReporter = reporting.NewReporter(reporting.Options{
		Endpoint: endpoint,
		Token:    commandReportToken,
	})
	reportSourceMode = sourceMode
	agentName = reportAgent
	Version = expectedCommandReportVersion
	cmdStartTime = time.Now().Add(-time.Second)
}

func closeCommandReport(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := requestReporter.Close(ctx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	requestReporter = nil
}

func receiveCommandReport(t *testing.T, received <-chan reporting.Event) reporting.Event {
	t.Helper()
	select {
	case event := <-received:
		return event
	case <-time.After(time.Second):
		t.Fatal("expected command report request")
	}
	return reporting.Event{}
}

func assertCommandReportEvent(t *testing.T, event reporting.Event, wantSource int32, wantAgent string) {
	t.Helper()
	if event.Source != wantSource {
		t.Fatalf("source = %d, want %d", event.Source, wantSource)
	}
	if event.AgentName != wantAgent {
		t.Fatalf("agent_name = %q, want %q", event.AgentName, wantAgent)
	}
	if event.Platform != reportPlatformMoomoo {
		t.Fatalf("platform = %q, want %q", event.Platform, reportPlatformMoomoo)
	}
	if event.Command != commandReportCommandUse {
		t.Fatalf("command = %q, want %s", event.Command, commandReportCommandUse)
	}
	if event.ToolID != commandReportToolID {
		t.Fatalf("tool_id = %q, want %q", event.ToolID, commandReportToolID)
	}
	if event.Outcome != commandOutcomeSuccess {
		t.Fatalf("outcome = %q, want %q", event.Outcome, commandOutcomeSuccess)
	}
	if event.Version != expectedCommandReportVersion {
		t.Fatalf("version = %q, want %q", event.Version, expectedCommandReportVersion)
	}
	if event.OS == "" || event.Arch == "" || event.DurationMs <= 0 {
		t.Fatalf("expected os, arch and duration in event: %+v", event)
	}
}
