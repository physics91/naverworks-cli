package web

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/physics91/naverworks-cli/internal/fileutil"
)

// SessionDir maps profile names to a dedicated browser directory. A profile
// name must never become a path segment because existing profiles can contain
// slashes or traversal components.
func SessionDir(configPath, profile string) (string, error) {
	if !filepath.IsAbs(configPath) || strings.TrimSpace(profile) == "" {
		return "", fmt.Errorf("브라우저 세션에는 절대 설정 경로와 프로필명이 필요합니다")
	}
	id := sha256.Sum256([]byte(profile))
	return filepath.Join(filepath.Dir(configPath), "browser", fmt.Sprintf("%x", id)), nil
}

func validateSessionPath(dir, profile string) error {
	id := sha256.Sum256([]byte(profile))
	if !filepath.IsAbs(dir) || strings.TrimSpace(profile) == "" ||
		filepath.Base(dir) != fmt.Sprintf("%x", id) || filepath.Base(filepath.Dir(dir)) != "browser" {
		return fmt.Errorf("유효하지 않은 전용 브라우저 세션 경로입니다")
	}
	return nil
}

func prepareSession(dir, profile string) error {
	if err := validateSessionPath(dir, profile); err != nil {
		return err
	}
	// WriteSecureJSON applies the existing owner-only Unix/Windows permission
	// policy to each directory before Chrome can persist any credentials.
	for _, target := range []string{filepath.Dir(dir), dir} {
		info, err := os.Lstat(target)
		if err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("브라우저 세션 경로 확인 실패: %w", err)
		}
		if err == nil && (!info.IsDir() || info.Mode()&os.ModeSymlink != 0) {
			return fmt.Errorf("브라우저 세션 경로는 심볼릭 링크가 아닌 디렉토리여야 합니다")
		}
		ownerPath := filepath.Join(target, "owner.json")
		ownerInfo, ownerErr := os.Lstat(ownerPath)
		if ownerErr != nil && !os.IsNotExist(ownerErr) {
			return ownerErr
		}
		if ownerErr == nil && !ownerInfo.Mode().IsRegular() {
			return fmt.Errorf("브라우저 세션 메타데이터는 일반 파일이어야 합니다")
		}
		owner := map[string]string{"kind": "browser-sessions"}
		if target == dir {
			owner["profile"] = profile
		}
		if err := fileutil.WriteSecureJSON(ownerPath, owner); err != nil {
			return fmt.Errorf("브라우저 세션 권한 설정 실패: %w", err)
		}
	}
	return nil
}

// Logout removes only this profile's local credentials. It does not revoke a
// server session or touch API tokens and does not require a browser executable.
func (b Browser) Logout() error {
	if err := validateSessionPath(b.SessionDir, b.Profile); err != nil {
		return err
	}
	for _, target := range []string{filepath.Dir(b.SessionDir), b.SessionDir} {
		info, err := os.Lstat(target)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("브라우저 세션 경로 확인 실패: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("브라우저 세션 경로는 심볼릭 링크가 아닌 디렉토리여야 합니다")
		}
	}
	for _, lock := range []string{"SingletonLock", "lockfile"} {
		if _, err := os.Lstat(filepath.Join(b.SessionDir, lock)); err == nil {
			return fmt.Errorf("브라우저 세션이 사용 중일 수 있습니다. 이 프로필의 브라우저와 CLI를 종료한 뒤 로그아웃하세요")
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	ownerPath := filepath.Join(b.SessionDir, "owner.json")
	info, err := os.Lstat(ownerPath)
	if err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("브라우저 세션 소유 정보를 확인하지 못했습니다")
	}
	data, err := os.ReadFile(ownerPath)
	if err != nil {
		return err
	}
	var owner struct{ Kind, Profile string }
	if json.Unmarshal(data, &owner) != nil || owner.Kind != "browser-sessions" || owner.Profile != b.Profile {
		return fmt.Errorf("브라우저 세션 소유 정보가 프로필과 일치하지 않습니다")
	}
	if err := os.RemoveAll(b.SessionDir); err != nil {
		return fmt.Errorf("브라우저 세션 삭제 실패: %w", err)
	}
	return nil
}
