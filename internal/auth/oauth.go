package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func GenerateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
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
	resp, err := http.PostForm(authBaseURL+"/revoke", data)
	if err != nil {
		return fmt.Errorf("revoke 요청 실패: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("revoke 실패: HTTP %d", resp.StatusCode)
	}
	return nil
}

func FindAvailablePort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("사용 가능한 포트를 찾을 수 없습니다 (%d-%d)", start, end)
}

func WaitForCallback(port int, expectedState string, timeout time.Duration) (string, error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		state := r.URL.Query().Get("state")
		if state != expectedState {
			errCh <- fmt.Errorf("state 불일치: expected %q, got %q", expectedState, state)
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

	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go server.ListenAndServe()
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

func HasScope(scopeStr string, target string) bool {
	for _, s := range strings.Fields(scopeStr) {
		if s == target {
			return true
		}
	}
	return false
}

func requestToken(tokenURL string, data url.Values) (*Token, error) {
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("토큰 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("토큰 응답 파싱 실패: %w", err)
	}
	if result.Error != "" {
		return nil, fmt.Errorf("토큰 발급 실패: %s - %s", result.Error, result.ErrorDesc)
	}

	return &Token{
		AuthMethod:   "oauth",
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		TokenType:    result.TokenType,
		ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
		Scope:        result.Scope,
	}, nil
}
