package cmd

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/MoomooOpen/moomoo-cli/internal/update"
)

func TestPrintUpdateInfoIncludesCurrentAndLatestVersion(t *testing.T) {
	output := captureUpdateStdout(t, func() {
		printUpdateInfo(&update.UpdateInfo{
			CurrentVersion: "v0.1.0",
			LatestVersion:  "v0.2.0",
			ReleaseNotes:   "https://github.com/MoomooOpen/moomoo-cli/releases/tag/v0.2.0",
		})
	})
	if !strings.Contains(output, "Update available: v0.1.0 -> v0.2.0") {
		t.Fatalf("update output does not include version transition: %q", output)
	}
}

func captureUpdateStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = oldStdout })

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(data)
}
