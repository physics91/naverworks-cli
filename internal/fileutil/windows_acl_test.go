package fileutil

import "testing"

func TestEvaluateWindowsACL(t *testing.T) {
	const path = `C:\Users\bob\key.pem`
	user := `DESKTOP-1\bob`

	tests := []struct {
		name      string
		out       string
		wantIssue bool
	}{
		{
			name: "only current user and built-in admins",
			out: path + ` NT AUTHORITY\SYSTEM:(F)` + "\n" +
				`                    BUILTIN\Administrators:(F)` + "\n" +
				`                    DESKTOP-1\bob:(F)` + "\n\n" +
				"Successfully processed 1 files; Failed processing 0 files\n",
			wantIssue: false,
		},
		{
			// The previous substring check passed this ACE because
			// "builtin\users" contains the user name "bob"? It does not — but
			// it did pass any group whose name embeds the account name. Exact
			// matching rejects broad groups unconditionally.
			name:      "builtin users group is rejected",
			out:       path + ` BUILTIN\Users:(RX)` + "\n" + `                    DESKTOP-1\bob:(F)` + "\n",
			wantIssue: true,
		},
		{
			name:      "everyone is rejected",
			out:       path + ` Everyone:(F)` + "\n" + `                    DESKTOP-1\bob:(F)` + "\n",
			wantIssue: true,
		},
		{
			// Regression for the substring bypass: an account name that
			// contains the current user name must not be accepted.
			name:      "different account containing the user name is rejected",
			out:       path + ` DESKTOP-1\bobby:(F)` + "\n",
			wantIssue: true,
		},
		{
			name:      "no parsable ACE fails closed",
			out:       "Successfully processed 1 files; Failed processing 0 files\n",
			wantIssue: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issue := EvaluateWindowsACL(tc.out, path, user)
			if tc.wantIssue && issue == "" {
				t.Fatalf("expected an ACL issue, got none")
			}
			if !tc.wantIssue && issue != "" {
				t.Fatalf("expected no ACL issue, got %q", issue)
			}
		})
	}
}

func TestEvaluateWindowsACL_RejectsBareAndCrossDomainIdentities(t *testing.T) {
	const path = `C:\key.pem`

	// A bare account name must not match a domain-qualified ACE: accepting it
	// allowed a principal from another domain to pass.
	if issue := EvaluateWindowsACL(path+` DESKTOP-1\bob:(F)`+"\n", path, "bob"); issue == "" {
		t.Error("bare identity must not match a domain-qualified ACE")
	}

	// A same-account principal from a different domain must be rejected.
	if issue := EvaluateWindowsACL(path+` OTHERDOMAIN\bob:(F)`+"\n", path, `DESKTOP-1\bob`); issue == "" {
		t.Error("cross-domain principal must be rejected")
	}

	// The exact qualified identity is accepted.
	if issue := EvaluateWindowsACL(path+` DESKTOP-1\bob:(F)`+"\n", path, `DESKTOP-1\bob`); issue != "" {
		t.Errorf("exact identity should be accepted, got %q", issue)
	}
}
