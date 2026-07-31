package fileutil

import (
	"fmt"
	"io"
	"os"
)

// SecureFileWarnWriter receives warnings emitted when insecure permissions are
// repaired. Tests may replace it to capture output.
var SecureFileWarnWriter io.Writer = os.Stderr

// EnsureSecureFilePermissions verifies that a file holding secrets (tokens,
// client secrets, access tokens) is reachable only by its owner, and tightens it
// when it is not.
//
// WriteSecureJSON enforces owner-only access at write time, but a file created
// by an older version, restored from a backup, copied with a permissive umask,
// or inherited from a parent directory's ACL can still be readable by others.
// Reading such a file silently would leave long-lived credentials exposed, so
// every load path calls this first.
//
// A missing file is not an error — callers treat absence as "not configured".
// A non-regular path (FIFO, device, socket) is rejected: a later os.ReadFile on
// a FIFO would block indefinitely, and no legitimate credential store is
// anything other than a regular file or a symlink to one.
//
// The enforcement mechanism is platform-specific: Unix mode bits on POSIX,
// ACL inspection via icacls on Windows.
func EnsureSecureFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s은(는) 일반 파일이 아닙니다", path)
	}
	return checkPlatformFilePermissions(path)
}
