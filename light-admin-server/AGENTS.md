# AGENTS.md — Architecture Constraints (light-admin-server)

This file is the **binding architectural contract** for the Go backend. Every
human and AI contributor must follow it. It exists so the codebase can be
maintained and evolved for 5–10 years without decaying into a big ball of mud.

If a change conflicts with a rule here, **stop and discuss** — do not silently
work around it. When something is ambiguous, analyze and ask; do not guess.

---

## 1. Stack (fixed decisions)

| Concern        | Choice                          | Notes |
|----------------|---------------------------------|-------|
| HTTP framework | **Echo** (`labstack/echo/v4`)   | Not Gin. Do not introduce a second web framework. |
| DI             | **Uber Fx**                     | No global singletons for services/repos. |
| Database       | **PostgreSQL / MySQL / SQLite** | Selected by `config.Database.Engine`; one binary serves all three. |
| DB drivers     | pgx/v5, go-sql-driver, modernc  | Confined to `lib/datalayer.go` and the `db/*store` adapters. |
| Data layer     | **sqlc** + `db/store` façade    | Business code depends only on `store.Store`; per-engine sqlc output stays behind the generated adapters. |
| Migrations     | **golang-migrate**              | `db/migrations/{postgres,mysql,sqlite}/*.sql`, same version numbers in all three. |
| Logging        | **slog** (`lib.Logger` façade)  | Never `fmt.Println` for logs. |
| Auth/RBAC      | JWT + **Casbin**                | Preserve existing behavior. |
| Config         | Viper, `config/config.yaml`     | One config source. |

### Multi-engine invariant (important)

Every SQL statement exists in **three dialects** (`db/queries/postgres` is the
reference; `mysql`/`sqlite` are derived per docs/multi-database-plan.md §5).
The engine-neutral surface is generated: `tools/gen-store` reads the sqlc
output and emits `db/store` (interface + neutral structs) plus one adapter per
engine (`db/pgstore`, `db/mysqlstore`, `db/sqlitestore`). Methods an engine
cannot express live in that adapter's hand-written `manual.go`; the generated
`var _ store.Store = (*Store)(nil)` turns a forgotten dialect into a compile
error.

---

## 2. Layering (strict)

```
route → controller → service → repository → store.Store (engine-neutral)
```

Dependencies point **downward only**. Never upward, never sideways within a layer.

### Controller
- **May:** bind/validate HTTP input, read JWT claims/context, call **one or more
  services**, build the `echox.Response`.
- **Must NOT:** contain SQL, business logic, transaction control, cache access,
  or touch a repository/`store.Store` directly.

### Service
- **Owns:** business logic, transactions, caching, permissions, events.
- **May:** call repositories and other **infra** (cache, txManager).
- **Must NOT:** contain SQL, or import Echo (`echo.Context` stays in controllers).
- **Service must NOT call another Service.** If two services need shared logic,
  extract it into a repository or a dedicated domain helper. (This prevents
  cyclic dependencies and hidden transaction nesting.)

### Repository
- **Owns:** persistence only: **calling `store.Store` methods and mapping rows
  ↔ domain models** — nothing else.
- **Must NOT:** contain business logic, cache access, hand-written SQL strings,
  or import any database driver / generated sqlc package (depguard enforces
  this).
- **Repository must NOT call another Repository.** Cross-entity reads/writes are
  composed in the **service** (inside a `TxManager` transaction when atomic).

### No circular dependencies
Packages must form a DAG. If you feel the need for an import cycle, the
responsibility is in the wrong layer.

---

## 3. SQL & sqlc rules

- **All SQL lives in `db/queries/{postgres,mysql,sqlite}/*.sql`** and is
  compiled by `sqlc generate`. There must be **zero** SQL string literals
  (`SELECT`/`INSERT`/`UPDATE`/`DELETE`) in `.go` files.
