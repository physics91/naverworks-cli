package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const refreshBuffer = 60 * time.Second

type Token struct {
	AuthMethod       string    `json:"auth_method"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token,omitempty"`
	TokenType        string    `json:"token_type"`
	ExpiresAt        time.Time `json:"expires_at"`
	Scope            string    `json:"scope"`
	ServiceAccountID string    `json:"service_account_id,omitempty"`
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

func (t *Token) NeedsRefresh() bool {
	return time.Now().After(t.ExpiresAt.Add(-refreshBuffer))
}

type TokenStore struct {
	path string
}

func DefaultTokenPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "naverworks", "token.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "naverworks", "token.json")
}

func NewTokenStore(path string) *TokenStore {
	return &TokenStore{path: path}
}

func (s *TokenStore) Load() (*Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("토큰 읽기 실패: %w", err)
	}
	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("토큰 파싱 실패: %w", err)
	}
	return &token, nil
}

func (s *TokenStore) Save(token *Token) error {
	return writeSecureJSON(s.path, token)
}

func (s *TokenStore) Delete() error {
	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("토큰 삭제 실패: %w", err)
	}
	return nil
}

type ProfileTokenStore struct {
	path    string
	profile string
}

type profileTokenFile struct {
	Tokens map[string]*Token `json:"tokens"`
}

func NewProfileTokenStore(path, profile string) *ProfileTokenStore {
	return &ProfileTokenStore{path: path, profile: profile}
}

func (s *ProfileTokenStore) Load() (*Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("토큰 읽기 실패: %w", err)
	}

	// Try profile format
	var pf profileTokenFile
	if err := json.Unmarshal(data, &pf); err == nil && pf.Tokens != nil {
		token, ok := pf.Tokens[s.profile]
		if !ok {
			return nil, nil
		}
		return token, nil
	}

	// Legacy format → default profile
	if s.profile == "default" {
		var token Token
		if err := json.Unmarshal(data, &token); err != nil {
			return nil, fmt.Errorf("토큰 파싱 실패: %w", err)
		}
		return &token, nil
	}

	return nil, nil
}

func (s *ProfileTokenStore) Save(token *Token) error {
	pf := &profileTokenFile{Tokens: make(map[string]*Token)}
	if data, err := os.ReadFile(s.path); err == nil {
		if err := json.Unmarshal(data, pf); err != nil || pf.Tokens == nil {
			var legacy Token
			if err := json.Unmarshal(data, &legacy); err == nil && legacy.AccessToken != "" {
				pf.Tokens = map[string]*Token{"default": &legacy}
			} else {
				pf.Tokens = make(map[string]*Token)
			}
		}
	}

	pf.Tokens[s.profile] = token
	return writeSecureJSON(s.path, pf)
}

// writeSecureJSON serializes v as indented JSON and writes it to path with
// secure permissions (0700 directory, 0600 file).
func writeSecureJSON(path string, v interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("디렉토리 권한 설정 실패: %w", err)
		}
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("토큰 직렬화 실패: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0600); err != nil {
			return fmt.Errorf("파일 권한 설정 실패: %w", err)
		}
	}
	return nil
}

func (s *ProfileTokenStore) Delete() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("토큰 읽기 실패: %w", err)
	}

	var pf profileTokenFile
	if err := json.Unmarshal(data, &pf); err != nil || pf.Tokens == nil {
		// Legacy → migrate to default first, then delete target
		var legacy Token
		if err := json.Unmarshal(data, &legacy); err == nil && legacy.AccessToken != "" {
			pf = profileTokenFile{Tokens: map[string]*Token{"default": &legacy}}
		} else {
			return os.Remove(s.path)
		}
	}

	delete(pf.Tokens, s.profile)

	if len(pf.Tokens) == 0 {
		return os.Remove(s.path)
	}

	return writeSecureJSON(s.path, pf)
}
