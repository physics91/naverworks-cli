package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var baselineEnvKeys = []string{
	"NW_PROFILE",
	"NW_CLIENT_ID",
	"NW_CLIENT_SECRET",
	"NW_SERVICE_ACCOUNT_ID",
	"NW_PRIVATE_KEY_PATH",
	"NW_DOMAIN_ID",
	"NW_BOT_ID",
	"NW_SCOPE",
	"NW_DEFAULT_CALENDAR_USER_ID",
	"NW_SCIM_ACCESS_TOKEN",
}

type Harness struct {
	t           *testing.T
	homeDir     string
	mu          sync.Mutex
	requestLogs []RequestLog
	scriptIndex int
	scripts     []ResponseScript
}

type CaptureResult struct {
	Stdout string
	Stderr string
}

type Runner func(args []string) error

type FailureCategory string

const (
	SetupFailure            FailureCategory = "SetupFailure"
	RequestShapeFailure     FailureCategory = "RequestShapeFailure"
	ResponseHandlingFailure FailureCategory = "ResponseHandlingFailure"
	SideEffectFailure       FailureCategory = "SideEffectFailure"
	UXContractFailure       FailureCategory = "UXContractFailure"
)

type ResponseScript struct {
	StatusCode int
	Headers    map[string]string
	Body       string
}

type RequestLog struct {
	Method   string
	Path     string
	RawQuery string
	Headers  map[string]string
	Body     string
}

func NewHarness(t *testing.T) *Harness {
	t.Helper()

	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, ".config"))
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", tmpDir)
	}

	for _, key := range baselineEnvKeys {
		t.Setenv(key, "")
	}

	return &Harness{
		t:       t,
		homeDir: tmpDir,
	}
}

func FromCurrentEnv(t *testing.T) *Harness {
	t.Helper()

	return &Harness{
		t:       t,
		homeDir: os.Getenv("HOME"),
	}
}

func (h *Harness) HomeDir() string {
	h.t.Helper()
	return h.homeDir
}

func (h *Harness) Capture(fn func() error) (CaptureResult, error) {
	h.t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		h.t.Fatalf("stdout pipe failed: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		h.t.Fatalf("stderr pipe failed: %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	stdoutCh := make(chan string, 1)
	stderrCh := make(chan string, 1)

	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stdoutR)
		_ = stdoutR.Close()
		stdoutCh <- buf.String()
	}()
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, stderrR)
		_ = stderrR.Close()
		stderrCh <- buf.String()
	}()

	runErr := fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	return CaptureResult{
		Stdout: <-stdoutCh,
		Stderr: <-stderrCh,
	}, runErr
}

func (h *Harness) Run(args []string, runner Runner) (CaptureResult, error) {
	h.t.Helper()

	return h.Capture(func() error {
		return runner(args)
	})
}

func (h *Harness) StartScriptedServer(scripts []ResponseScript) *httptest.Server {
	h.t.Helper()

	h.mu.Lock()
	h.scripts = append([]ResponseScript(nil), scripts...)
	h.scriptIndex = 0
	h.requestLogs = nil
	h.mu.Unlock()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.t.Fatalf("read request body failed: %v", err)
		}
		_ = r.Body.Close()

		headers := make(map[string]string, len(r.Header))
		for key, values := range r.Header {
			headers[key] = values[0]
		}

		h.mu.Lock()
		h.requestLogs = append(h.requestLogs, RequestLog{
			Method:   r.Method,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
			Headers:  headers,
			Body:     string(bodyBytes),
		})

		script := ResponseScript{StatusCode: http.StatusOK}
		if h.scriptIndex < len(h.scripts) {
			script = h.scripts[h.scriptIndex]
		}
		h.scriptIndex++
		h.mu.Unlock()

		for key, value := range script.Headers {
			w.Header().Set(key, value)
		}
		if script.StatusCode == 0 {
			script.StatusCode = http.StatusOK
		}
		w.WriteHeader(script.StatusCode)
		_, _ = io.WriteString(w, script.Body)
	}))
}

func (h *Harness) RequestLogs() []RequestLog {
	h.t.Helper()

	h.mu.Lock()
	defer h.mu.Unlock()

	logs := make([]RequestLog, 0, len(h.requestLogs))
	for _, entry := range h.requestLogs {
		headerCopy := make(map[string]string, len(entry.Headers))
		for key, value := range entry.Headers {
			headerCopy[key] = value
		}
		logs = append(logs, RequestLog{
			Method:   entry.Method,
			Path:     entry.Path,
			RawQuery: entry.RawQuery,
			Headers:  headerCopy,
			Body:     entry.Body,
		})
	}
	return logs
}

func Fatalf(t *testing.T, category FailureCategory, format string, args ...any) {
	t.Helper()
	t.Fatal(fmt.Errorf("%s", formatFailure(category, format, args...)))
}

func formatFailure(category FailureCategory, format string, args ...any) string {
	return fmt.Sprintf("%s: %s", category, fmt.Sprintf(format, args...))
}
