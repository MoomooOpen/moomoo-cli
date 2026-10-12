package cmd

import (
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	expectedRootCommandName = "moomoo"
	expectedZshCompletion   = `"${fpath[1]}/_moomoo"`
	legacyZshCompletion     = `"${fpath[1]}/_futu"`
)

func TestRootCommandUsesMoomooBrand(t *testing.T) {
	if rootCmd.Use != expectedRootCommandName {
		t.Fatalf("root command Use = %q, want %q", rootCmd.Use, expectedRootCommandName)
	}
}

func TestCoreCommandTextUsesEnglish(t *testing.T) {
	texts := []string{
		messageUsingToken, messageRefreshed, messageLoginOK,
	}
	collectCommandText(rootCmd, &texts)
	for _, value := range texts {
		if strings.IndexFunc(value, func(r rune) bool { return unicode.Is(unicode.Han, r) }) >= 0 {
			t.Errorf("user-facing text contains Chinese characters: %q", value)
		}
	}
}

func collectCommandText(cmd *cobra.Command, texts *[]string) {
	*texts = append(*texts, cmd.Use, cmd.Short, cmd.Long, cmd.Example)
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		*texts = append(*texts, flag.Usage)
	})
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		*texts = append(*texts, flag.Usage)
	})
	for _, child := range cmd.Commands() {
		collectCommandText(child, texts)
	}
}

func TestCompletionInstructionsUseMoomooZshFilename(t *testing.T) {
	if !strings.Contains(completionCmd.Long, expectedZshCompletion) {
		t.Fatalf("completion instructions do not contain %q", expectedZshCompletion)
	}
	if strings.Contains(completionCmd.Long, legacyZshCompletion) {
		t.Fatalf("completion instructions still contain legacy filename %q", legacyZshCompletion)
	}
}

func TestMetadataCommandsDoNotRequireConfig(t *testing.T) {
	oldCfg := cfg
	oldReporter := requestReporter
	defer func() {
		flushRequestReporter()
		cfg = oldCfg
		requestReporter = oldReporter
	}()
	cfg = nil
	requestReporter = nil

	for _, command := range []*cobra.Command{versionCmd, completionCmd, updateCmd} {
		if err := initSharedState(command); err != nil {
			t.Fatalf("initSharedState(%s) error = %v", command.Name(), err)
		}
		if cfg != nil {
			t.Fatalf("%s initialized config-dependent state", command.Name())
		}
	}
}

func TestReportSourceHeaderValueForMoomoo(t *testing.T) {
	previous := reportSourceMode
	t.Cleanup(func() { reportSourceMode = previous })

	reportSourceMode = ""
	got, err := reportSourceHeaderValue()
	if err != nil {
		t.Fatalf("default reportSourceHeaderValue() error = %v", err)
	}
	if got != reportSourceMoomooCLI {
		t.Fatalf("default report source = %q, want %q", got, reportSourceMoomooCLI)
	}

	reportSourceMode = reportSourceModeCLI
	got, err = reportSourceHeaderValue()
	if err != nil {
		t.Fatalf("reportSourceHeaderValue() error = %v", err)
	}
	if got != reportSourceMoomooCLI {
		t.Fatalf("CLI report source = %q, want %q", got, reportSourceMoomooCLI)
	}

	reportSourceMode = reportSourceModeSkill
	got, err = reportSourceHeaderValue()
	if err != nil {
		t.Fatalf("reportSourceHeaderValue() error = %v", err)
	}
	if got != reportSourceMoomooSkill {
		t.Fatalf("skill report source = %q, want %q", got, reportSourceMoomooSkill)
	}
}

func TestReportSourceHeaderValueFallsBackToCLI(t *testing.T) {
	previous := reportSourceMode
	t.Cleanup(func() { reportSourceMode = previous })

	reportSourceMode = "unknown"
	got, err := reportSourceHeaderValue()
	if err != nil {
		t.Fatalf("reportSourceHeaderValue() error = %v", err)
	}
	if got != reportSourceMoomooCLI {
		t.Fatalf("invalid report source fallback = %q, want %q", got, reportSourceMoomooCLI)
	}
}

func TestReportSourceFlagIsHiddenAndRenamed(t *testing.T) {
	if rootCmd.PersistentFlags().Lookup("source") != nil {
		t.Fatal("source must be handled as an internal argument, not a cobra flag")
	}
	if rootCmd.PersistentFlags().Lookup("agent") != nil {
		t.Fatal("agent must be handled as an internal argument, not a cobra flag")
	}
	if rootCmd.PersistentFlags().Lookup("report-source") != nil {
		t.Fatal("legacy report-source flag must not be registered")
	}
}

