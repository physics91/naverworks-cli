# GitHub Wiki Publishing

`docs/wiki/`는 `naverworks-cli` GitHub wiki의 원본 문서 디렉터리입니다.

## 왜 바로 푸시가 안 되는가

GitHub Docs 기준으로, 위키는 GitHub 웹 UI에서 초기 페이지를 한 번 만든 뒤에야
`.wiki.git` 저장소를 clone/push 할 수 있습니다.

참고:
- https://docs.github.com/en/communities/documenting-your-project-with-wikis/adding-or-editing-wiki-pages

문서의 핵심 문구:
- "Once you've created an initial page on GitHub, you can clone the repository..."

즉, 위키 기능이 켜져 있어도 첫 페이지가 없으면
`https://github.com/physics91/naverworks-cli.wiki.git`가 404를 반환할 수 있습니다.

## 첫 퍼블리시 순서

1. GitHub 웹에서 `https://github.com/physics91/naverworks-cli/wiki`로 이동
2. `New Page`로 초기 페이지를 하나 생성
   - 페이지명은 `Home` 권장
   - 내용은 임시 한 줄이어도 됨
3. 로컬에서 wiki 저장소 clone

```bash
git clone https://github.com/physics91/naverworks-cli.wiki.git /tmp/naverworks-cli.wiki
```

4. 원본 문서 복사

```bash
cp docs/wiki/* /tmp/naverworks-cli.wiki/
```

5. wiki 저장소에서 커밋 후 푸시

```bash
cd /tmp/naverworks-cli.wiki
git add .
git commit -m "docs(wiki): 사용자 가이드 위키 초안 추가"
git push origin master
```

## 포함할 파일

아래 파일들이 GitHub wiki 루트에 그대로 들어가야 합니다.

- `Home.md`
- `Installation.md`
- `Quick-Start.md`
- `Authentication-and-Profiles.md`
- `Configuration-Keys-and-Environment-Variables.md`
- `Output-and-Pagination.md`
- `Domain-Command-Guide.md`
- `SCIM.md`
- `Troubleshooting.md`
- `_Sidebar.md`
