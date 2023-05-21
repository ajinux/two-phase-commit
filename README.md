# two-phase-commit

## Setup instruction
1. To generate stubs from proto files. `cd contracts` and `buf generate proto`
2. Setup database `docker run -p 13454:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres  postgres:12`
3. Setup schema and data using `goose -dir ./db/migrations postgres 'postgres://postgres:postgres@localhost:13454/postgres?sslmode=disable' up`
4. Run delivery service `go run delivery/main.go`
5. Run store service `go run store/main.go`
6. Run order service `go run order/main.go`