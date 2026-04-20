package authdoctor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/physics91/naverworks-cli/internal/config"
)

type Status string

const (
	StatusPass    Status = "pass"
	StatusFail    Status = "fail"
	StatusUnknown Status = "unknown"
)

type Verification string

const (
	VerificationLocalStatic     Verification = "local_static"
	VerificationRemoteConfirmed Verification = "remote_confirmed"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
)

const (
	VerificationScopeLocalOnly   = "local_only"
	VerificationScopeRemoteOptIn = "remote_opt_in"
	ContractVersion              = "v1"
	limitationNotice             = "로컬 판정 한계: 서버 측 enforcement 및 실시간 endpoint 상태는 아직 검증되지 않았습니다"
)

type Input struct {
	SelectedProfile       string
	SelectedProfileSource string
	EffectiveConfig       *config.Config
	SourceMap             map[string]string
	InferredAuthMethod    auth.AuthMethod
	EffectiveScope        string
	EffectiveScopeSource  string
	ScimBaseURL           string
	Token                 *auth.Token
	SelectionErr          error
	TokenErr              error
	CallbackListenerErr   error
}

type Result struct {
	ContractVersion   string        `json:"contract_version"`
	Profile           ProfileInfo   `json:"profile"`
	VerificationScope string        `json:"verification_scope"`
	SummaryBanner     SummaryBanner `json:"summary_banner"`
	NotChecked        []string      `json:"not_checked,omitempty"`
	UncheckedCount    int           `json:"unchecked_count"`
	Checks            []Check       `json:"checks"`
}

type ProfileInfo struct {
	Selected string `json:"selected"`
	Source   string `json:"source"`
}

type SummaryBanner struct {
	UncheckedCount   int    `json:"unchecked_count"`
	LimitationNotice string `json:"limitation_notice"`
}

type Check struct {
	CheckID         string         `json:"check_id"`
	Group           string         `json:"group"`
	Severity        Severity       `json:"severity"`
	FailureClass    string         `json:"failure_class"`
	Status          Status         `json:"status"`
	Verification    Verification   `json:"verification"`
	Message         string         `json:"message"`
	Evidence        map[string]any `json:"evidence,omitempty"`
	NextAction      string         `json:"next_action,omitempty"`
	RemediationCode string         `json:"remediation_code,omitempty"`
}

type ReadonlyTransport interface {
	Get(path string, query map[string]string) ([]byte, int, error)
}

type RemoteTransports struct {
	Directory ReadonlyTransport
	SCIM      ReadonlyTransport
}

type checkDefinition struct {
	ID                  string
	Group               string
	DefaultSeverity     Severity
	FailureClass        string
	AllowedEvidenceKeys []string
	NextAction          string
	RemediationCode     string
}

var checkDefinitions = []checkDefinition{
	{
		ID:                  "config.profile_selected",
		Group:               "config",
		DefaultSeverity:     SeverityInfo,
		FailureClass:        "config",
		AllowedEvidenceKeys: []string{"selected_profile", "selected_profile_source"},
		NextAction:          "naverworks config list",
		RemediationCode:     "profile_selection_checked",
	},
	{
		ID:                  "config.client_credentials",
		Group:               "config",
		DefaultSeverity:     SeverityWarning,
		FailureClass:        "config",
		AllowedEvidenceKeys: []string{"missing_keys", "client_id_source", "client_secret_source"},
		NextAction:          "naverworks auth setup",
		RemediationCode:     "client_credentials_required",
	},
	{
		ID:                  "token.present",
		Group:               "token",
		DefaultSeverity:     SeverityWarning,
		FailureClass:        "token",
		AllowedEvidenceKeys: []string{"auth_method"},
		NextAction:          "naverworks auth login",
		RemediationCode:     "token_login_required",
	},
	{
		ID:                  "token.expiry",
		Group:               "token",
		DefaultSeverity:     SeverityWarning,
		FailureClass:        "token",
		AllowedEvidenceKeys: []string{"expires_at", "auth_method"},
		NextAction:          "naverworks auth refresh",
		RemediationCode:     "token_expiry_checked",
	},
	{
		ID:                  "oauth.callback_listener",
		Group:               "oauth",
		DefaultSeverity:     SeverityWarning,
		FailureClass:        "callback",
		AllowedEvidenceKeys: []string{"auth_method"},
		NextAction:          "naverworks auth login",
		RemediationCode:     "oauth_callback_listener_checked",
	},
	{
		ID:                  "scope.core_configured",
		Group:               "scope",
		DefaultSeverity:     SeverityWarning,
		FailureClass:        "scope",
		AllowedEvidenceKeys: []string{"configured_scopes", "missing_scopes", "value_source"},
		NextAction:          "naverworks auth doctor --verify-remote",
		RemediationCode:     "scope_core_config_checked",
	},
	{
		ID:                  "scim.token_present",
		Group:               "scim",
		DefaultSeverity:     SeverityInfo,
		FailureClass:        "scim",
		AllowedEvidenceKeys: []string{"value_source"},
		NextAction:          "naverworks config set scim_access_token --stdin",
		RemediationCode:     "scim_token_presence_checked",
	},
	{
		ID:                  "scim.endpoint_configured",
		Group:               "scim",
		DefaultSeverity:     SeverityInfo,
		FailureClass:        "scim",
		AllowedEvidenceKeys: []string{"endpoint"},
		NextAction:          "naverworks auth doctor --verify-remote",
		RemediationCode:     "scim_endpoint_config_checked",
	},
	{
		ID:                  "scope.directory_remote_probe",
		Group:               "scope",
		DefaultSeverity:     SeverityWarning,
		FailureClass:        "scope",
		AllowedEvidenceKeys: []string{"probe_endpoint", "expected_status", "status_code"},
		NextAction:          "naverworks auth login",
		RemediationCode:     "directory_remote_probe_checked",
	},
	{
		ID:                  "scim.endpoint_reachable",
		Group:               "scim",
		DefaultSeverity:     SeverityWarning,
		FailureClass:        "scim",
		AllowedEvidenceKeys: []string{"probe_endpoint", "expected_status", "status_code"},
		NextAction:          "naverworks config set scim_access_token --stdin",
		RemediationCode:     "scim_endpoint_reachability_checked",
	},
}

