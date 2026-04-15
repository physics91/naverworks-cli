# npm Trusted Publishing (OIDC) 트러블슈팅

이 저장소의 npm publish는 장기 토큰이 아닌 **OIDC Trusted Publishing** 방식을 사용한다. 관련 설정은 `.github/workflows/release.yml`의 `permissions.id-token: write`, Node 24, `--provenance` 플래그, 그리고 각 npm 패키지의 Trusted Publisher 등록이다.

## 증상별 원인

| 증상 | 의미 | 대응 |
|------|------|------|
| `ENEEDAUTH need auth` | OIDC 교환이 silent 실패 후 fallback | verbose 로그로 진짜 원인 확인 |
| `OIDC token exchange error - package not found` | 해당 패키지에 Trusted Publisher 미등록 또는 claim 불일치 | `npm trust list <pkg>`로 실제 설정 확인 |
| goreleaser `already_exists` 422 | 같은 tag의 GitHub Release가 이미 존재 | `gh release delete <tag>` 후 재시도 |
| `npm publish` 시 `NODE_AUTH_TOKEN: XXXXX-XXXXX` placeholder 주입 | setup-node가 registry-url 설정 시 자동 주입 — OIDC 경로 차단 가능 | `.npmrc`의 `_authToken` 라인 제거 또는 `NPM_CONFIG_PROVENANCE=true` + `--provenance`로 OIDC 강제 트리거 |

## 디버깅 단계

1. **verbose 로그로 실제 실패 메시지 확인**:
   워크플로 publish 스텝에 `--loglevel=verbose`를 붙여 재실행. npm verbose 로그의 `oidc` 키워드로 진짜 원인이 드러난다.

2. **서버 측 Trusted Publisher 설정 조회** (로컬에서):
   ```bash
   npm login                                   # 세션 기반 재로그인 (2FA 필수)
   npm trust list <package>                    # 서버에 저장된 publisher 확인
   ```
   전제: npm 11.10.0+, 계정 2FA 활성화. Granular access token으로는 403.

3. **웹 UI 대신 CLI로 직접 등록** (npm 웹 UI Save가 silent 실패하는 케이스 존재):
   ```bash
   npm trust github <package> \
     --repo physics91/naverworks-cli \
     --file release.yml --yes
   ```

4. **OIDC JWT claim 직접 확인** (workflow 내부):
   ```bash
   TOKEN=$(curl -sS \
     -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" \
     "${ACTIONS_ID_TOKEN_REQUEST_URL}&audience=npm:registry.npmjs.org" \
     | jq -r '.value')
   echo "$TOKEN" | cut -d'.' -f2 | base64 -d 2>/dev/null | jq '{
     repository, repository_owner, workflow_ref, environment
   }'
   ```
   이 claim이 npm Trusted Publisher 설정과 100% 일치해야 한다.

## 필수 전제 조건 체크리스트

- [ ] workflow에 `permissions.id-token: write`
- [ ] `actions/setup-node` node-version `"24"` (22.14 미만은 OIDC 미지원)
- [ ] 워크플로에 `npm install -g npm@latest` (Node 번들 npm은 11.5.1 미만일 수 있음)
- [ ] publish 시 `--provenance` 플래그 + `NPM_CONFIG_PROVENANCE: "true"`
- [ ] 6개 npm 패키지 각각에 Trusted Publisher 등록
    (Organization `physics91`, Repository `naverworks-cli`, Workflow filename `release.yml`, Environment 비어 있음)
- [ ] GitHub Actions가 GitHub-hosted runner에서 실행 (self-hosted 미지원)
