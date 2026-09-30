package web

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func browserExecutable(explicit string) (string, error) {
	if strings.TrimSpace(explicit) == "" {
		explicit = os.Getenv("NW_BROWSER_PATH")
	}
	if strings.TrimSpace(explicit) != "" {
		path, err := exec.LookPath(explicit)
		if err != nil {
			return "", fmt.Errorf("브라우저 실행 파일을 찾을 수 없습니다. NW_BROWSER_PATH를 확인하세요: %w", err)
		}
		return path, nil
	}

	// Use the same interactive-capable executable for login and headless room
	// commands. An automatically selected headless shell cannot show login UI.
	candidates := []string{"google-chrome", "google-chrome-stable"}
	switch runtime.GOOS {
	case "linux":
		candidates = append(candidates, "/usr/bin/google-chrome", "/opt/google/chrome/chrome")
	case "darwin":
		candidates = append(candidates, "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
	case "windows":
		candidates = append(candidates, "chrome", "chrome.exe",
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`)
		if home, err := os.UserHomeDir(); err == nil {
			candidates = append(candidates, filepath.Join(home, "AppData", "Local", "Google", "Chrome", "Application", "chrome.exe"))
		}
	}
	candidates = append(candidates, "chromium", "chromium-browser", "chrome", "chrome.exe")
	switch runtime.GOOS {
	case "linux":
		candidates = append(candidates, "/usr/bin/chromium", "/usr/bin/chromium-browser", "/snap/bin/chromium")
	case "darwin":
		candidates = append(candidates, "/Applications/Chromium.app/Contents/MacOS/Chromium")
	}
	return findBrowserExecutable(candidates)
}

func findBrowserExecutable(candidates []string) (string, error) {
	for _, candidate := range candidates {
		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("Chrome/Chromium을 찾을 수 없습니다. 브라우저를 설치하거나 NW_BROWSER_PATH로 실행 파일을 지정하세요 (자동 설치하지 않습니다)")
}
