package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/MoomooOpen/moomoo-cli/internal/auth"
	"github.com/MoomooOpen/moomoo-cli/internal/client"
	"github.com/MoomooOpen/moomoo-cli/internal/config"
	"github.com/MoomooOpen/moomoo-cli/internal/output"
	"github.com/MoomooOpen/moomoo-cli/internal/reporting"
	"github.com/MoomooOpen/moomoo-cli/internal/service"
	"github.com/MoomooOpen/moomoo-cli/internal/update"
)

// Version is set at build time via -ldflags.
var Version = "dev"

// Output format constants.
const (
	formatTable = "table"
	formatJSON  = "json"
)

const (
	quoteGroupID = "quote"
	simGroupID   = "sim"
)

const (
	reportSourceModeCLI     = "cli"
	reportSourceModeSkill   = "skill"
	reportSourceMoomooCLI   = "6"
	reportSourceMoomooSkill = "8"
	reportPlatformMoomoo    = "moomoo"
	commandReportToolID     = "cli_command"
	commandOutcomeSuccess   = "success"
	commandOutcomeFailure   = "failure"
	internalSourceFlag      = "--source"
	internalSourcePrefix    = "--source="
	internalAgentFlag       = "--agent"
	internalAgentPrefix     = "--agent="
	agentNameCLI            = "cli"
	agentNameCodex          = "codex"
	agentNameClaudeCode     = "claude_code"
	agentNameUnknown        = "unknown"
	envCodexSessionID       = "CODEX_SESSION_ID"
	envCodexVersion         = "CODEX_VERSION"
	envCodexSandbox         = "CODEX_SANDBOX"
	envClaudeCode           = "CLAUDECODE"
	envClaudeCodeAlt        = "CLAUDE_CODE"
	envClaudeCodeSession    = "CLAUDE_CODE_SESSION_ID"
	envClaudeCodeEntrypoint = "CLAUDE_CODE_ENTRYPOINT"
)

// Unix timestamp magnitude thresholds.
const (
	millisecondTimestampMin   = int64(1_000_000_000_000)
	microsecondTimestampMin   = int64(1_000_000_000_000_000)
	nanosecondTimestampMin    = int64(1_000_000_000_000_000_000)
	nanosecondsPerMillisecond = int64(time.Millisecond)
	nanosecondsPerMicrosecond = int64(time.Microsecond)
)

// Package-level shared state populated by PersistentPreRunE.
var (
	outputFormat     string
	reportSourceMode string
	agentName        string
	tokenFile        string
	cfg              *config.Config
	httpClient       *client.Client
	quoteSvc         *service.QuoteService
	tradeSvc         *service.TradeService
	cryptoSvc        *service.CryptoService
	requestReporter  *reporting.Reporter
	cmdStartTime     time.Time
	activeCommand    *cobra.Command
)

var (
	codexAgentEnvKeys = []string{
		envCodexSessionID,
		envCodexVersion,
		envCodexSandbox,
	}
	claudeCodeAgentEnvKeys = []string{
		envClaudeCode,
		envClaudeCodeAlt,
		envClaudeCodeSession,
		envClaudeCodeEntrypoint,
	}
)

// Commands that do not require authentication.
var noAuthCommands = map[string]bool{
	"auth":       true,
	"login":      true,
	"logout":     true,
	"status":     true,
	"version":    true,
	"completion": true,
	"config":     true,
	"check":      true,
	"update":     true,
	"help":       true,
}

var noConfigCommands = map[string]bool{
	"version":    true,
	"completion": true,
	"update":     true,
	"help":       true,
}

var rootCmd = &cobra.Command{
	Use:   "moomoo",
	Short: "moomoo command-line interface",
	Long:  "moomoo CLI V1: REST API access for market data, trading, and crypto.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		cmdStartTime = time.Now()
		activeCommand = cmd
		return initSharedState(cmd)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		reportAndCheckUpdate(cmd)
	},
}

// Execute runs the root command.
func Execute() error {
	args, internal, err := extractInternalInvocationArgs(os.Args[1:])
	if err != nil {
		return err
	}
	reportSourceMode = internal.source
	agentName = internal.agent
	rootCmd.SetArgs(args)
	defer flushRequestReporter()
	err = rootCmd.Execute()
	if err != nil {
		reportCommandExecution(activeCommand, commandOutcomeFailure, err)
	}
	return err
}

