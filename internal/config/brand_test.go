package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadMigratesNonDefaultAPIBaseToCurrentDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := configFilePath()
	if err := ensureDir(filepath.Dir(path)); err != nil {
		t.Fatalf("ensureDir() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(`{
		"api_base": "https://uat-webapi.moomoo.com",
		"client_id": "uat-client-id",
		"client_id_api_base": "https://uat-webapi.moomoo.com",
		"redirect_uri": "http://127.0.0.1:18888/callback",
		"token_file": "/tmp/token.json",
		"auto_update": true
	}`), filePermissions); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIBase != DefaultAPIBase {
		t.Fatalf("APIBase = %q, want %q", cfg.APIBase, DefaultAPIBase)
	}
	if cfg.ClientID != "uat-client-id" || cfg.ClientIDAPIBase != "https://uat-webapi.moomoo.com" {
		t.Fatalf("OAuth client binding should be preserved until the next login: %+v", cfg)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(data), `"api_base": "https://webapi.moomoo.com"`) {
		t.Fatalf("persisted config does not contain migrated API base: %s", data)
	}
	if !strings.Contains(string(data), `"client_id_api_base": "https://uat-webapi.moomoo.com"`) {
		t.Fatalf("persisted config should keep OAuth client binding until login refreshes it: %s", data)
	}
}

func TestLoadMigratesCustomAPIBaseToCurrentDefault(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := configFilePath()
	if err := ensureDir(filepath.Dir(path)); err != nil {
		t.Fatalf("ensureDir() error = %v", err)
	}
	const customBase = "https://sandbox.moomoo.com"
	if err := os.WriteFile(path, []byte(`{
		"api_base": "`+customBase+`",
		"client_id": "custom-client-id",
		"client_id_api_base": "`+customBase+`",
		"token_file": "/tmp/token.json",
		"auto_update": true
	}`), filePermissions); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.APIBase != DefaultAPIBase {
		t.Fatalf("APIBase = %q, want %q", cfg.APIBase, DefaultAPIBase)
	}
	if cfg.ClientID != "custom-client-id" || cfg.ClientIDAPIBase != customBase {
		t.Fatalf("OAuth client binding should be preserved until the next login: %+v", cfg)
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
