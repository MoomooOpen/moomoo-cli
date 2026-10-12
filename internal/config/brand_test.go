package config

import (
	"path/filepath"
	"testing"
)

const expectedMoomooAPIBase = "https://webapi.moomoo.com"

func TestDefaultConfigUsesMoomooDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := DefaultConfig()
	if cfg.APIBase != expectedMoomooAPIBase {
		t.Fatalf("APIBase = %q, want %q", cfg.APIBase, expectedMoomooAPIBase)
	}
	wantTokenFile := filepath.Join(ConfigDir(), tokenFileName)
	if cfg.TokenFile != wantTokenFile {
		t.Fatalf("TokenFile = %q, want %q", cfg.TokenFile, wantTokenFile)
	}
	if filepath.Base(ConfigDir()) != configDirName {
		t.Fatalf("config directory = %q, want base name %q", ConfigDir(), configDirName)
	}
}

func TestNormalizeAPIBaseUsesDefaultForEmptyValue(t *testing.T) {
	if got := NormalizeAPIBase(""); got != expectedMoomooAPIBase {
		t.Fatalf("NormalizeAPIBase(\"\") = %q, want %q", got, expectedMoomooAPIBase)
	}
}

func TestNormalizeAPIBaseTrimsSpaceAndTrailingSlash(t *testing.T) {
	got := NormalizeAPIBase(" https://example.moomoo.com/ ")
	want := "https://example.moomoo.com"
	if got != want {
		t.Fatalf("NormalizeAPIBase() = %q, want %q", got, want)
	}
}

func TestTelemetryIsNotConfigurable(t *testing.T) {
	cfg := DefaultConfig()
	if got := cfg.Get("telemetry"); got != "" {
		t.Fatalf("telemetry config value = %q, want empty", got)
	}
	if err := cfg.Set("telemetry", "false"); err == nil {
		t.Fatal("expected telemetry config key to be rejected")
	}
}
