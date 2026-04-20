package cmd

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
	"github.com/spf13/cobra"
)

// --- URL Validator Tests ---

func TestMakeAuthURLValidator_ValidHTTPS(t *testing.T) {
	validator := makeAuthURLValidator("https://auth.worksmobile.com/oauth2/v2.0")
	err := validator("https://auth.worksmobile.com/oauth2/v2.0/authorize?client_id=foo")
	if err != nil {
		t.Fatalf("valid URL should pass: %v", err)
	}
}

func TestMakeAuthURLValidator_RejectsHTTP(t *testing.T) {
	validator := makeAuthURLValidator("https://auth.worksmobile.com/oauth2/v2.0")
	err := validator("http://auth.worksmobile.com/oauth2/v2.0/authorize?client_id=foo")
	if err == nil {
		t.Fatal("http should be rejected")
	}
	if !strings.Contains(err.Error(), "스키마") {
		t.Errorf("error should mention scheme: %v", err)
	}
}

func TestMakeAuthURLValidator_RejectsWrongHost(t *testing.T) {
	validator := makeAuthURLValidator("https://auth.worksmobile.com/oauth2/v2.0")
	err := validator("https://evil.com/oauth2/v2.0/authorize?client_id=foo")
	if err == nil {
		t.Fatal("wrong host should be rejected")
	}
	if !strings.Contains(err.Error(), "호스트") {
		t.Errorf("error should mention host: %v", err)
	}
}

func TestMakeAuthURLValidator_RejectsFileScheme(t *testing.T) {
	validator := makeAuthURLValidator("https://auth.worksmobile.com/oauth2/v2.0")
	err := validator("file:///etc/passwd")
	if err == nil {
		t.Fatal("file:// scheme should be rejected")
	}
}

func TestSanitizeSensitiveOutput_RedactsUploadURL(t *testing.T) {
	body := []byte(`{"uploadUrl":"https://example.com/upload?sig=secret","offset":7}`)

	sanitized := sanitizeSensitiveOutput(body)

	if strings.Contains(string(sanitized), "uploadUrl") {
		t.Fatalf("uploadUrl should be redacted: %s", sanitized)
	}
	if strings.Contains(string(sanitized), "secret") {
		t.Fatalf("signature-bearing URL should not remain: %s", sanitized)
	}
	if !strings.Contains(string(sanitized), `"uploaded":true`) {
		t.Fatalf("expected uploaded marker: %s", sanitized)
	}
	if !strings.Contains(string(sanitized), `"offset":7`) {
		t.Fatalf("expected safe fields to remain: %s", sanitized)
	}
}

func TestSanitizeSensitiveOutput_PreservesOtherBodies(t *testing.T) {
	body := []byte(`{"id":"123","name":"test"}`)

	sanitized := sanitizeSensitiveOutput(body)

	if string(sanitized) != string(body) {
		t.Fatalf("non-upload payload should be unchanged: got %s want %s", sanitized, body)
	}
}

// --- auth setup NonTTY Tests ---

func TestNonTTYErrorMessage_SuggestsEnvVar(t *testing.T) {
	msg := nonTTYErrorMessage()
	if !strings.Contains(msg, "NW_CLIENT_SECRET") {
		t.Errorf("should suggest NW_CLIENT_SECRET env var: %s", msg)
	}
}

func TestNonTTYErrorMessage_SuggestsStdin(t *testing.T) {
	msg := nonTTYErrorMessage()
	if !strings.Contains(msg, "--stdin") {
		t.Errorf("should suggest --stdin flag: %s", msg)
	}
}

func TestNonTTYErrorMessage_NoArgvSuggestion(t *testing.T) {
	msg := nonTTYErrorMessage()
	// config set client_secret <값> 형태의 argv 노출 안내가 없어야 함
	if strings.Contains(msg, "config set client_secret <") {
		t.Error("should not suggest argv-based secret setting")
	}
}

// --- stdin Limit Tests ---

func TestReadStdinLimited_WithinLimit(t *testing.T) {
	data := []byte(`{"key":"value"}`)
	result, err := readStdinLimited(bytes.NewReader(data), 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != string(data) {
		t.Errorf("expected %q, got %q", string(data), string(result))
	}
}

func TestReadStdinLimited_ExceedsLimit(t *testing.T) {
	data := make([]byte, 1024+1)
	for i := range data {
		data[i] = 'a'
	}
	_, err := readStdinLimited(bytes.NewReader(data), 1024)
	if err == nil {
		t.Fatal("expected error for oversized input")
	}
	if !strings.Contains(err.Error(), "너무 큽니다") {
		t.Errorf("error should mention size exceeded: %v", err)
	}
}

func TestReadStdinLimited_ExactLimit(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = 'a'
	}
	result, err := readStdinLimited(bytes.NewReader(data), 1024)
	if err != nil {
		t.Fatalf("exact limit should succeed: %v", err)
	}
	if len(result) != 1024 {
		t.Errorf("expected %d bytes, got %d", 1024, len(result))
	}
}

func TestReadStdinLimited_Empty(t *testing.T) {
	result, err := readStdinLimited(bytes.NewReader(nil), 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty, got %d bytes", len(result))
	}
}

// --- File Limit Tests ---

