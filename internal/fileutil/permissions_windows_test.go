//go:build windows

package fileutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These tests exercise the real icacls paths. They only run on Windows, so
// without a Windows CI job the ACL inspection, foreign-ACE removal, and
// post-repair verification logic would ship unverified.

func TestWriteSecureJSON_ProducesExclusiveACL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.json")
	if err := WriteSecureJSON(path, map[string]string{"secret": "value"}); err != nil {
		t.Fatalf("WriteSecureJSON failed: %v", err)
	}
	if err := EnsureSecureFilePermissions(path); err != nil {
		t.Fatalf("freshly written file should pass the ACL check: %v", err)
	}
}

// An explicit broad ACE is exactly what `icacls /grant:r` alone does not remove,
// so this covers the removal-and-reverify path.
func TestEnsureSecureFilePermissions_RemovesExplicitEveryoneACE(t *testing.T) {
	path := filepath.Join(t.TempDir(), "creds.json")
	if err := WriteSecureJSON(path, map[string]string{"secret": "value"}); err != nil {
		t.Fatalf("WriteSecureJSON failed: %v", err)
	}

	if out, err := exec.Command("icacls", path, "/grant", "Everyone:F").CombinedOutput(); err != nil {
		t.Skipf("cannot grant Everyone on this runner: %v: %s", err, out)
	}

	user, err := currentWindowsACLUser()
	if err != nil {
		t.Fatalf("identity resolution failed: %v", err)
	}

	// Confirm the broad ACE is genuinely present before asserting on the repair.
	before, err := inspectWindowsACL(path, user)
	if err != nil {
		t.Fatalf("ACL inspection failed: %v", err)
	}
	if before == "" {
		t.Fatal("expected the injected Everyone ACE to be reported")
	}

	var warnings strings.Builder
	original := SecureFileWarnWriter
	SecureFileWarnWriter = &warnings
	defer func() { SecureFileWarnWriter = original }()

	if err := EnsureSecureFilePermissions(path); err != nil {
		t.Fatalf("repair failed: %v", err)
	}
	if warnings.Len() == 0 {
		t.Error("expected a warning about the repaired ACL")
	}

	after, err := inspectWindowsACL(path, user)
	if err != nil {
		t.Fatalf("post-repair inspection failed: %v", err)
	}
	if after != "" {
		t.Errorf("broad access survived repair: %s", after)
	}

	out, err := exec.Command("icacls", path).Output()
	if err != nil {
		t.Fatalf("icacls failed: %v", err)
	}
	// Check parsed principals, not the raw output: the temp directory name is
	// derived from this test's name and therefore contains "Everyone", which
	// makes a substring search on the output always match.
	if foreign := ForeignWindowsACLPrincipals(string(out), path, user); len(foreign) > 0 {
		t.Errorf("foreign ACEs survived repair: %v\nicacls output: %s", foreign, out)
	}

	// The repair must not lock the owner out of its own credentials.
	if _, err := os.ReadFile(path); err != nil {
		t.Errorf("owner lost read access after repair: %v", err)
	}
}

func TestEnsureSecureFilePermissions_RejectsDirectoryOnWindows(t *testing.T) {
	if err := EnsureSecureFilePermissions(t.TempDir()); err == nil {
		t.Error("expected a directory to be rejected")
	}
}

func TestCurrentWindowsACLUser_ReturnsQualifiedIdentity(t *testing.T) {
	user, err := currentWindowsACLUser()
	if err != nil {
		t.Fatalf("identity resolution failed: %v", err)
	}
	if !IsUsableWindowsIdentity(user) {
		t.Errorf("resolved identity %q is not usable for ACL comparison", user)
	}
}
