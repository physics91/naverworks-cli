# SCIM

SCIM은 일반 로그인 토큰이 아니라 별도 `scim_access_token` 설정을 사용합니다.

## 준비

```bash
naverworks config set scim_access_token --stdin <<< "YOUR_SCIM_TOKEN"
```

설정 확인:

```bash
naverworks config get scim_access_token
```

## 사용자 조회

```bash
naverworks scim list-users --count 100
naverworks scim get-user USER_ID
```

필터도 사용할 수 있습니다.

```bash
naverworks scim list-users --filter 'userName eq "user@example.com"'
```

## 사용자 생성/수정/삭제

`create-user`, `update-user`, `patch-user`는 `--data`에 JSON 문자열을 넣습니다.

```bash
naverworks scim create-user --data '{"userName":"user@example.com","active":true}'
naverworks scim update-user USER_ID --data '{"userName":"user@example.com","active":true}'
naverworks scim patch-user USER_ID --data '{"Operations":[{"op":"replace","path":"active","value":false}]}'
naverworks scim delete-user USER_ID
```

## 그룹 조회/수정

```bash
naverworks scim list-groups --count 100
naverworks scim get-group GROUP_ID
naverworks scim create-group --data '{"displayName":"Platform Team"}'
naverworks scim patch-group GROUP_ID --data '{"Operations":[{"op":"add","path":"members","value":[{"value":"USER_ID"}]}]}'
naverworks scim delete-group GROUP_ID
```

SCIM 호출이 실패하면 [Troubleshooting](Troubleshooting.md)에서 토큰 설정과 권한부터 다시 확인하면 됩니다.
