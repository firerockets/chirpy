# chirpy

## Dependencies
```
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

go get github.com/lib/pq
go get github.com/joho/godotenv

go get -u github.com/golang-jwt/jwt/v5
```

## Installing
We need to have Postgresql installed.
Create a database called chirpy
```
CREATE DATABASE chirpy;
```
