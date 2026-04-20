package fileutil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteSecureJSON serializes v as indented JSON and writes it to path with
// secure permissions (0700 directory, 0600 file).
// On non-Windows, it uses atomic write (temp file + sync + rename) to prevent
// file corruption on process crash. On Windows, it falls back to os.WriteFile
// because os.Rename cannot replace an existing file atomically.
func WriteSecureJSON(path string, v interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	if err := hardenJSONDirPermissions(dir); err != nil {
		return fmt.Errorf("디렉토리 권한 설정 실패: %w", err)
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("직렬화 실패: %w", err)
	}

	if filepath.Separator == '\\' {
		if err := os.WriteFile(path, data, 0600); err != nil {
			return err
		}
		if err := hardenJSONFilePermissions(path); err != nil {
			return fmt.Errorf("파일 권한 설정 실패: %w", err)
		}
		return nil
	}

	// Atomic write: temp file in same directory → sync → rename
	tmp, err := os.CreateTemp(dir, ".naverworks-*.tmp")
	if err != nil {
		return fmt.Errorf("임시 파일 생성 실패: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // 실패 시 정리

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("임시 파일 쓰기 실패: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("임시 파일 동기화 실패: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("임시 파일 닫기 실패: %w", err)
	}
	if err := hardenJSONFilePermissions(tmpPath); err != nil {
		return fmt.Errorf("파일 권한 설정 실패: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("원자적 파일 교체 실패: %w", err)
	}
	if err := hardenJSONFilePermissions(path); err != nil {
		return fmt.Errorf("파일 권한 설정 실패: %w", err)
	}
	return nil
}
