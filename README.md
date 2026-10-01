# GoModulith

**A lightweight, framework-agnostic modular monolith toolkit for Go.**

GoModulith takes inspiration from Spring Modulith while using idiomatic Go:
explicit module interfaces, deterministic lifecycle management, event-driven
collaboration, static import-boundary analysis and generated architecture docs.

> Active development. Core lifecycle, typed events, SQL/Mongo outbox stores, retryable dispatch,
> optional HTTP/gRPC integrations and OTel metrics are available. Integrations must be validated
> against your application's driver/framework versions before production deployment.

## Install

```sh
go get github.com/el211/GoModulith
go install github.com/el211/GoModulith/cmd/gomodulith@latest
```

Requires Go 1.23+. No mandatory external dependencies.

## Define a module

```go
package orders

import (
    "context"
    "github.com/el211/GoModulith/module"
)

func New() module.Definition {
    return module.Definition{
        ID:       "orders",
        Requires: []string{"users"},
        Metadata: module.Info{
            ImportPath: "example.com/shop/modules/orders",
            Description: "Order processing and public order contracts",
            Version: "0.1.0",
            Owners: []string{"commerce-team"},
            Exports: []string{"OrderService", "OrderPlaced"},
        },
        OnStart: func(ctx context.Context) error { return nil },
        OnStop:  func(ctx context.Context) error { return nil },
    }
}
```

## Assemble an application

```go
app := modulith.New(modulith.WithName("shop"))
if err := app.Register(users.New(), orders.New(), payments.New()); err != nil { log.Fatal(err) }
if err := app.Verify(); err != nil { log.Fatal(err) }
if err := app.Run(ctx); err != nil { log.Fatal(err) }
defer app.Shutdown(context.Background())
```

Modules start after their declared dependencies and stop in reverse order.
Cycles, duplicate names and missing dependencies are rejected.

## Typed event bus

```go
type OrderPlaced struct { ID string }
bus := events.New()
unsubscribe, err := events.SubscribeTyped(bus, func(ctx context.Context, e OrderPlaced) error {
    fmt.Println(e.ID)
    return nil
})
if err != nil { log.Fatal(err) }
defer unsubscribe()
_ = bus.Publish(ctx, OrderPlaced{ID: "42"})
```

Events are synchronous by default. `PublishAsync` returns a result channel.
`durable.Store` and `durable.Dispatcher` define an at-least-once outbox port;
`durable/memory` is only a test/development implementation, not crash-safe.

## Source-layout convention and package boundaries

```text
myapp/
  go.mod
  modules/
    users/
      doc.go           # package users: purpose and public API
      api.go           # exported contracts
      internal/        # private implementation
    orders/
      doc.go
      api.go
      internal/
    payments/
      doc.go
      api.go
      internal/
```

`architecture.Analyze` finds direct subdirectories of `modules/`, scans
their Go imports (including test files), builds a dependency graph, detects
cycles, forbids imports into another module's `internal/` directory, and
optionally checks per-module dependency allowlists. Note: this is source-level
convention enforcement, not a Go compiler extension; reflection and other
runtime coupling cannot be statically ruled out.

```sh
gomodulith verify -root . -dir modules
gomodulith inspect -root . -dir modules
gomodulith graph -root . -dir modules > architecture.mmd
gomodulith docs -root . -dir modules > ARCHITECTURE.md
```

`inspect` emits JSON, `graph` emits Mermaid, and `docs` emits Markdown with
a Mermaid diagram. CLI exits nonzero for validation failures.

## Production integrations

- `durable/sqloutbox`: call `Schema(sqloutbox.SQLite)` or `Schema(sqloutbox.Postgres)` in migrations; use `EnqueueTx(ctx, tx, event)` inside the same database transaction as business writes. Install and register your preferred database/sql driver yourself.
- `durable/mongooutbox`: run `EnsureIndexes`; use `EnqueueSession(sessionCtx, event)` in a MongoDB transaction for business atomicity. Requires replica set or sharded topology for transactions.
- `durable/reliable`: dispatch via claim leases, exponential retries and a persistent dead-letter state; consumers must deduplicate event IDs. Each dispatcher drain uses a fresh claim-owner token to prevent a stale drain from acknowledging a later claim by the same worker.
- `integrations/gin`, `integrations/fiber`, `integrations/echo`, `integrations/grpc`: optional context middleware and interceptors.
- `observability/otel`: optional metrics-backed Observer. Attach an Observer from your application instrumentation layer.
- `architecture.AnalyzeIncremental`: content-addressed source-file import cache. Additional rules: `NoDependency`, `LayerRule`, `NamingRule`, `MaxDependencies`, `RequiredDependency`; verify using `VerifyRules`.

## Package information

| Package | Responsibility |
|---|---|
| `modulith` (root) | Application registry, topological ordering, lifecycle and rollback |
| `module` | Lightweight function-backed module definition |
| `events` | Concurrent-safe synchronous/asynchronous typed event dispatch |
| `durable` | Event envelope, outbox ports and at-least-once dispatcher |
| `durable/memory` | In-memory Store for development/testing (non-durable) |
| `durable/reliable` | Leases, bounded delivery, retry/backoff and dead letters |
| `durable/sqloutbox` | Transactional database/sql outbox for SQLite/PostgreSQL |
| `durable/mongooutbox` | MongoDB session-bound enqueue and atomic lease claim |
| `integrations/gin`, `fiber`, `echo`, `grpc` | Optional framework middleware/interceptors |
| `observability/otel` | OpenTelemetry metrics Observer |
| `architecture` | Source import graph, boundary/allowlist/cycle verification |
| `documenter` | Mermaid/Markdown architecture export |
| `observability` | Optional instrumentation hooks |
| `testing` | Module-isolation harness and dependency stubs |
| `cmd/gomodulith` | CLI: verify, inspect, graph, docs |
| `examples/shop` | Runnable minimal application |

Every library package contains package-level GoDoc in its root file or
`doc.go`; public APIs carry GoDoc comments. The CLI and example are
`package main` because they are executables.

### Transactional outbox usage

Call `sqloutbox.Schema(dialect)` in a migration, then invoke `EnqueueTx` with the *same* `*sql.Tx` as the business update. Do not use standalone `Enqueue` when business-write atomicity is required. For MongoDB, invoke `EnqueueSession` within the `mongo.SessionContext` of a running transaction; MongoDB multi-document transactions require a supported replica set or sharded deployment. Outbox delivery is at least once; use event IDs for downstream idempotency. Lease duration must exceed publication time or be renewed externally to avoid redelivery during a long publish.

## Verification

```sh
gofmt -w .
go vet ./...
go test -race ./...
go build ./...
go run ./examples/shop
```

## Roadmap

- [x] Module registration, lifecycle, dependency ordering and rollback
- [x] Type-aware in-process event dispatch and async result channels
- [x] Import boundary and circular dependency analyzer
- [x] Dependency allowlists, CLI, JSON and Mermaid/Markdown documentation
- [x] Outbox interfaces and volatile reference store
- [x] Package GoDoc, tests and CI workflow
- [x] SQL transactional outbox (SQLite / PostgreSQL)
- [x] MongoDB adapter with transaction-bound enqueue
- [x] Framework integrations: Gin, Fiber, Echo, gRPC
- [x] OpenTelemetry adapter
- [x] Event retry/backoff, leases, dead-letter handling
- [x] Incremental package graph analysis and richer architecture rules

MIT licensed.
