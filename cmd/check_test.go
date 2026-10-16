package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MoomooOpen/moomoo-cli/internal/auth"
)

func TestProbeEndpointAcceptsAnyHTTPResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	result := probeEndpoint(context.Background(), diagnosticHTTPClient(), server.URL)
	if !result.OK || result.URL != server.URL {
		t.Fatalf("result = %#v", result)
	}
}

func TestTokenStatusFromPermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions auth.PermissionInfo
		want        string
	}{
		{name: "quote only", permissions: auth.PermissionInfo{Quote: true}, want: tokenValid},
		{name: "trade only", permissions: auth.PermissionInfo{Trade: true}, want: tokenValid},
		{name: "quote and trade", permissions: auth.PermissionInfo{Quote: true, Trade: true}, want: tokenValid},
		{name: "no permission", permissions: auth.PermissionInfo{}, want: tokenUnknown},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := tokenStatusFromPermissions(test.permissions).Token; got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}
}

func TestCheckResultErrorFailsWhenTokenIsNotValid(t *testing.T) {
	result := checkResult{Session: sessionCheck{Token: tokenMissing}}
	if err := checkResultError(result); err == nil {
		t.Fatal("expected missing token error")
	}
}

func TestCheckResultErrorPassesForValidToken(t *testing.T) {
	result := checkResult{Session: sessionCheck{Token: tokenValid}}
	if err := checkResultError(result); err != nil {
		t.Fatalf("checkResultError() error = %v", err)
	}
}
