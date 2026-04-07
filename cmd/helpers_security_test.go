package cmd

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

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
