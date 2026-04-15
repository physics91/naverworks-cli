# 크로스 빌드 — goreleaser 없이 수동 실행

`goreleaser`가 없을 때 Phase 2b에서 사용하는 플랫폼별 빌드 + 아카이브 명령. goreleaser/deploy 산출물 형식과 일치시킨다.

## 사전 변수

```bash
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
mkdir -p dist
```

`LDFLAGS`는 Phase 2b 시작 시 아래와 같이 구성한다:

```
-s -w -X github.com/physics91/naverworks-cli/cmd.version=$VERSION \
      -X github.com/physics91/naverworks-cli/cmd.commit=$COMMIT \
      -X github.com/physics91/naverworks-cli/cmd.buildDate=$DATE
```

## 플랫폼별 빌드 + 아카이브

### linux-amd64

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/naverworks .
tar -czf "dist/naverworks_${VERSION}_linux_amd64.tar.gz" -C dist naverworks
rm dist/naverworks
```

### linux-arm64

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/naverworks .
tar -czf "dist/naverworks_${VERSION}_linux_arm64.tar.gz" -C dist naverworks
rm dist/naverworks
```

### darwin-amd64

```bash
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/naverworks .
tar -czf "dist/naverworks_${VERSION}_darwin_amd64.tar.gz" -C dist naverworks
rm dist/naverworks
```

### darwin-arm64

```bash
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/naverworks .
tar -czf "dist/naverworks_${VERSION}_darwin_arm64.tar.gz" -C dist naverworks
rm dist/naverworks
```

### windows-amd64

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o dist/naverworks.exe .
zip dist/naverworks_${VERSION}_windows_amd64.zip -j dist/naverworks.exe
rm dist/naverworks.exe
```
