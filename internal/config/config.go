package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/physics91/naverworks-cli/internal/fileutil"
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

// AllKeys is the ordered list of all valid config keys.
var AllKeys = []string{
	"client_id", "client_secret", "service_account_id", "private_key_path",
	"domain_id", "bot_id", "scope", "default_calendar_user_id", "scim_access_token",
}

var validKeys = func() map[string]bool {
	m := make(map[string]bool, len(AllKeys))
	for _, k := range AllKeys {
		m[k] = true
	}
	return m
}()

var sensitiveKeys = map[string]bool{
	"client_secret":     true,
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

func EnvVarForKey(key string) string {
	return envMap[key]
}

func IsSensitiveKey(key string) bool {
	return sensitiveKeys[key]
}

func IsValidKey(key string) bool {
	return validKeys[key]
}

func DefaultPath() string {
	path, _ := DefaultPathOrError()
	return path
}

func DefaultPathOrError() (string, error) {
	return defaultPathOrError(os.UserConfigDir)
}

func defaultPathOrError(configDirFn func() (string, error)) (string, error) {
	configDir, err := configDirFn()
	if err != nil {
		return "", fmt.Errorf("설정 디렉토리 조회 실패: %w", err)
	}
	return configPathFromDir(configDir)
}

func configPathFromDir(configDir string) (string, error) {
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return "", fmt.Errorf("설정 디렉토리가 비어 있습니다")
	}
	if !filepath.IsAbs(configDir) {
		return "", fmt.Errorf("설정 디렉토리가 절대 경로가 아닙니다: %s", configDir)
	}
	return filepath.Join(configDir, "naverworks", "config.json"), nil
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
	return fileutil.WriteSecureJSON(path, c)
}

func (c *Config) fieldPtr(key string) *string {
	switch key {
	case "client_id":
		return &c.ClientID
	case "client_secret":
		return &c.ClientSecret
	case "service_account_id":
		return &c.ServiceAccountID
	case "private_key_path":
		return &c.PrivateKeyPath
	case "domain_id":
		return &c.DomainID
	case "bot_id":
		return &c.BotID
	case "scope":
		return &c.Scope
	case "default_calendar_user_id":
		return &c.DefaultCalendarUserID
	case "scim_access_token":
		return &c.ScimAccessToken
	}
	return nil
}

func (c *Config) Set(key, value string) error {
	if !validKeys[key] {
		return fmt.Errorf("유효하지 않은 설정 키: %s", key)
	}
	if ptr := c.fieldPtr(key); ptr != nil {
		*ptr = value
	}
	return nil
}

func (c *Config) Get(key string) (string, error) {
	if !validKeys[key] {
		return "", fmt.Errorf("유효하지 않은 설정 키: %s", key)
	}
	if ptr := c.fieldPtr(key); ptr != nil {
		return *ptr, nil
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

type ProfileConfig struct {
	CurrentProfile string             `json:"current_profile"`
	Profiles       map[string]*Config `json:"profiles"`
}

func NewProfileConfig() *ProfileConfig {
	return &ProfileConfig{
		CurrentProfile: "default",
		Profiles:       make(map[string]*Config),
	}
}

func (pc *ProfileConfig) EnsureProfile(name string) *Config {
	if pc.Profiles == nil {
		pc.Profiles = make(map[string]*Config)
	}
	if _, ok := pc.Profiles[name]; !ok {
		pc.Profiles[name] = &Config{}
	}
	return pc.Profiles[name]
}

func (pc *ProfileConfig) SetCurrentProfile(name string) {
	pc.CurrentProfile = name
}

// ActiveProfile returns the active profile by priority:
// flagProfile > NW_PROFILE env > current_profile > "default"
func (pc *ProfileConfig) ActiveProfile(flagProfile string) (*Config, string, error) {
	name := pc.CurrentProfile
	if name == "" {
		name = "default"
	}

	if envProfile := os.Getenv("NW_PROFILE"); envProfile != "" {
		name = envProfile
	}
	if flagProfile != "" {
		name = flagProfile
	}

	profile, ok := pc.Profiles[name]
	if !ok {
		return nil, name, fmt.Errorf("프로필 '%s'을(를) 찾을 수 없습니다", name)
	}
	return profile, name, nil
}

func LoadProfileConfig(path string) (*ProfileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			pc := NewProfileConfig()
			pc.EnsureProfile("default")
			return pc, nil
		}
		return nil, fmt.Errorf("config 읽기 실패: %w", err)
	}

	// Try profile format first
	pc := &ProfileConfig{}
	if err := json.Unmarshal(data, pc); err == nil && pc.Profiles != nil && len(pc.Profiles) > 0 {
		return pc, nil
	}

	// Legacy format → migrate to default profile
	legacy := &Config{}
	if err := json.Unmarshal(data, legacy); err != nil {
		return nil, fmt.Errorf("config 파싱 실패: %w", err)
	}
	pc = NewProfileConfig()
	pc.Profiles["default"] = legacy
	return pc, nil
}

func (pc *ProfileConfig) Save(path string) error {
	return fileutil.WriteSecureJSON(path, pc)
}