func TestEffectiveAgentName(t *testing.T) {
	previousSource := reportSourceMode
	previousAgent := agentName
	clearAgentEnv(t)
	t.Cleanup(func() {
		reportSourceMode = previousSource
		agentName = previousAgent
	})

	reportSourceMode = ""
	agentName = ""
	if got := effectiveAgentName(); got != agentNameCLI {
		t.Fatalf("default agent = %q, want %q", got, agentNameCLI)
	}

	reportSourceMode = reportSourceModeSkill
	agentName = ""
	if got := effectiveAgentName(); got != agentNameUnknown {
		t.Fatalf("skill agent = %q, want %q", got, agentNameUnknown)
	}

	t.Setenv(envCodexSessionID, "codex-session")
	if got := effectiveAgentName(); got != agentNameCodex {
		t.Fatalf("codex env agent = %q, want %q", got, agentNameCodex)
	}

	t.Setenv(envCodexSessionID, "")
	t.Setenv(envClaudeCode, "1")
	if got := effectiveAgentName(); got != agentNameClaudeCode {
		t.Fatalf("claude code env agent = %q, want %q", got, agentNameClaudeCode)
	}

	agentName = agentNameCodex
	if got := effectiveAgentName(); got != agentNameCodex {
		t.Fatalf("explicit agent = %q, want %q", got, agentNameCodex)
	}
}

func clearAgentEnv(t *testing.T) {
	t.Helper()
	for _, key := range codexAgentEnvKeys {
		t.Setenv(key, "")
	}
	for _, key := range claudeCodeAgentEnvKeys {
		t.Setenv(key, "")
	}
}

func TestExtractInternalInvocationArgs(t *testing.T) {
	commandArgs := []string{"order", "detail", "O1"}
	tests := []struct {
		name       string
		args       []string
		wantSource string
		wantAgent  string
	}{
		{name: "source split beginning", args: []string{"--source", "skill", "order", "detail", "O1"}, wantSource: reportSourceModeSkill},
		{name: "source split middle", args: []string{"order", "--source", "skill", "detail", "O1"}, wantSource: reportSourceModeSkill},
		{name: "source split end", args: []string{"order", "detail", "O1", "--source", "skill"}, wantSource: reportSourceModeSkill},
		{name: "source equals beginning", args: []string{"--source=skill", "order", "detail", "O1"}, wantSource: reportSourceModeSkill},
		{name: "source equals middle", args: []string{"order", "--source=skill", "detail", "O1"}, wantSource: reportSourceModeSkill},
		{name: "source equals end", args: []string{"order", "detail", "O1", "--source=skill"}, wantSource: reportSourceModeSkill},
		{name: "source and agent split end", args: []string{"order", "detail", "O1", "--source", "skill", "--agent", "codex"}, wantSource: reportSourceModeSkill, wantAgent: agentNameCodex},
		{name: "invalid source split falls back", args: []string{"--source", "unknown", "order", "detail", "O1"}, wantSource: reportSourceModeCLI},
		{name: "invalid source equals falls back", args: []string{"order", "--source=unknown", "detail", "O1"}, wantSource: reportSourceModeCLI},
		{name: "agent equals normalizes hyphen", args: []string{"order", "detail", "O1", "--agent=claude-code"}, wantAgent: agentNameClaudeCode},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args, internal, err := extractInternalInvocationArgs(test.args)
			if err != nil {
				t.Fatalf("extractInternalInvocationArgs() error = %v", err)
			}
			if internal.source != test.wantSource {
				t.Fatalf("source = %q, want %q", internal.source, test.wantSource)
			}
			if internal.agent != test.wantAgent {
				t.Fatalf("agent = %q, want %q", internal.agent, test.wantAgent)
			}
			if !reflect.DeepEqual(args, commandArgs) {
				t.Fatalf("args = %#v, want %#v", args, commandArgs)
			}
		})
	}
}

func TestExtractInternalSourceArgWithoutMarker(t *testing.T) {
	input := []string{"order", "detail", "O1"}
	args, internal, err := extractInternalInvocationArgs(input)
	if err != nil {
		t.Fatalf("extractInternalInvocationArgs() error = %v", err)
	}
	if len(internal.source) != 0 {
		t.Fatalf("source = %q, want empty", internal.source)
	}
	if len(internal.agent) != 0 {
		t.Fatalf("agent = %q, want empty", internal.agent)
	}
	if !reflect.DeepEqual(args, input) {
		t.Fatalf("args = %#v, want %#v", args, input)
	}
}

func TestExtractInternalInvocationArgsRejectsMalformedMarker(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing value", args: []string{"version", "--source"}},
		{name: "empty equals value", args: []string{"version", "--source="}},
		{name: "missing value before flag", args: []string{"--source", "--format", "json", "version"}},
		{name: "duplicate split", args: []string{"--source", "skill", "version", "--source", "cli"}},
		{name: "duplicate equals", args: []string{"--source=skill", "version", "--source=cli"}},
		{name: "duplicate mixed", args: []string{"--source", "skill", "version", "--source=cli"}},
		{name: "missing agent value", args: []string{"version", "--agent"}},
		{name: "empty agent equals value", args: []string{"version", "--agent="}},
		{name: "missing agent value before flag", args: []string{"--agent", "--format", "json", "version"}},
		{name: "invalid agent", args: []string{"version", "--agent", "cursor"}},
		{name: "duplicate agent", args: []string{"--agent", "codex", "version", "--agent=claude_code"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := extractInternalInvocationArgs(test.args); err == nil {
				t.Fatal("expected internal marker error")
			}
		})
	}
}