- **Adding or changing a query touches all three dialect files** (postgres is
  the source; derive mysql/sqlite per docs/multi-database-plan.md §5), then:
  `make sqlc` → `make gen-store` → `go build ./...`. If an engine cannot
  express the statement, implement the method in that adapter's `manual.go` —
  the compiler enforces completeness.
- **No `SELECT *` semantics in Go** — sqlc expands columns explicitly; keep it so.
- **`ORDER BY` / `LIMIT` / `OFFSET` must be injection-safe.** Never interpolate a
  column name or direction from user input into SQL. Use fixed `ORDER BY` clauses
  or a whitelist mapped inside the query; pass limit/offset as bound parameters.
- **Dynamic filters** use the nullable-argument pattern
  (`sqlc.narg('x') IS NULL OR col = sqlc.narg('x')`; postgres adds `::type`
  hints) so one prepared statement serves all filter combinations while staying
  fully parameterized. In the mysql/sqlite dialects keep every parameter a
  *named* `sqlc.arg`/`sqlc.narg` (except mysql `LIMIT ? OFFSET ?`).
- **Time comes from Go**: pass `sqlc.arg('now')`; never `NOW()` /
  `CURRENT_TIMESTAMP` inside queries.
- Never hand-edit generated packages (`db/pg`, `db/mysqlgen`, `db/sqlitegen`,
  or the generated `store.go` files).
- Schema changes go through a **new migration** written in all three
  `db/migrations/<engine>` directories with the **same version number and file
  name** (paired `*.up.sql` / `*.down.sql`), following the type map in
  docs/multi-database-plan.md §4. Migrations are the schema source of truth for
  both the DB and sqlc.

## 4. Transactions

- Use `lib.TxManager.RunInTx(ctx, func(q store.Store) error { ... })`
  (`lib.TxManager` = `store.TxManager`; the engine picks pgx or database/sql
  underneath).
- Inside the callback, bind each repository to the transaction with
  `repo.WithTx(q)` so every statement runs on the same transaction.

## 5. Errors

- Wrap/inspect with `errors.Is` / `errors.As`.
- The adapters translate each driver's not-found sentinel to `store.ErrNoRows`;
  map that → domain `errors.DatabaseRecordNotFound` and wrap other DB errors
  with `errors.DatabaseInternalError`. Never reference `pgx.ErrNoRows` /
  `sql.ErrNoRows` in business code.

## 6. Multi-tenancy

- New tables that hold tenant-scoped data must carry a tenant discriminator and
  be filtered by it in every query. Do not add cross-tenant queries without an
  explicit, reviewed reason.

## 7. API compatibility

- Public HTTP contract is **frozen**: do not change JSON field names, URLs, or
  Swagger shapes during refactors. Behavior-preserving refactors only, unless a
  change is explicitly requested and versioned.

## 8. Code hygiene

- Functions target **< 50 lines**; files target **< 500 lines**. Split when
  exceeded. No God services/controllers.
- Every exported symbol has a GoDoc comment.
- New repository and service logic ships with unit tests (mock the
  `store.Store` interface — see `api/system/repository/user_repository_test.go`).
- `go fmt`, `go vet`, `golangci-lint`, `go test`, `go build` must all pass in CI.

## 9. Directory map

```
cmd/            CLI entrypoints (runserver, migrate, setup)
api/<module>/   route / controller / service / repository per domain
db/
  migrations/   golang-migrate *.sql per engine (postgres/mysql/sqlite)
  queries/      sqlc query definitions per engine (postgres is the reference)
  pg/ mysqlgen/ sqlitegen/   sqlc output per engine (DO NOT EDIT)
  store/        engine-neutral Store interface + structs (generated)
  pgstore/ mysqlstore/ sqlitestore/  adapters (generated store.go + manual.go/helpers.go)
lib/            infra providers (config, logger, datalayer, cache, ...)
models/         domain models & DTOs
tools/gen-store/  generator for db/store + the adapters
pkg/            reusable, framework-agnostic packages
```

---

**Golden rule:** prefer stability and readability over cleverness. Write more
code rather than less if it makes the boundaries above clearer.
