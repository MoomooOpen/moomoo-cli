package cmd

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestOpenBrowserPreservesAuthorizationURL(t *testing.T) {
	const authorizationURL = "https://example.com/oauth2/authorize/confirm?client_id=client&code_challenge=challenge&state=state"
	previous := shellExecute
	defer func() { shellExecute = previous }()

	shellExecute = func(hwnd windows.Handle, verb, file, args, cwd *uint16, showCmd int32) error {
		if got := windows.UTF16PtrToString(file); got != authorizationURL {
			t.Errorf("opened URL = %q, want %q", got, authorizationURL)
		}
		if verb != nil || args != nil || cwd != nil {
			t.Error("unexpected ShellExecute arguments")
		}
		return nil
	}

	if err := openBrowser(authorizationURL); err != nil {
		t.Fatalf("open browser: %v", err)
	}
}
