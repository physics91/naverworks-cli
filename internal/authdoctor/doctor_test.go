package authdoctor

import (
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
)

func TestRunLocalReportsMissingCredentialsAndToken(t *testing.T) {
	result := RunLocal(Input{
		SelectedProfile:       "default",
		SelectedProfileSource: "config",
		EffectiveConfig:       &config.Config{},
		SourceMap:             map[string]string{},
		InferredAuthMethod:    auth.AuthMethodOAuth,
		EffectiveScope:        "openid profile bot directory calendar",
		EffectiveScopeSource:  "default",
		ScimBaseURL:           "https://www.worksapis.com/scim/v2",
	})

	if result.ContractVersion != ContractVersion {
		t.Fatalf("contract version = %q, want %q", result.ContractVersion, ContractVersion)
	}
	if result.UncheckedCount != 2 {
		t.Fatalf("unchecked_count = %d, want 2", result.UncheckedCount)
	}

	checks := indexChecks(result.Checks)
	if got := checks["config.client_credentials"].Status; got != StatusFail {
		t.Fatalf("config.client_credentials status = %q, want %q", got, StatusFail)
	}
	if got := checks["token.present"].Status; got != StatusFail {
		t.Fatalf("token.present status = %q, want %q", got, StatusFail)
	}
	if got := checks["scope.core_configured"].Status; got != StatusPass {
		t.Fatalf("scope.core_configured status = %q, want %q", got, StatusPass)
	}
	if got := checks["scim.token_present"].Status; got != StatusUnknown {
		t.Fatalf("scim.token_present status = %q, want %q", got, StatusUnknown)
	}
}

func TestBuildCheckSanitizesEvidenceDefaultDeny(t *testing.T) {
	check := buildCheck("config.client_credentials", StatusFail, VerificationLocalStatic, "msg", map[string]any{
		"client_id_source":     "env",
		"client_secret":        "secret",
		"client_secret_source": "env",
		"unexpected_key":       "drop-me",
	})

	if _, ok := check.Evidence["unexpected_key"]; ok {
		t.Fatalf("unexpected_key should be removed: %#v", check.Evidence)
	}
	if _, ok := check.Evidence["client_secret"]; ok {
		t.Fatalf("client_secret should be removed by default-deny: %#v", check.Evidence)
	}
	if got := check.Evidence["client_secret_source"]; got != "env" {
		t.Fatalf("client_secret_source = %#v, want %q", got, "env")
	}
}

func TestApplyRemoteVerificationAppendsProbeChecks(t *testing.T) {
	result := RunLocal(Input{
		SelectedProfile:       "default",
		SelectedProfileSource: "config",
		EffectiveConfig: &config.Config{
			ScimAccessToken: "scim-token",
		},
		SourceMap:            map[string]string{"scim_access_token": "env"},
		InferredAuthMethod:   auth.AuthMethodOAuth,
		EffectiveScope:       "openid profile directory",
		EffectiveScopeSource: "env",
		Token: &auth.Token{
			AuthMethod:  auth.AuthMethodOAuth,
			AccessToken: "access-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
		ScimBaseURL: "https://www.worksapis.com/scim/v2",
	})

	ApplyRemoteVerification(&result, Input{
		Token: &auth.Token{
			AuthMethod:  auth.AuthMethodOAuth,
			AccessToken: "access-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
		EffectiveConfig: &config.Config{ScimAccessToken: "scim-token"},
	}, RemoteTransports{
		Directory: staticTransport{status: 200, body: []byte(`{}`)},
		SCIM:      staticTransport{status: 200, body: []byte(`{}`)},
	})

	if result.VerificationScope != VerificationScopeRemoteOptIn {
		t.Fatalf("verification_scope = %q, want %q", result.VerificationScope, VerificationScopeRemoteOptIn)
	}
	if result.UncheckedCount != 0 {
		t.Fatalf("unchecked_count = %d, want 0", result.UncheckedCount)
	}
	checks := indexChecks(result.Checks)
	if got := checks["scope.directory_remote_probe"].Status; got != StatusPass {
		t.Fatalf("scope.directory_remote_probe status = %q, want %q", got, StatusPass)
	}
	if got := checks["scim.endpoint_reachable"].Status; got != StatusPass {
		t.Fatalf("scim.endpoint_reachable status = %q, want %q", got, StatusPass)
	}
}

func TestEvaluateRemoteProbeClassifiesNetworkFailureAsUnknown(t *testing.T) {
	check := evaluateRemoteProbe("scope.directory_remote_probe", staticTransport{err: assertErr("boom")}, "/users", map[string]string{"count": "1"}, "ok", "fail")
	if check.Status != StatusUnknown {
		t.Fatalf("status = %q, want %q", check.Status, StatusUnknown)
	}
	if !strings.Contains(check.Message, "boom") {
		t.Fatalf("message = %q, want boom", check.Message)
	}
}

type staticTransport struct {
	body   []byte
	status int
	err    error
}

func (s staticTransport) Get(path string, query map[string]string) ([]byte, int, error) {
	return s.body, s.status, s.err
}

type assertErr string

func (e assertErr) Error() string { return string(e) }

func indexChecks(checks []Check) map[string]Check {
	out := make(map[string]Check, len(checks))
	for _, check := range checks {
		out[check.CheckID] = check
	}
	return out
}
