package fileutil

import "testing"

// `icacls /grant:r` replaces only the named principal's entries, so an explicit
// ACE for another principal survives an exclusive grant. These are the entries
// that must be removed for the grant to actually be exclusive.
// The identity is the reference an ACL is compared against, so a broad
// principal must never be accepted — the fallback path reads the
// attacker-controllable USERNAME variable.
func TestIsUsableWindowsIdentity(t *testing.T) {
	usable := []string{`DESKTOP-1\bob`, `CORP\alice`}
	for _, user := range usable {
		if !IsUsableWindowsIdentity(user) {
			t.Errorf("%q should be usable as an identity", user)
		}
	}

	rejected := []string{
		"", "   ",
		"Everyone", "everyone", `BUILTIN\Users`, "Users",
		"Authenticated Users", "Guest",
		// A bare account name cannot be matched against a real MACHINE\user ACE.
		"bob", `bob\`, `\bob`,
		// Administrative names must not stand in as the identity.
		`CORP\Administrators`, `CORP\Administrator`, `CORP\SYSTEM`,
		`NT AUTHORITY\SYSTEM`, `BUILTIN\Administrators`,
		`CORP\Everyone`,
	}
	for _, user := range rejected {
		if IsUsableWindowsIdentity(user) {
			t.Errorf("%q must be rejected as an identity", user)
		}
	}
}

func TestForeignWindowsACLPrincipals(t *testing.T) {
	const path = `C:\Users\bob\token.json`
	user := `DESKTOP-1\bob`

	tests := []struct {
		name string
		out  string
		want []string
	}{
		{
			name: "clean acl yields nothing to remove",
			out: path + ` NT AUTHORITY\SYSTEM:(F)` + "\n" +
				`    BUILTIN\Administrators:(F)` + "\n" +
				`    DESKTOP-1\bob:(F)` + "\n",
			want: nil,
		},
		{
			name: "explicit everyone and users are reported",
			out: path + ` Everyone:(F)` + "\n" +
				`    BUILTIN\Users:(RX)` + "\n" +
				`    DESKTOP-1\bob:(F)` + "\n",
			want: []string{"Everyone", `BUILTIN\Users`},
		},
		{
			name: "duplicates are collapsed case-insensitively",
			out: path + ` Everyone:(F)` + "\n" +
				`    EVERYONE:(RX)` + "\n",
			want: []string{"Everyone"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ForeignWindowsACLPrincipals(tc.out, path, user)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("index %d: got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