func init() {
	rootCmd.AddGroup(&cobra.Group{ID: quoteGroupID, Title: "Quote Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: simGroupID, Title: "Sim Trade Commands:"})
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", formatTable,
		"Output format: table or json (default: table)")
	rootCmd.PersistentFlags().StringVar(&tokenFile, "token-file", "",
		"Path to the encrypted OAuth token file (default: token_file from config)")
}

// initSharedState loads config, ensures auth (when needed), and initializes
// the HTTP client and services.
func initSharedState(cmd *cobra.Command) error {
	if commandOrParentMatches(cmd, noConfigCommands) {
		newRequestReporter("")
		return nil
	}
	var err error
	cfg, err = config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cfg.AutoUpdate {
		update.RefreshCacheIfStale(Version)
	}

	newRequestReporter("")
	if requiresAuth(cmd) {
		return initAuthenticatedClient()
	}
	return nil
}

// requiresAuth returns false for commands that can run without a token.
// It walks up the command tree so that subcommands (e.g. "config init")
// inherit the no-auth status of their parent.
func requiresAuth(cmd *cobra.Command) bool {
	return !commandOrParentMatches(cmd, noAuthCommands)
}

func commandOrParentMatches(cmd *cobra.Command, commands map[string]bool) bool {
	for current := cmd; current != nil; current = current.Parent() {
		if commands[current.Name()] {
			return true
		}
	}
	return false
}

// initAuthenticatedClient ensures credentials are configured and sets up the
// HTTP client and all service instances using OAuth2.
func initAuthenticatedClient() error {
	return initOAuthClient()
}

// initOAuthClient sets up the client using OAuth token authentication.
func initOAuthClient() error {
	if len(cfg.ClientID) == 0 {
		return fmt.Errorf("client_id is not configured\nRun 'moomoo config init' to register an OAuth client")
	}
	if !clientIDMatchesCurrentAPIBase() {
		return fmt.Errorf("client_id is not bound to current api_base %s\nRun 'moomoo auth login' to register a new OAuth client", currentAPIBase())
	}

	token, err := auth.EnsureValidWithBase(currentAPIBase(), cfg.ClientID, effectiveTokenFile())
	if err != nil {
		return fmt.Errorf("authentication required: %w\nRun 'moomoo auth login' to authenticate", err)
	}
	reportSource, err := reportSourceHeaderValue()
	if err != nil {
		return err
	}
	httpClient = client.New(
		currentAPIBase(),
		client.WithToken(token),
		client.WithClientName(auth.BuildClientName("")),
		client.WithVersion(Version),
		client.WithReportSource(reportSource),
		client.WithReportMetadata(reportPlatformMoomoo, effectiveAgentName()),
		client.WithRequestReporter(newRequestReporter(token)),
	)
	initServices()
	return nil
}

func newRequestReporter(token string) *reporting.Reporter {
	flushRequestReporter()
	requestReporter = reporting.NewReporter(reporting.Options{
		Host:  reporting.DefaultHost,
		Token: token,
	})
	return requestReporter
}

func reportSourceHeaderValue() (string, error) {
	mode := effectiveReportSourceMode()
	switch mode {
	case reportSourceModeCLI:
		return reportSourceMoomooCLI, nil
	case reportSourceModeSkill:
		return reportSourceMoomooSkill, nil
	default:
		return reportSourceMoomooCLI, nil
	}
}

func effectiveReportSourceMode() string {
	return normalizeReportSourceMode(reportSourceMode)
}

func normalizeReportSourceMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case reportSourceModeSkill:
		return reportSourceModeSkill
	case reportSourceModeCLI:
		return reportSourceModeCLI
	default:
		return reportSourceModeCLI
	}
}

func effectiveAgentName() string {
	if len(agentName) > 0 {
		return agentName
	}
	if effectiveReportSourceMode() != reportSourceModeSkill {
		return agentNameCLI
	}
	if detected := detectAgentNameFromEnv(); len(detected) > 0 {
		return detected
	}
	return agentNameUnknown
}

func detectAgentNameFromEnv() string {
	if hasAnyNonEmptyEnv(codexAgentEnvKeys) {
		return agentNameCodex
	}
	if hasAnyNonEmptyEnv(claudeCodeAgentEnvKeys) {
		return agentNameClaudeCode
	}
	return ""
}

func hasAnyNonEmptyEnv(keys []string) bool {
	for _, key := range keys {
		if len(os.Getenv(key)) > 0 {
			return true
		}
	}
	return false
}

type internalInvocationArgs struct {
	source string
	agent  string
}

func extractInternalInvocationArgs(args []string) ([]string, internalInvocationArgs, error) {
	filtered := make([]string, 0, len(args))
	internal := internalInvocationArgs{}
	foundSource := false
	foundAgent := false
	for index := 0; index < len(args); index++ {
		arg, value, consumed, ok, err := parseInternalArg(args, index)
		if err != nil {
			return nil, internalInvocationArgs{}, err
		}
		if !ok {
			filtered = append(filtered, args[index])
			continue
		}
		if consumed {
			index++
		}
		if err := applyInternalArg(arg, value, &internal, &foundSource, &foundAgent); err != nil {
			return nil, internalInvocationArgs{}, err
		}
	}
	return filtered, internal, nil
}

func parseInternalArg(args []string, index int) (string, string, bool, bool, error) {
	arg := args[index]
	switch {
	case arg == internalSourceFlag || arg == internalAgentFlag:
		if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
			return "", "", false, false, fmt.Errorf("%s requires a value", arg)
		}
		return arg, args[index+1], true, true, nil
	case strings.HasPrefix(arg, internalSourcePrefix):
		return parseInternalEqualsArg(internalSourceFlag, internalSourcePrefix, arg)
	case strings.HasPrefix(arg, internalAgentPrefix):
		return parseInternalEqualsArg(internalAgentFlag, internalAgentPrefix, arg)
	default:
		return "", "", false, false, nil
	}
}

