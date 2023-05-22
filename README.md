# two-phase-commit

## Setup instruction
1. To generate stubs from proto files. `cd contracts` and `buf generate proto`
2. Setup database `docker run -p 13454:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres  postgres:12`
3. Setup schema and data using `goose -dir ./db/migrations postgres 'postgres://postgres:postgres@localhost:13454/postgres?sslmode=disable' up`
4. Run delivery service `go run delivery/main.go`
5. Run store service `go run store/main.go`
6. Run order service `go run order/main.go`

## Overall flow
```mermaid
sequenceDiagram
    autonumber
    actor u as user
    participant o as order
    participant s as store
    participant d as delivery 
    u -->> o: place an order
    rect rgb(191, 223, 255)
    note right of o: Reserve/plan phase
    o ->> s: reserve food
    note over s: start timer
    o ->> d: reserve a delivery agent
    note over d: start timer
    end
    rect rgb(200, 150, 255)
    note right of o: commit phase
    o ->> s: order the food
    o ->> d: assign the delivery agent
    end
    o -->> u: order confirmed 
```
1. **Reserve phase**: Reserve all the required resources with a timer. If the timer expires, release the resource.
2. **Commit phase**: Book/commit the reserved resource. The probability of failure during the commit phase should be minimal, ideally occurring only in exceptional scenarios such as network failure, store , or delivery service disruption. In the event of such failure, the commit phase can be retried.

## Reference
1. [Distributed Transactions: Two-Phase Commit Protocol by Arpit Bhayani](https://www.youtube.com/watch?v=7FgU1D4EnpQ)
