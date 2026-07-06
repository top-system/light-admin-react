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
| Database       | **PostgreSQL only**             | MySQL and SQLite are being removed. |
| DB driver      | **pgx/v5 + pgxpool**            | No `database/sql` in application code. |
| Data layer     | **sqlc** (generated)            | The single source of DB access. |
| Migrations     | **golang-migrate**              | `db/migrations/*.sql`. |
| Logging        | **zap** (`lib.Logger`)          | Never `fmt.Println` for logs. |
| Auth/RBAC      | JWT + **Casbin**                | Preserve existing behavior. |
| Config         | Viper, `config/config.yaml`     | One config source. |

### Transition state (important)

The project is migrating **module by module** from GORM to sqlc. During the
transition, **both** access layers coexist against the **same PostgreSQL**
database:

- **Migrated modules** (currently: `user`, `user_role`) use `*sqlc.Queries` +
  `lib.TxManager`.
- **Not-yet-migrated modules** still use GORM (`lib.Database`).

Migration order: `Auth → User → Role → Dept → Menu → Dict → Tenant → Config →
Monitor → Log → Job`. Migrate one module at a time; each step must
`go build ./...` and `go test ./...` clean before moving on.

---

## 2. Layering (strict)

```
route → controller → service → repository → (sqlc | GORM)
```

Dependencies point **downward only**. Never upward, never sideways within a layer.

### Controller
- **May:** bind/validate HTTP input, read JWT claims/context, call **one or more
  services**, build the `echox.Response`.
- **Must NOT:** contain SQL, business logic, transaction control, cache access,
  or touch a repository/`*sqlc.Queries`/`gorm.DB` directly.

### Service
- **Owns:** business logic, transactions, caching, permissions, events.
- **May:** call repositories and other **infra** (cache, txManager).
- **Must NOT:** contain SQL, or import Echo (`echo.Context` stays in controllers).
- **Service must NOT call another Service.** If two services need shared logic,
  extract it into a repository or a dedicated domain helper. (This prevents
  cyclic dependencies and hidden transaction nesting.)

### Repository
- **Owns:** persistence only. For migrated modules this means **calling
  `sqlc.Querier` methods and mapping rows ↔ domain models** — nothing else.
- **Must NOT:** contain business logic, cache access, or hand-written SQL strings.
- **Repository must NOT call another Repository.** Cross-entity reads/writes are
  composed in the **service** (inside a `TxManager` transaction when atomic).

### No circular dependencies
Packages must form a DAG. If you feel the need for an import cycle, the
responsibility is in the wrong layer.

---

## 3. SQL & sqlc rules

- **All SQL lives in `db/queries/*.sql`** and is compiled by `sqlc generate`.
  There must be **zero** SQL string literals (`SELECT`/`INSERT`/`UPDATE`/`DELETE`)
  in `.go` files.
- **No `SELECT *` semantics in Go** — sqlc expands columns explicitly; keep it so.
- **`ORDER BY` / `LIMIT` / `OFFSET` must be injection-safe.** Never interpolate a
  column name or direction from user input into SQL. Use fixed `ORDER BY` clauses
  or a whitelist mapped inside the query; pass limit/offset as bound parameters.
- **Dynamic filters** use the nullable-argument pattern
  (`sqlc.narg('x')::type IS NULL OR col = sqlc.narg('x')`) so one prepared
  statement serves all filter combinations while staying fully parameterized.
- After editing any `.sql`, run `make sqlc` and commit the regenerated
  `db/sqlc/*` alongside it. Never hand-edit generated files.
- Schema changes go through a **new migration** in `db/migrations` (paired
  `*.up.sql` / `*.down.sql`). Migrations are the schema source of truth for both
  the DB and sqlc.

## 4. Transactions

- Use `lib.TxManager.RunInTx(ctx, func(q *sqlc.Queries) error { ... })`.
- Inside the callback, bind each repository to the transaction with
  `repo.WithTx(q)` so every statement runs on the same transaction.
- **Do not** use `gorm.Transaction()` or pass `*gorm.DB` handles for any
  sqlc-backed module. The per-request GORM transaction middleware is legacy and
  is removed from a module when that module is migrated.

## 5. Errors

- Wrap/inspect with `errors.Is` / `errors.As`.
- Map `pgx.ErrNoRows` → domain `errors.DatabaseRecordNotFound`; wrap other DB
  errors with `errors.DatabaseInternalError`.
- Do not reference `gorm.ErrRecordNotFound` in new/migrated code.

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
  `sqlc.Querier` interface — see `api/system/repository/user_repository_test.go`).
- `go fmt`, `go vet`, `golangci-lint`, `go test`, `go build` must all pass in CI.

## 9. Directory map

```
cmd/            CLI entrypoints (runserver, migrate, setup)
api/<module>/   route / controller / service / repository per domain
db/
  migrations/   golang-migrate *.sql (schema source of truth)
  queries/      sqlc query definitions
  sqlc/         generated code (DO NOT EDIT)
lib/            infra providers (config, logger, db, pgx pool, queries, txmanager, cache)
models/         domain models & DTOs
pkg/            reusable, framework-agnostic packages
```

---

**Golden rule:** prefer stability and readability over cleverness. Write more
code rather than less if it makes the boundaries above clearer.