func parseInternalEqualsArg(flag, prefix, arg string) (string, string, bool, bool, error) {
	value := strings.TrimPrefix(arg, prefix)
	if len(value) == 0 {
		return "", "", false, false, fmt.Errorf("%s requires a value", flag)
	}
	return flag, value, false, true, nil
}

func applyInternalArg(arg, value string, internal *internalInvocationArgs, foundSource, foundAgent *bool) error {
	switch arg {
	case internalSourceFlag:
		if *foundSource {
			return fmt.Errorf("%s may only be specified once", internalSourceFlag)
		}
		internal.source = normalizeReportSourceMode(value)
		*foundSource = true
	case internalAgentFlag:
		normalized, err := normalizeAgentName(value)
		if err != nil {
			return err
		}
		if *foundAgent {
			return fmt.Errorf("%s may only be specified once", internalAgentFlag)
		}
		internal.agent = normalized
		*foundAgent = true
	}
	return nil
}

func normalizeAgentName(value string) (string, error) {
	normalized := strings.ReplaceAll(strings.ToLower(strings.TrimSpace(value)), "-", "_")
	switch normalized {
	case agentNameCLI, agentNameCodex, agentNameClaudeCode, agentNameUnknown:
		return normalized, nil
	default:
		return "", fmt.Errorf("invalid agent %q: supported values are %s, %s, %s and %s",
			value, agentNameCLI, agentNameCodex, agentNameClaudeCode, agentNameUnknown)
	}
}

func effectiveTokenFile() string {
	if len(tokenFile) > 0 {
		return tokenFile
	}
	return cfg.TokenFile
}

// initServices creates service instances from the shared HTTP client.
func initServices() {
	quoteSvc = service.NewQuoteService(httpClient)
	tradeSvc = service.NewTradeService(httpClient)
	cryptoSvc = service.NewCryptoService(httpClient)
}

// reportAndCheckUpdate flushes request reporting and prints an update notice if available.
func reportAndCheckUpdate(cmd *cobra.Command) {
	defer flushRequestReporter()
	reportCommandExecution(cmd, commandOutcomeSuccess, nil)

	if cfg != nil && cfg.AutoUpdate && !isJSONOutput() {
		update.NotifyIfAvailable(Version)
		update.NotifyReleaseNotesIfVersionChanged(Version)
	}
}

func reportCommandExecution(cmd *cobra.Command, outcome string, commandErr error) {
	if requestReporter == nil || cmd == nil {
		return
	}
	reportSource, err := reportSourceHeaderValue()
	if err != nil {
		return
	}
	errorText := ""
	if commandErr != nil {
		errorText = commandErr.Error()
	}
	event := reporting.BuildEvent(reporting.BuildInput{
		Source:     reportSource,
		AgentName:  effectiveAgentName(),
		Platform:   reportPlatformMoomoo,
		Command:    cmd.CommandPath(),
		DurationMs: commandDurationMillis(),
		Error:      errorText,
		Version:    Version,
	})
	event.ToolID = commandReportToolID
	event.Outcome = outcome
	requestReporter.Report(event)
}

func commandDurationMillis() int64 {
	if cmdStartTime.IsZero() {
		return 0
	}
	return time.Since(cmdStartTime).Milliseconds()
}

func flushRequestReporter() {
	if requestReporter == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), reporting.DefaultFlushTimeout)
	defer cancel()
	_ = requestReporter.Close(ctx)
	requestReporter = nil
}

// isJSONOutput returns true when the user requested JSON output format.
func isJSONOutput() bool {
	return outputFormat == formatJSON
}

// printResult outputs data either as JSON or via the provided table renderer.
func printResult(data any, renderTable func()) {
	if isJSONOutput() {
		_ = output.PrintJSON(os.Stdout, data)
		return
	}
	renderTable()
}

// resolveAccount returns the specified account or falls back to the default.
func resolveAccount(account string) string {
	if len(account) > 0 {
		return account
	}
	if cfg != nil {
		return cfg.DefaultAcct
	}
	return ""
}

// formatTimestamp converts a Unix timestamp to a readable string.
// It accepts seconds, milliseconds, microseconds, or nanoseconds.
func formatTimestamp(ts int64) string {
	if ts == 0 {
		return "-"
	}
	switch {
	case ts >= nanosecondTimestampMin:
		return time.Unix(0, ts).Format("2006-01-02 15:04:05")
	case ts >= microsecondTimestampMin:
		return time.Unix(0, ts*nanosecondsPerMicrosecond).Format("2006-01-02 15:04:05")
	case ts >= millisecondTimestampMin:
		return time.Unix(0, ts*nanosecondsPerMillisecond).Format("2006-01-02 15:04:05")
	default:
		return time.Unix(ts, 0).Format("2006-01-02 15:04:05")
	}
}

// cmdContext returns a context annotated with the command name.
func cmdContext(cmd *cobra.Command) context.Context {
	return context.WithValue(
		context.Background(),
		client.CtxKeyCommand,
		cmd.CommandPath(),
	)
}
