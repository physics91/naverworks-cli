package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstallScript_FailsWhenChecksumMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell install script test")
	}

	tempDir := t.TempDir()
	mockBin := filepath.Join(tempDir, "bin")
	installDir := filepath.Join(tempDir, "install")
	if err := os.MkdirAll(mockBin, 0o755); err != nil {
		t.Fatalf("mkdir mock bin: %v", err)
	}
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		t.Fatalf("mkdir install dir: %v", err)
	}

	writeMock := func(name, body string) {
		path := filepath.Join(mockBin, name)
		if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	writeMock("curl", `#!/bin/sh
set -eu
out=""
url=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      out="$2"
      shift 2
      ;;
    -s|-L|-sL)
      shift
      ;;
    *)
      url="$1"
      shift
      ;;
  esac
done
if [ -z "$out" ]; then
  printf '{"tag_name":"v9.9.9"}'
  exit 0
fi
case "$url" in
  *checksums.txt)
    printf 'deadbeef  some-other-file.tar.gz\n' > "$out"
    ;;
  *)
    printf 'fake-archive' > "$out"
    ;;
esac
`)

	writeMock("tar", `#!/bin/sh
set -eu
dest=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -C)
      dest="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done
cat > "$dest/naverworks" <<'EOF'
#!/bin/sh
printf 'naverworks test version\n'
EOF
chmod +x "$dest/naverworks"
`)

	writeMock("sha256sum", `#!/bin/sh
printf 'cafebabe  %s\n' "$1"
`)

	cmd := exec.Command("sh", "install.sh")
	cmd.Dir = "."
	cmd.Env = append(os.Environ(),
		"PATH="+mockBin+":"+os.Getenv("PATH"),
		"INSTALL_DIR="+installDir,
	)

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected install.sh to fail when checksum is missing, output:\n%s", out)
	}
	if !strings.Contains(string(out), "체크섬") {
		t.Fatalf("expected checksum failure message, got:\n%s", out)
	}
}
