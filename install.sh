#!/bin/sh
set -e

REPO="physics91/naverworks-cli"
BINARY="naverworks"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

get_os() {
  case "$(uname -s)" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *) echo "unsupported" ;;
  esac
}

get_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "unsupported" ;;
  esac
}

OS=$(get_os)
ARCH=$(get_arch)

if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
  echo "지원하지 않는 플랫폼입니다: $(uname -s) $(uname -m)" >&2
  exit 1
fi

LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "최신 버전을 찾을 수 없습니다" >&2
  exit 1
fi

VERSION="${LATEST#v}"
EXT="tar.gz"
if [ "$OS" = "windows" ]; then
  EXT="zip"
fi

FILENAME="${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${FILENAME}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

CHECKSUM_URL="https://github.com/${REPO}/releases/download/${LATEST}/checksums.txt"

echo "다운로드 중: ${URL}"
curl -sL "$URL" -o "${TMPDIR}/${FILENAME}"
curl -sL "$CHECKSUM_URL" -o "${TMPDIR}/checksums.txt"

echo "체크섬 검증 중..."
EXPECTED=$(grep "${FILENAME}" "${TMPDIR}/checksums.txt" | awk '{print $1}')
if [ -n "$EXPECTED" ]; then
  ACTUAL=$(sha256sum "${TMPDIR}/${FILENAME}" 2>/dev/null || shasum -a 256 "${TMPDIR}/${FILENAME}" 2>/dev/null | awk '{print $1}')
  ACTUAL=$(echo "$ACTUAL" | awk '{print $1}')
  if [ "$EXPECTED" != "$ACTUAL" ]; then
    echo "체크섬 불일치! 다운로드가 손상되었을 수 있습니다." >&2
    echo "  기대: $EXPECTED" >&2
    echo "  실제: $ACTUAL" >&2
    exit 1
  fi
  echo "체크섬 확인 완료"
else
  echo "체크섬을 확인할 수 없습니다: ${FILENAME}" >&2
  exit 1
fi

BIN_EXT=""
if [ "$OS" = "windows" ]; then
  BIN_EXT=".exe"
fi

echo "설치 중: ${INSTALL_DIR}/${BINARY}${BIN_EXT}"
if [ "$EXT" = "tar.gz" ]; then
  tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"
else
  unzip -q "${TMPDIR}/${FILENAME}" -d "$TMPDIR"
fi

if [ -w "$INSTALL_DIR" ]; then
  cp "${TMPDIR}/${BINARY}${BIN_EXT}" "${INSTALL_DIR}/${BINARY}${BIN_EXT}"
else
  sudo cp "${TMPDIR}/${BINARY}${BIN_EXT}" "${INSTALL_DIR}/${BINARY}${BIN_EXT}"
fi
chmod +x "${INSTALL_DIR}/${BINARY}${BIN_EXT}"

echo "설치 완료: $("${INSTALL_DIR}/${BINARY}${BIN_EXT}" version)"
