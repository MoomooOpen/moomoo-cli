package cmd

import (
	"os"
	"testing"
	"time"

	"github.com/MoomooOpen/moomoo-cli/internal/auth"
	"github.com/MoomooOpen/moomoo-cli/internal/config"
	"github.com/zalando/go-keyring"
)

func TestEnsureClientIDReusesStoredRegistration(t *testing.T) {
	previousConfig := cfg
	cfg = &config.Config{
		APIBase:         config.DefaultAPIBase,
		ClientID:        "stored-client-id",
		ClientIDAPIBase: config.DefaultAPIBase,
		RedirectURI:     "http://127.0.0.1:18888/callback",
	}
	t.Cleanup(func() { cfg = previousConfig })

	clientID, err := ensureClientID([]string{cfg.RedirectURI})
	if err != nil {
		t.Fatalf("ensureClientID() error = %v", err)
	}
	if clientID != cfg.ClientID {
		t.Fatalf("client ID = %q, want %q", clientID, cfg.ClientID)
	}
}

func TestClientIDDoesNotMatchWithoutAPIBaseBinding(t *testing.T) {
	previousConfig := cfg
	cfg = &config.Config{
		APIBase:     config.DefaultAPIBase,
		ClientID:    "stored-client-id",
		RedirectURI: "http://127.0.0.1:18888/callback",
	}
	t.Cleanup(func() { cfg = previousConfig })

	if clientIDMatchesCurrentAPIBase() {
		t.Fatal("client ID without API base binding must not match")
	}
}

func TestClientIDDoesNotMatchDifferentAPIBase(t *testing.T) {
	previousConfig := cfg
	cfg = &config.Config{
		APIBase:         "https://webapi.moomoo.com",
		ClientID:        "stored-client-id",
		ClientIDAPIBase: "https://sandbox.moomoo.com",
		RedirectURI:     "http://127.0.0.1:18888/callback",
	}
	t.Cleanup(func() { cfg = previousConfig })

	if clientIDMatchesCurrentAPIBase() {
		t.Fatal("client ID bound to another API base must not match")
	}
}

func TestRegisterAndSaveClientIDStoresAPIBaseBinding(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	previousConfig := cfg
	previousRegister := registerOAuthClient
	cfg = &config.Config{
		APIBase:   "https://sandbox.moomoo.com/",
		TokenFile: t.TempDir() + "/token.json",
	}
	registerOAuthClient = func(baseURL, clientName string, redirectURIs []string) (*auth.RegistrationResponse, error) {
		if baseURL != "https://sandbox.moomoo.com" {
			t.Fatalf("baseURL = %q, want normalized sandbox base", baseURL)
		}
		return &auth.RegistrationResponse{ClientID: "new-client-id"}, nil
	}
	t.Cleanup(func() {
		cfg = previousConfig
		registerOAuthClient = previousRegister
	})

	clientID, err := registerAndSaveClientID([]string{"http://127.0.0.1:18888/callback"})
	if err != nil {
		t.Fatalf("registerAndSaveClientID() error = %v", err)
	}
	if clientID != "new-client-id" {
		t.Fatalf("clientID = %q, want new-client-id", clientID)
	}
	if cfg.ClientIDAPIBase != "https://sandbox.moomoo.com" {
		t.Fatalf("ClientIDAPIBase = %q, want sandbox base", cfg.ClientIDAPIBase)
	}
}

func TestNewAuthStatusValidToken(t *testing.T) {
	now := time.Now()
	store := &auth.TokenStore{
		LoggedInAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour),
		ClientName: "test-client",
		Account: auth.AccountInfo{
			FutuID:      "moomoo-1",
			MemberID:    "member-1",
			AccountNo:   "account-1",
			AccountID:   "account-id-1",
			AccountType: "margin",
		},
		QuotePermissions: []string{"HK", "US"},
	}
	status := newAuthStatus(store, "/tmp/token.enc")
	if status.Token.Status != statusLoggedIn {
		t.Fatalf("status = %q", status.Token.Status)
	}
	if status.Token.LoggedInAt == nil || status.Token.ExpiresAt == nil {
		t.Fatal("expected login and expiry timestamps")
	}
	if status.Account == nil {
		t.Fatalf("account = %#v", status.Account)
	}
	if status.Account.AccountID != "account-id-1" || status.Account.AccountType != "margin" {
		t.Fatalf("non-sensitive account fields were not preserved: %#v", status.Account)
	}
	if status.Account.FutuID != "" || status.Account.MemberID != "" || status.Account.AccountNo != "" {
		t.Fatalf("sensitive user IDs must not be exposed: %#v", status.Account)
	}
	if len(status.QuotePermissions) != 2 {
		t.Fatalf("quote permissions = %#v", status.QuotePermissions)
	}
}

func TestNewAuthStatusOmitsAccountWithOnlyUserIdentifiers(t *testing.T) {
	store := &auth.TokenStore{
		ExpiresAt: time.Now().Add(time.Hour),
		Account: auth.AccountInfo{
			FutuID:    "moomoo-1",
			MemberID:  "member-1",
			AccountNo: "account-1",
		},
	}
	status := newAuthStatus(store, "/tmp/token.enc")
	if status.Account != nil {
		t.Fatalf("account with only sensitive user IDs must be omitted: %#v", status.Account)
	}
}

func TestNewAuthStatusExpiredToken(t *testing.T) {
	store := &auth.TokenStore{ExpiresAt: time.Now().Add(-time.Minute)}
	status := newAuthStatus(store, "/tmp/token.enc")
	if status.Token.Status != statusExpired {
		t.Fatalf("status = %q", status.Token.Status)
	}
}

func TestNewAuthStatusWithoutToken(t *testing.T) {
	status := newAuthStatus(nil, "/tmp/token.enc")
	if status.Token.Status != statusNotLoggedIn {
		t.Fatalf("status = %q", status.Token.Status)
	}
	if status.Token.LoggedInAt != nil || status.Token.ExpiresAt != nil {
		t.Fatal("timestamps must be nil without a token")
	}
}

func TestRunLogoutDeletesLocalToken(t *testing.T) {
	keyring.MockInit()
	previousTokenFile := tokenFile
	t.Cleanup(func() {
		tokenFile = previousTokenFile
	})

	path := t.TempDir() + "/token.json"
	store := &auth.TokenStore{
		AccessToken: "access-token",
		ClientID:    "stored-client-id",
	}
	if err := store.Save(path); err != nil {
		t.Fatalf("save token: %v", err)
	}

	tokenFile = path
	if err := runLogout(nil, nil); err != nil {
		t.Fatalf("runLogout() error = %v", err)
	}
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Fatalf("local token still exists after revoke failure: %v", statErr)
	}
}
