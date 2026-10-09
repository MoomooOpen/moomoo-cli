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
	if cfg.ReportHost != DefaultReportHost {
		t.Fatalf("ReportHost = %q, want %q", cfg.ReportHost, DefaultReportHost)
	}
}