var checkDefinitionByID = func() map[string]checkDefinition {
	out := make(map[string]checkDefinition, len(checkDefinitions))
	for _, def := range checkDefinitions {
		out[def.ID] = def
	}
	return out
}()

var remoteCheckIDs = []string{
	"scope.directory_remote_probe",
	"scim.endpoint_reachable",
}

func RunLocal(input Input) Result {
	result := Result{
		ContractVersion:   ContractVersion,
		Profile:           ProfileInfo{Selected: defaultString(input.SelectedProfile, "default"), Source: defaultString(input.SelectedProfileSource, "config")},
		VerificationScope: VerificationScopeLocalOnly,
		NotChecked:        append([]string(nil), remoteCheckIDs...),
	}

	checks := []Check{
		evaluateProfileSelected(input),
		evaluateClientCredentials(input),
		evaluateTokenPresent(input),
		evaluateTokenExpiry(input),
		evaluateCallbackListener(input),
		evaluateScopeCoreConfigured(input),
		evaluateScimTokenPresent(input),
		evaluateScimEndpointConfigured(input),
	}
	result.Checks = checks
	result.UncheckedCount = len(result.NotChecked)
	result.SummaryBanner = SummaryBanner{
		UncheckedCount:   result.UncheckedCount,
		LimitationNotice: limitationNotice,
	}
	return result
}

func ApplyRemoteVerification(result *Result, input Input, transports RemoteTransports) {
	if result == nil {
		return
	}
	result.VerificationScope = VerificationScopeRemoteOptIn

	if input.Token != nil && transports.Directory != nil {
		result.Checks = append(result.Checks, evaluateRemoteProbe(
			"scope.directory_remote_probe",
			transports.Directory,
			"/users",
			map[string]string{"count": "1"},
			"directory read 원격 확인에 성공했습니다",
			"directory read 원격 확인이 실패했습니다",
		))
		removeNotChecked(result, "scope.directory_remote_probe")
	}

	if input.EffectiveConfig != nil && strings.TrimSpace(input.EffectiveConfig.ScimAccessToken) != "" && transports.SCIM != nil {
		result.Checks = append(result.Checks, evaluateRemoteProbe(
			"scim.endpoint_reachable",
			transports.SCIM,
			"/Users",
			map[string]string{"count": "1"},
			"SCIM endpoint 원격 확인에 성공했습니다",
			"SCIM endpoint 원격 확인이 실패했습니다",
		))
		removeNotChecked(result, "scim.endpoint_reachable")
	}

	result.UncheckedCount = len(result.NotChecked)
	result.SummaryBanner.UncheckedCount = result.UncheckedCount
}

func evaluateProfileSelected(input Input) Check {
	evidence := map[string]any{
		"selected_profile":        defaultString(input.SelectedProfile, "default"),
		"selected_profile_source": defaultString(input.SelectedProfileSource, "config"),
	}
	if input.SelectionErr != nil {
		return buildCheck("config.profile_selected", StatusFail, VerificationLocalStatic, fmt.Sprintf("활성 프로필을 해석하지 못했습니다: %v", input.SelectionErr), evidence)
	}
	return buildCheck("config.profile_selected", StatusPass, VerificationLocalStatic, fmt.Sprintf("활성 프로필은 %q입니다", defaultString(input.SelectedProfile, "default")), evidence)
}

