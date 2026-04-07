package auth

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/physics91/naverworks-cli/internal/httputil"
)

var authHTTPClient = &http.Client{
	Timeout:   30 * time.Second,
	Transport: httputil.NewSecureTransport(),
}

func GenerateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("state 생성 실패: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func BuildAuthorizationURL(authBaseURL, clientID, redirectURI, state, scope string) string {
	params := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"state":         {state},
		"scope":         {scope},
		"response_type": {"code"},
	}
	return authBaseURL + "/authorize?" + params.Encode()
}

func ExchangeCode(authBaseURL, clientID, clientSecret, code, redirectURI string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	}
	return requestToken(authBaseURL+"/token", data)
}

func RefreshAccessToken(authBaseURL, clientID, clientSecret string, token *Token) error {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"refresh_token": {token.RefreshToken},
	}
	newToken, err := requestToken(authBaseURL+"/token", data)
	if err != nil {
		return err
	}
	token.AccessToken = newToken.AccessToken
	if newToken.RefreshToken != "" {
		token.RefreshToken = newToken.RefreshToken
	}
	token.ExpiresAt = newToken.ExpiresAt
	token.Scope = newToken.Scope
	return nil
}

func RequestJWTToken(authBaseURL, clientID, clientSecret, assertion, scope string) (*Token, error) {
	data := url.Values{
		"grant_type":    {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"assertion":     {assertion},
		"scope":         {scope},
	}
	token, err := requestToken(authBaseURL+"/token", data)
	if err != nil {
		return nil, err
	}
	token.AuthMethod = "jwt"
	return token, nil
}

func RevokeToken(authBaseURL, clientID, clientSecret, token, tokenTypeHint string) error {
	data := url.Values{
		"client_id":       {clientID},
		"client_secret":   {clientSecret},
		"token":           {token},
		"token_type_hint": {tokenTypeHint},
	}
	resp, err := authHTTPClient.PostForm(authBaseURL+"/revoke", data)
	if err != nil {
		return fmt.Errorf("revoke 요청 실패: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("revoke 실패: HTTP %d", resp.StatusCode)
	}
	return nil
}

func FindAvailableListener(start, end int) (net.Listener, int, error) {
	for port := start; port <= end; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, port, nil
		}
	}
	return nil, 0, fmt.Errorf("사용 가능한 포트를 찾을 수 없습니다 (%d-%d)", start, end)
}

func WaitForCallback(ln net.Listener, expectedState string, timeout time.Duration) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			errCh <- fmt.Errorf("state 불일치: 인증 요청이 유효하지 않습니다")
			w.WriteHeader(400)
			fmt.Fprint(w, "인증 실패: state 불일치")
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("인증 코드가 없습니다: %s", r.URL.Query().Get("error"))
			w.WriteHeader(400)
			fmt.Fprint(w, "인증 실패")
			return
		}
		w.WriteHeader(200)
		fmt.Fprint(w, "인증 완료! 이 창을 닫아도 됩니다.")
		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go server.Serve(ln)
	defer server.Close()

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("인증 타임아웃 (%v)", timeout)
	}
}

func ReadCallbackURLFromStdin(expectedState string) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		u, err := url.Parse(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "URL 형식 오류, 다시 붙여넣으세요:")
			fmt.Fprint(os.Stderr, "> ")
			continue
		}
		state := u.Query().Get("state")
		if state != expectedState {
			fmt.Fprintln(os.Stderr, "state 불일치, 다시 붙여넣으세요:")
			fmt.Fprint(os.Stderr, "> ")
			continue
		}
		code := u.Query().Get("code")
		if code == "" {
			fmt.Fprintln(os.Stderr, "code 파라미터 없음, 다시 붙여넣으세요:")
			fmt.Fprint(os.Stderr, "> ")
			continue
		}
		return code, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("stdin 읽기 실패: %w", err)
	}
	return "", fmt.Errorf("stdin 입력 없음")
}

func HasScope(scopeStr string, target string) bool {
	for _, s := range strings.Fields(scopeStr) {
		if s == target {
			return true
		}
	}
	return false
}

func FetchUserName(accessToken, authBaseURL string) (string, error) {
	req, err := http.NewRequest("GET", authBaseURL+"/userinfo", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := authHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("userinfo 조회 실패: HTTP %d", resp.StatusCode)
	}
	var info struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.Name, nil
}

func requestToken(tokenURL string, data url.Values) (*Token, error) {
	resp, err := authHTTPClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("토큰 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("토큰 응답 읽기 실패: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error     string `json:"error"`
			ErrorDesc string `json:"error_description"`
		}
		json.Unmarshal(body, &errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf("토큰 발급 실패 (HTTP %d): %s - %s", resp.StatusCode, errResp.Error, errResp.ErrorDesc)
		}
		return nil, fmt.Errorf("토큰 발급 실패: HTTP %d", resp.StatusCode)
	}

	var result struct {
		AccessToken  string          `json:"access_token"`
		RefreshToken string          `json:"refresh_token"`
		TokenType    string          `json:"token_type"`
		ExpiresIn    json.RawMessage `json:"expires_in"`
		Scope        string          `json:"scope"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("토큰 응답 파싱 실패: %w", err)
	}

	expiresIn := 3600
	if result.ExpiresIn != nil {
		raw := strings.Trim(string(result.ExpiresIn), `"`)
		if v, err := strconv.Atoi(raw); err == nil {
			expiresIn = v
		}
	}

	if result.AccessToken == "" {
		return nil, fmt.Errorf("토큰 응답에 access_token이 없습니다")
	}

	return &Token{
		AuthMethod:   "oauth",
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(expiresIn) * time.Second),
		Scope:        result.Scope,
	}, nil
}
