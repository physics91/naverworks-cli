package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

type binaryCanaryResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

var (
	canaryBuildOnce sync.Once
	canaryBuildDir  string
	canaryBuildPath string
	canaryBuildErr  error
)

func TestMain(m *testing.M) {
	code := m.Run()
	if canaryBuildDir != "" {
		_ = os.RemoveAll(canaryBuildDir)
	}
	os.Exit(code)
}

func TestBinaryCanaryVersion(t *testing.T) {
	result := runBinaryCanary(t, "version")
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr=%q)", result.ExitCode, result.Stderr)
	}
	if result.Stderr != "" {
		t.Fatalf("stderr = %q, want empty", result.Stderr)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(strings.TrimSpace(result.Stdout)), &payload); err != nil {
		t.Fatalf("version output is not valid JSON: %v\noutput: %q", err, result.Stdout)
	}
	for _, key := range []string{"version", "commit", "build_date"} {
		if _, ok := payload[key]; !ok {
			t.Fatalf("version output missing %q", key)
		}
	}
}

func runBinaryCanary(t *testing.T, args ...string) binaryCanaryResult {
	t.Helper()

	cmd := exec.Command(buildCanaryBinary(t), args...)
	cmd.Dir = repoRootDir(t)

	homeDir := t.TempDir()
	cmd.Env = append(os.Environ(),
		"HOME="+homeDir,
		"USERPROFILE="+homeDir,
		"APPDATA="+homeDir,
		"XDG_CONFIG_HOME="+filepath.Join(homeDir, ".config"),
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	result := binaryCanaryResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	err := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()
	if err == nil {
		return result
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		result.ExitCode = exitErr.ExitCode()
		return result
	}

	t.Fatalf("binary canary exec failed: %v", err)
	return binaryCanaryResult{}
}

func buildCanaryBinary(t *testing.T) string {
	t.Helper()

	canaryBuildOnce.Do(func() {
		buildDir, err := os.MkdirTemp("", "naverworks-canary-*")
		if err != nil {
			canaryBuildErr = fmt.Errorf("create canary build dir: %w", err)
			return
		}
		canaryBuildDir = buildDir
		canaryBuildPath = filepath.Join(buildDir, binaryName())

		cmd := exec.Command("go", "build", "-o", canaryBuildPath, ".")
		cmd.Dir = repoRootDir(t)
		output, err := cmd.CombinedOutput()
		if err != nil {
			canaryBuildErr = fmt.Errorf("build canary binary: %w\n%s", err, strings.TrimSpace(string(output)))
		}
	})

	if canaryBuildErr != nil {
		t.Fatalf("build canary binary failed: %v", canaryBuildErr)
	}

	return canaryBuildPath
}

func repoRootDir(t *testing.T) string {
	t.Helper()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve current file path")
	}

	return filepath.Dir(filepath.Dir(currentFile))
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "naverworks.exe"
	}
	return "naverworks"
}
