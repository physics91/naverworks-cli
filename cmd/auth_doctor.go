package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/authdoctor"
	"github.com/physics91/naverworks-cli/internal/config"
	"github.com/physics91/naverworks-cli/internal/httputil"
	"github.com/physics91/naverworks-cli/internal/output"
	"github.com/spf13/cobra"
)

var authDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "인증 진단",
	RunE: func(cmd *cobra.Command, args []string) error {
		verifyRemote, _ := cmd.Flags().GetBool("verify-remote")

		input, err := buildAuthDoctorInput()
		if err != nil {
			return err
		}

		result := authdoctor.RunLocal(input)
		if verifyRemote {
			authdoctor.ApplyRemoteVerification(&result, input, buildDoctorRemoteTransports(input))
		}
		return printDoctorResult(result)
	},
}

type readonlyHTTPTransport struct {
	baseURL     string
	accessToken string
	httpClient  *http.Client
}

func init() {
	authDoctorCmd.Flags().Bool("verify-remote", false, "읽기 전용 원격 검증 실행")
	authCmd.AddCommand(authDoctorCmd)
}

func buildAuthDoctorInput() (authdoctor.Input, error) {
	configPath, err := config.DefaultPathOrError()
	if err != nil {
		return authdoctor.Input{}, err
	}
	pc, profileLoadErr := config.LoadProfileConfig(configPath)
	if profileLoadErr != nil {
		pc = config.NewProfileConfig()
		pc.EnsureProfile("default")
	}

	selectedProfile, _ := selectedProfileName(pc)
	selectedSource := selectedProfileSource()

	effectiveConfig, activeProfile, selectionErr := loadActiveConfig()
	if activeProfile != "" {
		selectedProfile = activeProfile
	}
	if effectiveConfig == nil {
		effectiveConfig = &config.Config{}
	}
	if profileLoadErr != nil && selectionErr == nil {
		selectionErr = profileLoadErr
	}

	tokenPath, err := auth.DefaultTokenPathOrError()
	if err != nil {
		return authdoctor.Input{}, err
	}
	store := auth.NewProfileTokenStore(tokenPath, selectedProfile)
	token, tokenErr := store.Load()

	sourceMap := buildDoctorSourceMap(pc, selectedProfile)
	inferredMethod := inferAuthMethod(effectiveConfig, token)
	effectiveScope, effectiveScopeSource := effectiveScopeForDoctor(effectiveConfig, token, sourceMap)
	callbackErr := callbackListenerCheck(inferredMethod)

	return authdoctor.Input{
		SelectedProfile:       selectedProfile,
		SelectedProfileSource: selectedSource,
		EffectiveConfig:       cloneConfig(effectiveConfig),
		SourceMap:             sourceMap,
		InferredAuthMethod:    inferredMethod,
		EffectiveScope:        effectiveScope,
		EffectiveScopeSource:  effectiveScopeSource,
		ScimBaseURL:           scimBaseURL,
		Token:                 token,
		SelectionErr:          selectionErr,
		TokenErr:              tokenErr,
		CallbackListenerErr:   callbackErr,
	}, nil
}

func selectedProfileSource() string {
	if strings.TrimSpace(profileName) != "" {
		return "flag"
	}
	if strings.TrimSpace(os.Getenv("NW_PROFILE")) != "" {
		return "env"
	}
	return "config"
}

func inferAuthMethod(cfg *config.Config, token *auth.Token) auth.AuthMethod {
	if token != nil && token.AuthMethod != "" {
		return token.AuthMethod
	}
	if cfg != nil && (strings.TrimSpace(cfg.ServiceAccountID) != "" || strings.TrimSpace(cfg.PrivateKeyPath) != "") {
		return auth.AuthMethodJWT
	}
	return auth.AuthMethodOAuth
}

func effectiveScopeForDoctor(cfg *config.Config, token *auth.Token, sourceMap map[string]string) (string, string) {
	if cfg != nil && strings.TrimSpace(cfg.Scope) != "" {
		return cfg.Scope, sourceMap["scope"]
	}
	method := inferAuthMethod(cfg, token)
	if method == auth.AuthMethodJWT {
		return defaultJWTScope, "default"
	}
	return defaultOAuthScope, "default"
}

func callbackListenerCheck(method auth.AuthMethod) error {
	if method == auth.AuthMethodJWT {
		return nil
	}
	ln, _, err := auth.FindAvailableListener(8484, 8494)
	if err != nil {
		return err
	}
	return ln.Close()
}

func buildDoctorSourceMap(pc *config.ProfileConfig, selectedProfile string) map[string]string {
	sourceMap := make(map[string]string, len(config.AllKeys))
	var profileCfg *config.Config
	if pc != nil && pc.Profiles != nil {
		profileCfg = pc.Profiles[selectedProfile]
	}
	for _, key := range config.AllKeys {
		envVar := config.EnvVarForKey(key)
		if envVar != "" && strings.TrimSpace(os.Getenv(envVar)) != "" {
			sourceMap[key] = "env"
			continue
		}
		if profileCfg != nil {
			if value, err := profileCfg.Get(key); err == nil && strings.TrimSpace(value) != "" {
				sourceMap[key] = "profile"
				continue
			}
		}
		sourceMap[key] = "default"
	}
	return sourceMap
}

func cloneConfig(cfg *config.Config) *config.Config {
	if cfg == nil {
		return &config.Config{}
	}
	cloned := *cfg
	return &cloned
}

func buildDoctorRemoteTransports(input authdoctor.Input) authdoctor.RemoteTransports {
	transports := authdoctor.RemoteTransports{}
	if input.Token != nil && strings.TrimSpace(input.Token.AccessToken) != "" {
		transports.Directory = newReadonlyHTTPTransport(apiBaseURL, input.Token.AccessToken)
	}
	if input.EffectiveConfig != nil && strings.TrimSpace(input.EffectiveConfig.ScimAccessToken) != "" {
		transports.SCIM = newReadonlyHTTPTransport(scimBaseURL, input.EffectiveConfig.ScimAccessToken)
	}
	return transports
}

func newReadonlyHTTPTransport(baseURL, accessToken string) authdoctor.ReadonlyTransport {
	return readonlyHTTPTransport{
		baseURL:     strings.TrimRight(baseURL, "/"),
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: httputil.NewSecureTransport(),
		},
	}
}

func (t readonlyHTTPTransport) Get(path string, query map[string]string) ([]byte, int, error) {
	values := url.Values{}
	for key, value := range query {
		values.Set(key, value)
	}
	fullURL := t.baseURL + path
	if encoded := values.Encode(); encoded != "" {
		fullURL += "?" + encoded
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("원격 probe 요청 생성 실패: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.accessToken)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("원격 probe 네트워크 에러: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("원격 probe 응답 읽기 실패: %w", err)
	}
	if resp.StatusCode >= 400 {
		return body, resp.StatusCode, api.DecodeAPIError(resp.StatusCode, body)
	}
	return body, resp.StatusCode, nil
}

func printDoctorResult(result authdoctor.Result) error {
	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("doctor 결과 직렬화 실패: %w", err)
	}

	formatter := output.NewFormatter(outputFormat, os.Stdout).WithTable(
		[]string{"check_id", "status", "severity", "verification", "message"},
		"checks",
	)
	formatter.PrintRaw(body)
	return nil
}
