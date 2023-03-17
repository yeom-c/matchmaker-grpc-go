## Matchmaker Server gRPC Go

### 소개
> grpc bidirectional streaming 을 활용한 양방향 통신 golang 매칭서버 저장소입니다.
- [Go gRPC](https://grpc.io/docs/languages/go/quickstart)
- [SQLC](https://github.com/kyleconroy/sqlc)

### DB 마이그레이션
- db/migration/battle 에 있는 schema.sql 로 battle db에 테이블들을 생성한다.
- db/migration/game 에 있는 schema.sql 로 game_0, game_1 db에 테이블들을 생성한다.

### SQLC generate
- sqlc 설치: [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
1. db/query 에 테이블에 맞는 sql 파일을 생성한다.
2. [docs](https://docs.sqlc.dev/en/latest/index.html#) 를 참고해서 generate 를 위한 쿼리를 작성한다.
3. 아래 generate 명령어를 통해 db/sqlc 경로에 sql.go 파일로 변환한다.
```shell
# sql generate
$ make gen.sqlc
```

### Protobuf Submodule
- protobuf 사용하기 위해 git submodule 을 설정한다.
```shell
$ make reset.protobuf
$ make add.protobuf
```

### 시작하기
```bash
# .env.example 파일을 .env 파일로 복사한다.
# Makefile.example 파일을 Makefile 파일로 복사한다.
# 복사한 파일들을 자신의 환경에 맞게 수정한다.

# 실행
$ go run application.go
# or air 가 설치되어 있다면
$ air
```
