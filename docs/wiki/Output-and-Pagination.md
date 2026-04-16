# Output and Pagination

## 출력 형식

기본 출력은 pretty JSON입니다.

```bash
naverworks directory list-users
```

일부 목록형 명령은 테이블 출력도 지원합니다.

```bash
naverworks directory list-users --output table
```

테이블 출력은 빠르게 눈으로 확인할 때 편하고, JSON 출력은 파이프/스크립트에 붙이기 좋습니다.

## 페이지네이션

목록형 명령은 아래 플래그를 자주 씁니다.

- `--count`: 페이지 크기
- `--cursor`: 다음 페이지 커서
- `--all`: 가능한 페이지를 전부 순회

예시:

```bash
# 첫 페이지
naverworks directory list-users --count 10

# 다음 페이지
naverworks directory list-users --cursor "NEXT_CURSOR"

# 가능한 페이지 전부 자동 순회
naverworks directory list-users --all
```

다른 목록형 커맨드도 거의 같은 패턴으로 쓸 수 있습니다.

```bash
naverworks task list --user-id me --all
naverworks mail list FOLDER_ID --count 50
naverworks scim list-users --count 100
```
