//go:build windows

package fileutil

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func hardenJSONDirPermissions(path string) error {
	return applyWindowsACL(path, true)
}

func hardenJSONFilePermissions(path string) error {
	return applyWindowsACL(path, false)
}

func applyWindowsACL(path string, isDir bool) error {
	user, err := currentWindowsACLUser()
	if err != nil {
		return err
	}

	grant := user + ":F"
	if isDir {
		grant = user + ":(OI)(CI)F"
	}

	out, err := exec.Command("icacls", path, "/inheritance:r", "/grant:r", grant).CombinedOutput()
	if err != nil {
		return fmt.Errorf("icacls 실패: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func currentWindowsACLUser() (string, error) {
	user := strings.TrimSpace(os.Getenv("USERNAME"))
	if user == "" {
		out, err := exec.Command("whoami").Output()
		if err != nil {
			return "", fmt.Errorf("현재 Windows 사용자 확인 실패: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	domain := strings.TrimSpace(os.Getenv("USERDOMAIN"))
	if domain == "" {
		return user, nil
	}
	return domain + `\` + user, nil
}
