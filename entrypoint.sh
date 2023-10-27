wait-for "${PG_HOST}:${PG_PORT}" -- "$@"
go build -o main main.go
./main 