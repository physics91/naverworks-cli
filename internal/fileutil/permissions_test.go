//go:build !windows

package fileutil

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func TestEnsureSecureFilePermissions(t *testing.T) {
	tests := []struct {
		name      string
		perm      os.FileMode
		wantWarn  bool
		wantFinal os.FileMode
	}{
		{name: "world readable is tightened", perm: 0644, wantWarn: true, wantFinal: 0600},
		{name: "group readable is tightened", perm: 0640, wantWarn: true, wantFinal: 0600},
		{name: "world writable is tightened", perm: 0666, wantWarn: true, wantFinal: 0600},
		{name: "already private is untouched", perm: 0600, wantWarn: false, wantFinal: 0600},
		{name: "owner read only is untouched", perm: 0400, wantWarn: false, wantFinal: 0400},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "secret.json")
			if err := os.WriteFile(path, []byte(`{}`), tc.perm); err != nil {
				t.Fatal(err)
			}
			// os.WriteFile is subject to umask; force the intended mode.
			if err := os.Chmod(path, tc.perm); err != nil {
				t.Fatal(err)
			}

			var warnings strings.Builder
			original := SecureFileWarnWriter
			SecureFileWarnWriter = &warnings
			defer func() { SecureFileWarnWriter = original }()

			if err := EnsureSecureFilePermissions(path); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != tc.wantFinal {
				t.Errorf("expected final mode %04o, got %04o", tc.wantFinal, got)
			}
			if warned := warnings.Len() > 0; warned != tc.wantWarn {
				t.Errorf("expected warning=%v, got %q", tc.wantWarn, warnings.String())
			}
		})
	}
}

func TestEnsureSecureFilePermissions_MissingFileIsNotAnError(t *testing.T) {
	if err := EnsureSecureFilePermissions(filepath.Join(t.TempDir(), "absent.json")); err != nil {
		t.Fatalf("expected missing file to be tolerated, got %v", err)
	}
}

// A FIFO would block a later os.ReadFile indefinitely, so a non-regular path is
// rejected rather than waved through.
func TestEnsureSecureFilePermissions_RejectsNonRegularFile(t *testing.T) {
	dir := t.TempDir()

	if err := EnsureSecureFilePermissions(dir); err == nil {
		t.Error("expected a directory to be rejected")
	}

	fifo := filepath.Join(dir, "creds.fifo")
	if err := syscall.Mkfifo(fifo, 0600); err != nil {
		t.Skipf("mkfifo unsupported: %v", err)
	}
	if err := EnsureSecureFilePermissions(fifo); err == nil {
		t.Error("expected a FIFO to be rejected")
	}
}

// A symlinked config (dotfiles-style setup) must have its target tightened, not
// be silently accepted.
func TestEnsureSecureFilePermissions_FollowsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real.json")
	if err := os.WriteFile(target, []byte(`{}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(target, 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.json")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}

	var warnings strings.Builder
	original := SecureFileWarnWriter
	SecureFileWarnWriter = &warnings
	defer func() { SecureFileWarnWriter = original }()

	if err := EnsureSecureFilePermissions(link); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected symlink target tightened to 0600, got %04o", perm)
	}
}
