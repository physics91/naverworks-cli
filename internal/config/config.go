package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	ClientID              string `json:"client_id,omitempty"`
	ClientSecret          string `json:"client_secret,omitempty"`
	ServiceAccountID      string `json:"service_account_id,omitempty"`
	PrivateKeyPath        string `json:"private_key_path,omitempty"`
	DomainID              string `json:"domain_id,omitempty"`
	BotID                 string `json:"bot_id,omitempty"`
	Scope                 string `json:"scope,omitempty"`
	DefaultCalendarUserID string `json:"default_calendar_user_id,omitempty"`
	ScimAccessToken       string `json:"scim_access_token,omitempty"`
}

var validKeys = map[string]bool{
	"client_id": true, "client_secret": true, "service_account_id": true,
	"private_key_path": true, "domain_id": true, "bot_id": true,
	"scope": true, "default_calendar_user_id": true, "scim_access_token": true,
}

var sensitiveKeys = map[string]bool{
	"client_secret":    true,
	"scim_access_token": true,
}

var envMap = map[string]string{
	"client_id":                "NW_CLIENT_ID",
	"client_secret":            "NW_CLIENT_SECRET",
	"service_account_id":       "NW_SERVICE_ACCOUNT_ID",
	"private_key_path":         "NW_PRIVATE_KEY_PATH",
	"domain_id":                "NW_DOMAIN_ID",
	"bot_id":                   "NW_BOT_ID",
	"scope":                    "NW_SCOPE",
	"default_calendar_user_id": "NW_DEFAULT_CALENDAR_USER_ID",
	"scim_access_token":        "NW_SCIM_ACCESS_TOKEN",
}

func IsValidKey(key string) bool {
	return validKeys[key]
}

func DefaultPath() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "naverworks", "config.json")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "naverworks", "config.json")
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("config 읽기 실패: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config 파싱 실패: %w", err)
	}
	return cfg, nil
}

func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("디렉토리 생성 실패: %w", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dir, 0700); err != nil {
			return fmt.Errorf("디렉토리 권한 설정 실패: %w", err)
		}
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("config 직렬화 실패: %w", err)
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

func (c *Config) Set(key, value string) error {
	if !validKeys[key] {
		return fmt.Errorf("유효하지 않은 설정 키: %s", key)
	}
	switch key {
	case "client_id":
		c.ClientID = value
	case "client_secret":
		c.ClientSecret = value
	case "service_account_id":
		c.ServiceAccountID = value
	case "private_key_path":
		c.PrivateKeyPath = value
	case "domain_id":
		c.DomainID = value
	case "bot_id":
		c.BotID = value
	case "scope":
		c.Scope = value
	case "default_calendar_user_id":
		c.DefaultCalendarUserID = value
	case "scim_access_token":
		c.ScimAccessToken = value
	}
	return nil
}

func (c *Config) Get(key string) (string, error) {
	if !validKeys[key] {
		return "", fmt.Errorf("유효하지 않은 설정 키: %s", key)
	}
	switch key {
	case "client_id":
		return c.ClientID, nil
	case "client_secret":
		return c.ClientSecret, nil
	case "service_account_id":
		return c.ServiceAccountID, nil
	case "private_key_path":
		return c.PrivateKeyPath, nil
	case "domain_id":
		return c.DomainID, nil
	case "bot_id":
		return c.BotID, nil
	case "scope":
		return c.Scope, nil
	case "default_calendar_user_id":
		return c.DefaultCalendarUserID, nil
	case "scim_access_token":
		return c.ScimAccessToken, nil
	}
	return "", nil
}

func (c *Config) GetMasked(key string) string {
	val, _ := c.Get(key)
	if sensitiveKeys[key] && val != "" {
		return "****"
	}
	return val
}

func (c *Config) ApplyEnvOverrides() {
	for key, envVar := range envMap {
		if val := os.Getenv(envVar); val != "" {
			c.Set(key, val)
		}
	}
}
