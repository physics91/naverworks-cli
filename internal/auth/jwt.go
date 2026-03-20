package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func BuildJWTAssertion(clientID, serviceAccountID, privateKeyPath string) (string, error) {
	keyData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("private key 파일 읽기 실패: %w", err)
	}

	key, err := parsePrivateKey(keyData)
	if err != nil {
		return "", err
	}

	now := time.Now()
	header := map[string]string{"alg": "RS256", "typ": "JWT"}
	payload := map[string]interface{}{
		"iss": clientID,
		"sub": serviceAccountID,
		"iat": now.Unix(),
		"exp": now.Add(1 * time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := headerB64 + "." + payloadB64

	hash := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("JWT 서명 실패: %w", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)
	return signingInput + "." + signatureB64, nil
}

func parsePrivateKey(data []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("유효하지 않은 PEM 형식입니다. RSA PRIVATE KEY 또는 PRIVATE KEY 블록이 필요합니다")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("PKCS8 키 파싱 실패: %w", err)
		}
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("RSA 키가 아닙니다")
		}
		return rsaKey, nil
	default:
		return nil, fmt.Errorf("지원하지 않는 PEM 블록 타입: %s", block.Type)
	}
}

func CheckKeyPermissions(path string) string {
	if runtime.GOOS == "windows" {
		return checkWindowsKeyPermissions(path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	perm := info.Mode().Perm()
	if perm != 0600 {
		return fmt.Sprintf("경고: %s 파일 권한이 %04o입니다. 0600을 권장합니다", path, perm)
	}
	return ""
}

func checkWindowsKeyPermissions(path string) string {
	out, err := exec.Command("icacls", path).Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(string(out), "\n")
	user := os.Getenv("USERNAME")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "Successfully") {
			continue
		}
		if !strings.Contains(trimmed, ":") {
			continue
		}
		if !strings.Contains(strings.ToLower(trimmed), strings.ToLower(user)) {
			return fmt.Sprintf("경고: %s에 현재 사용자(%s) 외의 접근 권한이 설정되어 있습니다", path, user)
		}
	}
	return ""
}
