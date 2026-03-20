#!/bin/bash
set -e

VERSION="${1:?버전을 지정하세요 (예: 0.1.0)}"
DIST_DIR="${2:?dist 디렉토리를 지정하세요}"
NPM_DIR="$(cd "$(dirname "$0")" && pwd)"

PLATFORMS=(
  "linux-x64:linux:amd64"
  "linux-arm64:linux:arm64"
  "darwin-x64:darwin:amd64"
  "darwin-arm64:darwin:arm64"
  "win32-x64:windows:amd64"
)

# 메인 패키지 버전 업데이트
cd "$NPM_DIR/cli"
node -e "
  const pkg = require('./package.json');
  pkg.version = '${VERSION}';
  for (const dep of Object.keys(pkg.optionalDependencies || {})) {
    pkg.optionalDependencies[dep] = '${VERSION}';
  }
  require('fs').writeFileSync('package.json', JSON.stringify(pkg, null, 2) + '\n');
"

# 플랫폼별 패키지 생성
for entry in "${PLATFORMS[@]}"; do
  IFS=":" read -r NPM_PLATFORM GOOS GOARCH <<< "$entry"
  PKG_DIR="$NPM_DIR/$NPM_PLATFORM"

  EXT=""
  if [ "$GOOS" = "windows" ]; then
    EXT=".exe"
  fi

  ARCHIVE_NAME="nw-cli_${VERSION}_${GOOS}_${GOARCH}"
  ARCHIVE_PATH=""

  if [ "$GOOS" = "windows" ]; then
    ARCHIVE_PATH="$DIST_DIR/${ARCHIVE_NAME}.zip"
  else
    ARCHIVE_PATH="$DIST_DIR/${ARCHIVE_NAME}.tar.gz"
  fi

  if [ ! -f "$ARCHIVE_PATH" ]; then
    echo "경고: $ARCHIVE_PATH 를 찾을 수 없습니다. 건너뜁니다."
    continue
  fi

  # package.json 생성
  cat > "$PKG_DIR/package.json" << EOF
{
  "name": "@nw-cli/${NPM_PLATFORM}",
  "version": "${VERSION}",
  "description": "nw-cli binary for ${NPM_PLATFORM}",
  "os": ["${GOOS/linux/linux}"],
  "cpu": ["${GOARCH/amd64/x64}"],
  "license": "MIT",
  "repository": {
    "type": "git",
    "url": "https://github.com/physics91/naverworks-cli"
  }
}
EOF

  # os 필드 보정
  node -e "
    const pkg = require('${PKG_DIR}/package.json');
    const osMap = { linux: 'linux', darwin: 'darwin', windows: 'win32' };
    const cpuMap = { amd64: 'x64', arm64: 'arm64' };
    pkg.os = [osMap['${GOOS}']];
    pkg.cpu = [cpuMap['${GOARCH}']];
    require('fs').writeFileSync('${PKG_DIR}/package.json', JSON.stringify(pkg, null, 2) + '\n');
  "

  # 바이너리 추출
  TMPDIR=$(mktemp -d)
  if [ "$GOOS" = "windows" ]; then
    unzip -q "$ARCHIVE_PATH" -d "$TMPDIR"
  else
    tar -xzf "$ARCHIVE_PATH" -C "$TMPDIR"
  fi
  cp "$TMPDIR/nw-cli${EXT}" "$PKG_DIR/nw-cli${EXT}"
  chmod +x "$PKG_DIR/nw-cli${EXT}"
  rm -rf "$TMPDIR"

  echo "준비 완료: @nw-cli/${NPM_PLATFORM}@${VERSION}"
done

echo ""
echo "npm 퍼블리시 명령어:"
echo ""
for entry in "${PLATFORMS[@]}"; do
  IFS=":" read -r NPM_PLATFORM _ _ <<< "$entry"
  if [ -f "$NPM_DIR/$NPM_PLATFORM/nw-cli" ] || [ -f "$NPM_DIR/$NPM_PLATFORM/nw-cli.exe" ]; then
    echo "  cd $NPM_DIR/$NPM_PLATFORM && npm publish --access public"
  fi
done
echo "  cd $NPM_DIR/cli && npm publish --access public"