func evaluateClientCredentials(input Input) Check {
	cfg := ensureConfig(input.EffectiveConfig)
	missing := make([]string, 0, 2)
	if strings.TrimSpace(cfg.ClientID) == "" {
		missing = append(missing, "client_id")
	}
	if strings.TrimSpace(cfg.ClientSecret) == "" {
		missing = append(missing, "client_secret")
	}
	evidence := map[string]any{
		"missing_keys":         missing,
		"client_id_source":     sourceForKey(input.SourceMap, "client_id"),
		"client_secret_source": sourceForKey(input.SourceMap, "client_secret"),
	}
	if len(missing) > 0 {
		return buildCheck("config.client_credentials", StatusFail, VerificationLocalStatic, fmt.Sprintf("인증에 필요한 설정이 누락되었습니다: %s", strings.Join(missing, ", ")), evidence)
	}
	return buildCheck("config.client_credentials", StatusPass, VerificationLocalStatic, "client_id와 client_secret이 준비되어 있습니다", evidence)
}

func evaluateTokenPresent(input Input) Check {
	evidence := map[string]any{}
	if input.Token != nil {
		evidence["auth_method"] = string(input.Token.AuthMethod)
	}
	if input.TokenErr != nil {
		return buildCheck("token.present", StatusFail, VerificationLocalStatic, fmt.Sprintf("토큰 정보를 읽지 못했습니다: %v", input.TokenErr), evidence)
	}
	if input.Token == nil {
		return buildCheck("token.present", StatusFail, VerificationLocalStatic, "로그인 토큰이 없습니다", evidence)
	}
	return buildCheck("token.present", StatusPass, VerificationLocalStatic, "로그인 토큰이 존재합니다", evidence)
}

func evaluateTokenExpiry(input Input) Check {
	if input.Token == nil {
		return buildCheck("token.expiry", StatusUnknown, VerificationLocalStatic, "토큰이 없어 만료 시각을 판단할 수 없습니다", nil)
	}
	evidence := map[string]any{
		"expires_at":  input.Token.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
		"auth_method": string(input.Token.AuthMethod),
	}
	if input.Token.IsExpired() {
		return buildCheck("token.expiry", StatusFail, VerificationLocalStatic, "토큰이 이미 만료되었습니다", evidence)
	}
	if input.Token.NeedsRefresh() {
		return buildCheck("token.expiry", StatusPass, VerificationLocalStatic, "토큰이 곧 만료될 예정입니다", evidence)
	}
	return buildCheck("token.expiry", StatusPass, VerificationLocalStatic, "토큰 만료 시각이 유효합니다", evidence)
}

func evaluateCallbackListener(input Input) Check {
	evidence := map[string]any{"auth_method": string(input.InferredAuthMethod)}
	if input.InferredAuthMethod == auth.AuthMethodJWT {
		return buildCheck("oauth.callback_listener", StatusUnknown, VerificationLocalStatic, "JWT 기준으로 추정되어 callback listener 점검을 건너뜁니다", evidence)
	}
	if input.CallbackListenerErr != nil {
		return buildCheck("oauth.callback_listener", StatusFail, VerificationLocalStatic, fmt.Sprintf("OAuth callback listener를 열 수 없습니다: %v", input.CallbackListenerErr), evidence)
	}
	return buildCheck("oauth.callback_listener", StatusPass, VerificationLocalStatic, "OAuth callback listener를 열 수 있습니다", evidence)
}

func evaluateScopeCoreConfigured(input Input) Check {
	scopeList := strings.Fields(strings.TrimSpace(input.EffectiveScope))
	evidence := map[string]any{
		"configured_scopes": scopeList,
		"value_source":      defaultString(input.EffectiveScopeSource, "default"),
	}

	if len(scopeList) == 0 {
		return buildCheck("scope.core_configured", StatusUnknown, VerificationLocalStatic, "유효한 scope 정보를 확인하지 못했습니다", evidence)
	}

	missing := missingCoreScopes(input.InferredAuthMethod, scopeList)
	if len(missing) > 0 {
		evidence["missing_scopes"] = missing
		return buildCheck("scope.core_configured", StatusFail, VerificationLocalStatic, fmt.Sprintf("핵심 scope가 누락되었을 가능성이 있습니다: %s", strings.Join(missing, ", ")), evidence)
	}

	return buildCheck("scope.core_configured", StatusPass, VerificationLocalStatic, "핵심 scope 구성이 확인되었습니다", evidence)
}

