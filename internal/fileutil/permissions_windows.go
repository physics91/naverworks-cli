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

// checkPlatformFilePermissions inspects the file ACL and re-applies an exclusive
// grant when a principal other than the current user (or a built-in
// administrative principal) has access.
//
// Unix mode bits are meaningless on Windows, so without this the credential
// permission check would be a no-op there and a config or token file inherited
// with `Everyone` access would be read as-is.
func checkPlatformFilePermissions(path string) error {
	user, err := currentWindowsACLUser()
	if err != nil {
		return err
	}

	issue, err := inspectWindowsACL(path, user)
	if err != nil {
		return err
	}
	if issue == "" {
		return nil
	}

	// applyWindowsACL re-verifies the resulting ACL and fails if a broad grant
	// survives, so reaching this point means the repair took effect.
	if err := applyWindowsACL(path, false); err != nil {
		return fmt.Errorf("%s: ACL을 현재 사용자 전용으로 변경하지 못했습니다: %w", issue, err)
	}

	fmt.Fprintf(
		SecureFileWarnWriter,
		"경고: %s 현재 사용자 전용으로 ACL을 변경했습니다. 자격 증명이 노출되었을 수 있으니 재발급을 검토하세요\n",
		issue,
	)
	return nil
}

func inspectWindowsACL(path, user string) (string, error) {
	out, err := exec.Command("icacls", path).Output()
	if err != nil {
		return "", fmt.Errorf("파일 ACL 확인 실패: %w", err)
	}
	return EvaluateWindowsACL(string(out), path, user), nil
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

	// /inheritance:r drops inherited entries and /grant:r replaces only the
	// named principal's entries; an explicit ACE for any other principal
	// survives. Remove those so the grant is genuinely exclusive.
	if err := removeForeignWindowsACEs(path, user); err != nil {
		return err
	}

	// Confirm the resulting ACL is actually exclusive. Reporting failure is the
	// safe outcome: callers either refuse to write or refuse to read.
	remaining, err := inspectWindowsACL(path, user)
	if err != nil {
		return err
	}
	if remaining != "" {
		return fmt.Errorf("ACL 적용 후에도 광범위한 접근 권한이 남아 있습니다: %s", remaining)
	}
	return nil
}

func removeForeignWindowsACEs(path, user string) error {
	out, err := exec.Command("icacls", path).Output()
	if err != nil {
		return fmt.Errorf("파일 ACL 확인 실패: %w", err)
	}

	for _, principal := range ForeignWindowsACLPrincipals(string(out), path, user) {
		cmdOut, err := exec.Command(
			"icacls", path, "/remove:g", principal, "/remove:d", principal,
		).CombinedOutput()
		if err != nil {
			return fmt.Errorf(
				"ACL 항목(%s) 제거 실패: %w: %s",
				principal, err, strings.TrimSpace(string(cmdOut)),
			)
		}
	}
	return nil
}

// currentWindowsACLUser resolves the account that should own the file ACL,
// preferring `whoami` over USERNAME/USERDOMAIN. Environment variables are
// attacker-controllable, and trusting them first would let a caller redirect
// the exclusive grant to an over-broad principal (e.g. USERNAME=Everyone).
// The resolved name is validated for the same reason, since the fallback path
// still reads the environment.
func currentWindowsACLUser() (string, error) {
	if out, err := exec.Command("whoami").Output(); err == nil {
		user := strings.TrimSpace(string(out))
		if IsUsableWindowsIdentity(user) {
			return user, nil
		}
	}

	user := strings.TrimSpace(os.Getenv("USERNAME"))
	if user == "" {
		return "", fmt.Errorf("현재 Windows 사용자 확인 실패")
	}

	if domain := strings.TrimSpace(os.Getenv("USERDOMAIN")); domain != "" {
		user = domain + `\` + user
	}
	if !IsUsableWindowsIdentity(user) {
		return "", fmt.Errorf("현재 Windows 사용자로 사용할 수 없는 이름입니다: %s", user)
	}
	return user, nil
}
