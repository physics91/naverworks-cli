package cmd

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/physics91/naverworks-cli/internal/api"
	"github.com/physics91/naverworks-cli/internal/auth"
)

func installHTTPSIntercept(t *testing.T, targetAddr string) {
	t.Helper()

	oldTransport := http.DefaultTransport
	baseTransport := http.DefaultTransport.(*http.Transport).Clone()
	baseTransport.TLSClientConfig.InsecureSkipVerify = true

	dialer := &net.Dialer{}
	baseTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		switch addr {
		case "www.worksapis.com:443", "example.com:443":
			addr = targetAddr
		}
		return dialer.DialContext(ctx, network, addr)
	}

	http.DefaultTransport = baseTransport
	t.Cleanup(func() {
		http.DefaultTransport = oldTransport
	})
}

func TestDoUploadFromResponse_WithoutOffsetUploadsWholeFile(t *testing.T) {
	tmpDir := setupTestEnv(t)
	filePath := filepath.Join(tmpDir, "upload.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0600); err != nil {
		t.Fatal(err)
	}

	var contentRange string
	var body string
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		contentRange = r.Header.Get("Content-Range")
		data, _ := io.ReadAll(r.Body)
		body = string(data)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	installHTTPSIntercept(t, server.Listener.Addr().String())
	t.Setenv("NW_UPLOAD_ALLOWED_HOSTS", "example.com")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("ALL_PROXY", "")

	client := api.NewClient("https://www.worksapis.com/v1.0", &auth.Token{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil)

	respBody := []byte(`{"uploadUrl":"https://example.com/upload"}`)
	if _, err := doUploadFromResponse(client, respBody, filePath); err != nil {
		t.Fatalf("doUploadFromResponse failed: %v", err)
	}

	if contentRange != "" {
		t.Fatalf("expected empty Content-Range, got %q", contentRange)
	}
	if body != "hello" {
		t.Fatalf("expected full body 'hello', got %q", body)
	}
}

func TestDriveUploadResume_UsesServerOffsetAcrossCommands(t *testing.T) {
	tests := []struct {
		name        string
		args        func(filePath string) []string
		expectedAPI string
	}{
		{
			name: "my drive",
			args: func(filePath string) []string {
				return []string{"drive", "upload", filePath, "--resume", "--user-id", "u1"}
			},
			expectedAPI: "/v1.0/users/u1/drive/files",
		},
		{
			name: "shared drive",
			args: func(filePath string) []string {
				return []string{"drive", "shared", "upload", "d1", "--file", filePath, "--resume"}
			},
			expectedAPI: "/v1.0/sharedrives/d1/files",
		},
		{
			name: "group folder",
			args: func(filePath string) []string {
				return []string{"drive", "group", "upload", "g1", "--file", filePath, "--resume"}
			},
			expectedAPI: "/v1.0/groups/g1/folder/files",
		},
		{
			name: "shared folder",
			args: func(filePath string) []string {
				return []string{"drive", "shared-folder", "upload", "sf1", "--file", filePath, "--resume", "--user-id", "u1"}
			},
			expectedAPI: "/v1.0/users/u1/drive/sharedfolders/sf1/files",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := setupTestEnv(t)
			writeTestConfig(t, tmpDir)
			filePath := filepath.Join(tmpDir, "resume.txt")
			if err := os.WriteFile(filePath, []byte("0123456789"), 0600); err != nil {
				t.Fatal(err)
			}

			var createBody string
			var contentRange string
			var uploadBody string

			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, _ := io.ReadAll(r.Body)

				switch {
				case r.Method == http.MethodPost && r.URL.Path == tc.expectedAPI:
					createBody = string(data)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"uploadUrl":"https://example.com/upload","offset":7}`))
				case r.Method == http.MethodPut && r.Host == "example.com" && r.URL.Path == "/upload":
					contentRange = r.Header.Get("Content-Range")
					uploadBody = string(data)
					w.WriteHeader(http.StatusOK)
				default:
					t.Fatalf("unexpected request: %s %s host=%s body=%s", r.Method, r.URL.Path, r.Host, string(data))
				}
			}))
			defer server.Close()

			installHTTPSIntercept(t, server.Listener.Addr().String())
			t.Setenv("NW_UPLOAD_ALLOWED_HOSTS", "example.com")
			t.Setenv("HTTPS_PROXY", "")
			t.Setenv("HTTP_PROXY", "")
			t.Setenv("ALL_PROXY", "")

			out, err := runCLI(t, tc.args(filePath)...)
			if err != nil {
				t.Fatalf("command failed: %v\nstdout: %s", err, out)
			}
			if strings.Contains(out, "uploadUrl") || strings.Contains(out, "example.com/upload") {
				t.Fatalf("upload URL should be redacted from command output: %s", out)
			}
			if !strings.Contains(out, `"uploaded": true`) && !strings.Contains(out, `"uploaded":true`) {
				t.Fatalf("expected uploaded marker in command output: %s", out)
			}

			if !strings.Contains(createBody, `"resume":true`) {
				t.Fatalf("expected resume=true in create body, got %s", createBody)
			}
			if !strings.Contains(createBody, `"fileName":"resume.txt"`) {
				t.Fatalf("expected fileName in create body, got %s", createBody)
			}
			if !strings.Contains(createBody, `"fileSize":10`) {
				t.Fatalf("expected fileSize=10 in create body, got %s", createBody)
			}
			if contentRange != "bytes 7-9/10" {
				t.Fatalf("expected Content-Range bytes 7-9/10, got %q", contentRange)
			}
			if uploadBody != "789" {
				t.Fatalf("expected resumed upload body '789', got %q", uploadBody)
			}
		})
	}
}
