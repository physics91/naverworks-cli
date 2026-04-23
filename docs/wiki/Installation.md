# Installation

## npm

가장 일반적인 설치 방법입니다.

```bash
npm install -g naverworks
```

설치 후 바로 버전을 확인합니다.

```bash
naverworks version
```

## bun

Bun으로도 전역 설치할 수 있습니다.

```bash
bun add -g naverworks
naverworks version
```

현재 전역 명령은 Node 호환 런처를 사용합니다.
Node 없이 Bun만 쓰는 환경이면 아래 `bunx --bun` 또는 설치 스크립트를 권장합니다.

## npx

전역 설치 없이 한 번 실행할 때 편합니다.

```bash
npx naverworks version
```

## bunx

Bun으로 한 번 실행할 때는 `bunx`를 사용하면 됩니다.

```bash
bunx --bun naverworks version
```

## 설치 스크립트

GitHub Releases에서 현재 플랫폼에 맞는 바이너리를 받아 설치합니다.

```bash
curl -sSL https://raw.githubusercontent.com/physics91/naverworks-cli/main/install.sh | sh
```

- 기본 설치 경로: `/usr/local/bin`
- 설치 경로 변경: `INSTALL_DIR=/your/path`

예시:

```bash
curl -sSL https://raw.githubusercontent.com/physics91/naverworks-cli/main/install.sh | INSTALL_DIR="$HOME/.local/bin" sh
```

## Go install

Go 환경에서 직접 설치할 수도 있습니다.

```bash
go install github.com/physics91/naverworks-cli@latest
```

## GitHub Releases 바이너리

[Releases](https://github.com/physics91/naverworks-cli/releases)에서 바이너리를 직접 내려받을 수 있습니다.

| 플랫폼 | 아키텍처 |
| --- | --- |
| Linux | amd64, arm64 |
| macOS | amd64, arm64 |
| Windows | amd64 |

설치가 끝나면 [Quick Start](Quick-Start.md)로 넘어가서 바로 인증을 진행하면 됩니다.
