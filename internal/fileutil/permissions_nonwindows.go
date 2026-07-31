//go:build !windows

package fileutil

import (
	"fmt"
	"os"
)

func hardenJSONDirPermissions(path string) error {
	return os.Chmod(path, 0700)
}

func hardenJSONFilePermissions(path string) error {
	return os.Chmod(path, 0600)
}

// checkPlatformFilePermissions tightens a credential file whose mode allows
// group or other access.
func checkPlatformFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("파일 권한 확인 실패: %w", err)
	}

	perm := info.Mode().Perm()
	if perm&0077 == 0 {
		return nil
	}

	if err := os.Chmod(path, 0600); err != nil {
		return fmt.Errorf("%s 파일 권한이 %04o입니다. 0600으로 변경하지 못했습니다: %w", path, perm, err)
	}
	fmt.Fprintf(
		SecureFileWarnWriter,
		"경고: %s 파일 권한이 %04o였습니다. 0600으로 변경했습니다. 자격 증명이 노출되었을 수 있으니 재발급을 검토하세요\n",
		path, perm,
	)
	return nil
}
