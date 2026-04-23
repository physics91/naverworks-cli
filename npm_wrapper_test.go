package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestNpmWrapperLauncherFallsBackToPlatformPackage(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("node launcher test uses unix executable fixtures")
	}

	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is required for npm wrapper test")
	}

	packageName, packageDir, binaryName := currentNpmPlatformFixture(t)

	tempDir := t.TempDir()
	cliDir := filepath.Join(tempDir, "cli")
	binDir := filepath.Join(cliDir, "bin")
	depDir := filepath.Join(cliDir, "node_modules", packageDir)
	launcherPath := filepath.Join(binDir, "naverworks.js")

	mkdirAll(t, binDir)
	mkdirAll(t, depDir)

	copyFile(t, filepath.Join("npm", "cli", "bin", "naverworks"), launcherPath)
	copyFileIfExists(t, filepath.Join("npm", "cli", "platform.js"), filepath.Join(cliDir, "platform.js"))

	writeFile(t, filepath.Join(depDir, "package.json"), `{
  "name": "`+packageName+`",
  "version": "0.0.0"
}
`)
	writeFile(t, filepath.Join(depDir, binaryName), "#!/bin/sh\nprintf 'bun fallback ok\\n'\n")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "node", launcherPath, "version")
	cmd.Dir = cliDir

	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("launcher run timed out, likely recursive launcher path\n%s", out)
	}
	if err != nil {
		t.Fatalf("launcher run failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "bun fallback ok") {
		t.Fatalf("expected fallback binary output, got:\n%s", out)
	}
}

func TestNpmWrapperInstallCopiesSidecarBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("node installer test uses unix executable fixtures")
	}

	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is required for npm wrapper test")
	}

	packageName, packageDir, binaryName := currentNpmPlatformFixture(t)

	tempDir := t.TempDir()
	cliDir := filepath.Join(tempDir, "cli")
	depDir := filepath.Join(cliDir, "node_modules", packageDir)
	sidecarPath := filepath.Join(cliDir, "bin", "naverworks-bin")

	mkdirAll(t, depDir)

	copyFile(t, filepath.Join("npm", "cli", "install.js"), filepath.Join(cliDir, "install.js"))
	copyFileIfExists(t, filepath.Join("npm", "cli", "platform.js"), filepath.Join(cliDir, "platform.js"))

	writeFile(t, filepath.Join(depDir, "package.json"), `{
  "name": "`+packageName+`",
  "version": "0.0.0"
}
`)
	writeFile(t, filepath.Join(depDir, binaryName), "#!/bin/sh\nprintf 'sidecar copy ok\\n'\n")

	cmd := exec.Command("node", filepath.Join(cliDir, "install.js"))
	cmd.Dir = cliDir

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("install.js failed: %v\n%s", err, out)
	}
	if _, err := os.Stat(sidecarPath); err != nil {
		t.Fatalf("expected sidecar binary at %s: %v\n%s", sidecarPath, err, out)
	}
}

func currentNpmPlatformFixture(t *testing.T) (packageName, packageDir, binaryName string) {
	t.Helper()

	switch {
	case runtime.GOOS == "linux" && runtime.GOARCH == "amd64":
		return "@physics91org/linux-x64", filepath.Join("@physics91org", "linux-x64"), "naverworks"
	case runtime.GOOS == "linux" && runtime.GOARCH == "arm64":
		return "@physics91org/linux-arm64", filepath.Join("@physics91org", "linux-arm64"), "naverworks"
	case runtime.GOOS == "darwin" && runtime.GOARCH == "amd64":
		return "@physics91org/darwin-x64", filepath.Join("@physics91org", "darwin-x64"), "naverworks"
	case runtime.GOOS == "darwin" && runtime.GOARCH == "arm64":
		return "@physics91org/darwin-arm64", filepath.Join("@physics91org", "darwin-arm64"), "naverworks"
	default:
		t.Skipf("unsupported platform fixture for %s/%s", runtime.GOOS, runtime.GOARCH)
		return "", "", ""
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func copyFile(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	writeFile(t, dst, string(data))
}

func copyFileIfExists(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("read %s: %v", src, err)
	}
	writeFile(t, dst, string(data))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