func TestReadFileFlagWithLimit_ExceedsLimit(t *testing.T) {
	tmp, err := os.CreateTemp("", "test-oversize-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	data := make([]byte, 1024+1)
	tmp.Write(data)
	tmp.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("file", tmp.Name(), "")

	_, _, err = readFileFlagWithLimit(cmd, "file", 1024)
	if err == nil {
		t.Fatal("expected error for oversized file")
	}
	if !strings.Contains(err.Error(), "크기 초과") {
		t.Errorf("error should mention size exceeded: %v", err)
	}
}

func TestReadFileFlagWithLimit_WithinLimit(t *testing.T) {
	tmp, err := os.CreateTemp("", "test-ok-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	tmp.Write([]byte("hello"))
	tmp.Close()

	cmd := &cobra.Command{}
	cmd.Flags().String("file", tmp.Name(), "")

	data, name, err := readFileFlagWithLimit(cmd, "file", 1024)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}
	if name == "" {
		t.Error("expected non-empty filename")
	}
}

func TestReadFileFlagWithLimit_EmptyFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("file", "", "")

	_, _, err := readFileFlagWithLimit(cmd, "file", 1024)
	if err == nil {
		t.Fatal("expected error for empty flag")
	}
}

func TestReadFileFlagWithLimit_NonRegularFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-only device file test")
	}
	cmd := &cobra.Command{}
	cmd.Flags().String("file", os.DevNull, "")

	_, _, err := readFileFlagWithLimit(cmd, "file", 1024)
	if err == nil {
		t.Fatal("expected error for non-regular file")
	}
	if !strings.Contains(err.Error(), "일반 파일만 허용합니다") {
		t.Errorf("error should mention regular files only: %v", err)
	}
}

func TestStatFileForUpload_NonRegularFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-only device file test")
	}

	_, _, err := statFileForUpload(os.DevNull)
	if err == nil {
		t.Fatal("expected error for non-regular upload file")
	}
	if !strings.Contains(err.Error(), "일반 파일만 허용합니다") {
		t.Errorf("error should mention regular files only: %v", err)
	}
}

func TestDoUploadFromResponse_PreviewSkipsUpload(t *testing.T) {
	client := api.NewClient("https://example.com", &auth.Token{
		AccessToken: "preview",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil).WithPreview(api.PreviewOptions{DryRun: true})

	body, err := doUploadFromResponse(client, []byte(`{"dry_run":true}`), "missing-file.txt")
	if err != nil {
		t.Fatalf("preview mode should skip upload follow-up: %v", err)
	}
	if !strings.Contains(string(body), `"next_step"`) {
		t.Fatalf("preview upload plan should include next_step: %s", body)
	}
}

func TestDoUploadFromResponse_GenerateInputPreservesBody(t *testing.T) {
	origGenerateInput := generateInput
	generateInput = true
	t.Cleanup(func() {
		generateInput = origGenerateInput
	})

	client := api.NewClient("https://example.com", &auth.Token{
		AccessToken: "preview",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil).WithPreview(api.PreviewOptions{GenerateInput: true})

	body, err := doUploadFromResponse(client, []byte(`{"fileName":"sample.txt","fileSize":12}`), "sample.txt")
	if err != nil {
		t.Fatalf("generate-input preview should preserve original body: %v", err)
	}
	if strings.Contains(string(body), `"next_step"`) {
		t.Fatalf("generate-input should not include upload next_step: %s", body)
	}
}

func TestDoUploadFromResponse_PreviewRewritesPlanFile(t *testing.T) {
	origPlanOutPath := planOutPath
	planOutPath = t.TempDir() + "/upload-plan.json"
	t.Cleanup(func() {
		planOutPath = origPlanOutPath
	})

	client := api.NewClient("https://example.com", &auth.Token{
		AccessToken: "preview",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil).WithPreview(api.PreviewOptions{DryRun: true, PlanOutPath: planOutPath})

	if _, err := doUploadFromResponse(client, []byte(`{"dry_run":true}`), "missing-file.txt"); err != nil {
		t.Fatalf("preview mode should rewrite plan file: %v", err)
	}
	data, err := os.ReadFile(planOutPath)
	if err != nil {
		t.Fatalf("expected rewritten plan file: %v", err)
	}
	if !strings.Contains(string(data), `"next_step"`) {
		t.Fatalf("rewritten plan file should include next_step: %s", data)
	}
}

func TestDoUploadFromResponse_GenerateInputRewritesPlanFile(t *testing.T) {
	origGenerateInput := generateInput
	origPlanOutPath := planOutPath
	generateInput = true
	planOutPath = t.TempDir() + "/upload-plan.json"
	t.Cleanup(func() {
		generateInput = origGenerateInput
		planOutPath = origPlanOutPath
	})

	if err := os.WriteFile(planOutPath, []byte(`{"dry_run":true}`), 0600); err != nil {
		t.Fatalf("failed to seed plan file: %v", err)
	}

	client := api.NewClient("https://example.com", &auth.Token{
		AccessToken: "preview",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil).WithPreview(api.PreviewOptions{GenerateInput: true, PlanOutPath: planOutPath})

	body, err := doUploadFromResponse(client, []byte(`{"fileName":"sample.txt","fileSize":12}`), "sample.txt")
	if err != nil {
		t.Fatalf("generate-input preview should rewrite plan file: %v", err)
	}
	if strings.Contains(string(body), `"next_step"`) {
		t.Fatalf("stdout should remain pure generated input: %s", body)
	}
	data, err := os.ReadFile(planOutPath)
	if err != nil {
		t.Fatalf("expected rewritten plan file: %v", err)
	}
	if !strings.Contains(string(data), `"next_step"`) {
		t.Fatalf("rewritten plan file should include next_step: %s", data)
	}
}