func evaluateScimTokenPresent(input Input) Check {
	cfg := ensureConfig(input.EffectiveConfig)
	evidence := map[string]any{
		"value_source": sourceForKey(input.SourceMap, "scim_access_token"),
	}
	if strings.TrimSpace(cfg.ScimAccessToken) == "" {
		return buildCheck("scim.token_present", StatusUnknown, VerificationLocalStatic, "SCIM 토큰이 설정되지 않았습니다. SCIM을 사용하지 않으면 무시해도 됩니다", evidence)
	}
	return buildCheck("scim.token_present", StatusPass, VerificationLocalStatic, "SCIM 토큰이 설정되어 있습니다", evidence)
}

func evaluateScimEndpointConfigured(input Input) Check {
	evidence := map[string]any{
		"endpoint": defaultString(strings.TrimSpace(input.ScimBaseURL), "https://www.worksapis.com/scim/v2"),
	}
	if strings.TrimSpace(input.ScimBaseURL) == "" {
		return buildCheck("scim.endpoint_configured", StatusFail, VerificationLocalStatic, "SCIM endpoint 기본값이 비어 있습니다", evidence)
	}
	return buildCheck("scim.endpoint_configured", StatusPass, VerificationLocalStatic, "SCIM endpoint 기본값이 준비되어 있습니다", evidence)
}

func evaluateRemoteProbe(checkID string, transport ReadonlyTransport, path string, query map[string]string, successMessage string, failurePrefix string) Check {
	evidence := map[string]any{
		"probe_endpoint":  path,
		"expected_status": 200,
	}
	_, statusCode, err := transport.Get(path, query)
	if statusCode > 0 {
		evidence["status_code"] = statusCode
	}
	if err != nil {
		status := StatusUnknown
		if statusCode == 401 || statusCode == 403 {
			status = StatusFail
		}
		return buildCheck(checkID, status, VerificationRemoteConfirmed, fmt.Sprintf("%s: %v", failurePrefix, err), evidence)
	}
	return buildCheck(checkID, StatusPass, VerificationRemoteConfirmed, successMessage, evidence)
}

func buildCheck(checkID string, status Status, verification Verification, message string, evidence map[string]any) Check {
	def := checkDefinitionByID[checkID]
	check := Check{
		CheckID:         checkID,
		Group:           def.Group,
		Severity:        severityForStatus(def.DefaultSeverity, status),
		FailureClass:    def.FailureClass,
		Status:          status,
		Verification:    verification,
		Message:         message,
		Evidence:        sanitizeEvidence(def, evidence),
		NextAction:      def.NextAction,
		RemediationCode: def.RemediationCode,
	}
	if len(check.Evidence) == 0 {
		check.Evidence = nil
	}
	return check
}

func severityForStatus(defaultSeverity Severity, status Status) Severity {
	if status == StatusFail {
		return SeverityWarning
	}
	if status == StatusUnknown {
		return SeverityInfo
	}
	if defaultSeverity == "" {
		return SeverityInfo
	}
	return SeverityInfo
}

func sanitizeEvidence(def checkDefinition, evidence map[string]any) map[string]any {
	if len(evidence) == 0 {
		return nil
	}

	allowed := make(map[string]struct{}, len(def.AllowedEvidenceKeys))
	for _, key := range def.AllowedEvidenceKeys {
		allowed[key] = struct{}{}
	}

	out := make(map[string]any, len(allowed))
	keys := make([]string, 0, len(evidence))
	for key := range evidence {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, ok := allowed[key]; !ok {
			continue
		}
		value := evidence[key]
		if isSensitiveEvidenceKey(key) {
			out[key] = "****"
			continue
		}
		out[key] = value
	}
	return out
}

func isSensitiveEvidenceKey(key string) bool {
	if config.IsSensitiveKey(key) {
		return true
	}
	switch key {
	case "access_token", "refresh_token", "token", "authorization":
		return true
	default:
		return false
	}
}

func removeNotChecked(result *Result, checkID string) {
	if result == nil {
		return
	}
	filtered := result.NotChecked[:0]
	for _, id := range result.NotChecked {
		if id != checkID {
			filtered = append(filtered, id)
		}
	}
	result.NotChecked = filtered
}

func ensureConfig(cfg *config.Config) *config.Config {
	if cfg == nil {
		return &config.Config{}
	}
	return cfg
}

func sourceForKey(sourceMap map[string]string, key string) string {
	if sourceMap == nil {
		return "default"
	}
	if value, ok := sourceMap[key]; ok && value != "" {
		return value
	}
	return "default"
}

func missingCoreScopes(method auth.AuthMethod, configured []string) []string {
	required := []string{}
	switch method {
	case auth.AuthMethodJWT:
		required = []string{"directory"}
	default:
		required = []string{"openid", "profile"}
	}
	missing := make([]string, 0, len(required))
	for _, req := range required {
		if !contains(configured, req) {
			missing = append(missing, req)
		}
	}
	return missing
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
