# 多数据库支持方案(PostgreSQL / MySQL / SQLite)

> 状态:方案评审中,未开工。
> 前置:GORM → sqlc/pgx 迁移已完成(fb005ef)。
> 目标读者:维护者与贡献者。落地后本文档的「方言改写规则」「新增 SQL 流程」两节会沉淀为贡献者指南。

## 1. 背景与目标

当前数据层为 sqlc(postgresql 引擎)+ pgx/v5,PostgreSQL 独占。作为开源项目,
用户应能通过 `config.Database.Engine` 自由选择 **postgres / mysql / sqlite** 三种数据库。

**目标**

- 一份业务代码(service / controller / route 层零改动)跑三种数据库;
- 保留 sqlc 的编译期类型安全,不回退 ORM / 运行时反射;
- 三种引擎共享同一套 repository 测试,CI 三引擎全绿。

**非目标**

- 不支持运行时切换引擎(启动时决定);
- 不追求三种引擎性能一致(SQLite 写并发天然受限);
- 不引入跨库的在线数据迁移工具(换库 = 重新 migrate + 自行导数据)。

## 2. 总体架构

核心思路:**按引擎各生成一套 sqlc 代码,用手写的引擎中立 `Store` 接口 + 薄适配层统一,
启动时按配置装配。**

```
                    ┌────────────────────────────┐
 service/repository │   db/store.Store (接口)     │  ← 全部业务代码只依赖这一层
                    │   db/store.TxManager (接口) │
                    └──────────┬─────────────────┘
                               │ 按 config.Database.Engine 装配(FX 工厂)
          ┌────────────────────┼────────────────────┐
   ┌──────┴──────┐      ┌──────┴──────┐      ┌──────┴──────┐
   │ pgstore     │      │ mysqlstore  │      │ sqlitestore │   ← 薄适配层(可脚本生成)
   │ (pgx/v5)    │      │(database/sql)│     │(database/sql)│
   └──────┬──────┘      └──────┴──────┘      └──────┬──────┘
   ┌──────┴──────┐      ┌──────┴──────┐      ┌──────┴──────┐
   │ db/pg       │      │ db/mysqlgen │      │ db/sqlitegen│   ← sqlc 生成
   └─────────────┘      └─────────────┘      └─────────────┘
```

### 2.1 目录结构

```
db/
  store/                     # 手写:引擎中立层
    store.go                 #   Store 接口(方法集 = 现 Querier,~140 方法)
    models.go                #   引擎中立的行/参数结构体(从生成代码收敛而来)
    tx.go                    #   TxManager 接口
  pgstore/                   # PG 适配层:实现 store.Store,内部委托 db/pg
  mysqlstore/                # MySQL 适配层
  sqlitestore/               # SQLite 适配层
  migrations/
    postgres/                # 现 db/migrations 平移
    mysql/
    sqlite/
  queries/
    postgres/                # 现 db/queries 平移
    mysql/
    sqlite/
  pg/                        # sqlc 生成(现 db/sqlc 平移改名)
  mysqlgen/                  # sqlc 生成
  sqlitegen/                 # sqlc 生成
  embed.go                   # 三套迁移分别 embed,按引擎取
  migrate.go                 # golang-migrate,按引擎选 driver + 目录
```

### 2.2 sqlc.yaml(三引擎)

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "db/migrations/postgres"
    queries: "db/queries/postgres"
    gen:
      go:
        package: "pg"
        out: "db/pg"
        sql_package: "pgx/v5"
        emit_interface: true
        emit_pointers_for_null_types: true
        emit_empty_slices: true
  - engine: "mysql"
    schema: "db/migrations/mysql"
    queries: "db/queries/mysql"
    gen:
      go:
        package: "mysqlgen"
        out: "db/mysqlgen"
        sql_package: "database/sql"
        emit_interface: true
        emit_pointers_for_null_types: true
        emit_empty_slices: true
  - engine: "sqlite"
    schema: "db/migrations/sqlite"
    queries: "db/queries/sqlite"
    gen:
      go:
        package: "sqlitegen"
        out: "db/sqlitegen"
        sql_package: "database/sql"
        emit_interface: true
        emit_pointers_for_null_types: true
        emit_empty_slices: true
```

必要时用每引擎 `overrides` 把列类型钉到相同的 Go 类型(见 §4)。

### 2.3 驱动选型

| 引擎 | 运行时驱动 | golang-migrate driver | 最低版本要求 |
|---|---|---|---|
| PostgreSQL | pgx/v5(现状) | `database/pgx/v5`(现状) | PG 13+ |
| MySQL | `go-sql-driver/mysql`(DSN 必须带 `parseTime=true&loc=Asia%2FShanghai`) | `database/mysql` | MySQL 8.0+ |
| SQLite | `modernc.org/sqlite`(纯 Go,无 CGO,交叉编译友好) | `database/sqlite` | 内嵌(≥3.35,自带 RETURNING) |

SQLite 连接参数(在 DSN 构造里固定):`_pragma=journal_mode(WAL)`、
`_pragma=busy_timeout(5000)`、`_pragma=foreign_keys(1)`;连接池 `MaxOpenConns=1`
规避写锁竞争(读多场景可放开为「1 写池 + N 读池」,首版不做)。

## 3. Store 接口层设计

### 3.1 接口与中立类型

`store.Store` 的方法集就是现在 sqlc 生成的 `Querier`(~140 个方法),差别只在
类型引用从 `sqlc.*` 换成 `store.*`:

```go
package store

type Store interface {
    GetUser(ctx context.Context, id string) (User, error)
    ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error)
    // ... 与现 Querier 一一对应
}
```

`store/models.go` 的结构体从现生成代码原样搬来(字段名、Go 类型完全一致)。
之后 `db/sqlc` 不再被业务代码引用。

### 3.2 适配层 ≈ 免费

三个引擎的生成结构体通过 `overrides` 收敛到与 `store` 完全相同的字段名和 Go 类型后,
Go 的结构体类型转换(字段相同即可转)让适配方法都是一行:

```go
package pgstore

func (s *Store) GetUser(ctx context.Context, id string) (store.User, error) {
    row, err := s.q.GetUser(ctx, id)
    return store.User(row), err
}
```

~140 个方法 × 3 引擎的适配代码用一个小生成器(`tools/gen-store/`,读 Querier 的
AST 吐适配文件)维护,**不手写**。字段收敛不了的个别类型(见 §4)在适配层手工转换。

### 3.3 事务

```go
package store

type TxManager interface {
    RunInTx(ctx context.Context, fn func(Store) error) error
}
```

- PG 实现:现有 pgx 事务代码平移;
- MySQL/SQLite 实现:`database/sql` 的 `BeginTx/Commit/Rollback`,回调里传
  `sql.Tx` 绑定的生成 Queries 再包一层适配。

现有 22 处 `RunInTx` 调用点只改回调参数类型(`*sqlc.Queries` → `store.Store`)。

### 3.4 FX 装配

`lib.Module` 中 `NewPgxPool / NewQueries / NewTxManager` 收敛为按引擎分发的工厂:

```go
func NewDataLayer(lc fx.Lifecycle, cfg Config, logger Logger) (store.Store, store.TxManager)
```

内部 switch `cfg.Database.Engine`(配置字段与 `IsPostgreSQL/IsMySQL/IsSqlite`
判断方法 GORM 时代已存在,复用)。`pkg/queue` 的 `TaskRepository` 同样经由
`store.Store`,自动获得三引擎支持。

## 4. 类型映射与 DDL 改写

现有 DDL 很克制(CHAR/VARCHAR/INTEGER + 应用层生成的 CHAR(32) 主键),需要处理的差异:

| PostgreSQL(现状) | MySQL | SQLite | Go 侧(收敛目标) |
|---|---|---|---|
| `TIMESTAMPTZ` | `DATETIME(6)` | `DATETIME`(文本存储) | `time.Time` |
| `JSONB` | `JSON` | `TEXT` | `[]byte`(需 override 统一) |
| `BIGSERIAL`(仅 sys_tasks) | `BIGINT AUTO_INCREMENT` | `INTEGER PRIMARY KEY AUTOINCREMENT` | `int64` |
| `DEFAULT NOW()` | `DEFAULT CURRENT_TIMESTAMP(6)` | `DEFAULT CURRENT_TIMESTAMP` | —(见 §5,尽量移到 Go 传参) |
| `CREATE INDEX IF NOT EXISTS` | **不支持 IF NOT EXISTS**,直接 `CREATE INDEX`(迁移有版本号保证幂等) | 支持 | — |
| `BOOLEAN` | `TINYINT(1)` | `INTEGER` | `bool` |

时区策略:PG 现状 pin `Asia/Shanghai`;MySQL 在 DSN `loc` 参数对齐;SQLite 统一存
UTC 文本、Go 侧转换。**三引擎的时间语义以「Go 侧 time.Time 正确」为验收标准,测试矩阵覆盖。**

## 5. 查询方言改写规则

`db/queries/postgres` 为源本,MySQL/SQLite 版按以下规则派生(17 个文件,~157 处):

| 模式 | PostgreSQL(现状) | MySQL / SQLite 改法 |
|---|---|---|
| 占位符 | `$1` | `?`(sqlc 按引擎自动生成,查询文件里直接写 `?`) |
| narg 类型标注 | `sqlc.narg('x')::text IS NULL` | 删掉 `::text`(MySQL/SQLite 不需要参数类型提示) |
| 数组参数 | `x = ANY(sqlc.narg('ids')::text[])` | `x IN (sqlc.slice('ids'))`(sqlc 对 mysql/sqlite 引擎原生支持) |
| 模糊匹配 | `ILIKE` | `LIKE`(MySQL 用 `utf8mb4_0900_ai_ci` 排序规则天然不分大小写;SQLite ASCII 默认不分大小写,中文场景无影响) |
| 时间函数 | `NOW()`(53 处) | **阶段一统一移到 Go 传参**,三种方言都不再含时间函数 |
| `RETURNING *` | 原生(~20 处,集中在 queue/download/task) | SQLite ≥3.35 原生支持;MySQL 改 `:execresult` + `LastInsertId` 回读(相关表主键恰为自增 bigint),适配层抹平差异 |

**注意:`sqlc.slice()` 不支持 postgresql 引擎,`ANY(数组)` 不支持 mysql/sqlite——
这是查询文件无法完全共享、必须一式三份的根本原因。**

派生流程工具化:`tools/gen-queries/`(或一次性脚本)从 postgres 版做机械替换生成
初稿,人工复核 diff 后入库;之后的日常维护直接改三份(见 §8 贡献者流程)。

## 6. 迁移(golang-migrate)

- `db/embed.go` 分别 embed 三个迁移目录;`db/migrate.go` 按引擎注册对应 driver
  并选目录(pgx5 URL / mysql DSN / sqlite 文件路径);
- 11 对现有迁移按 §4 规则各改写一份;三个目录的**版本号与文件名保持一致**,
  内容允许方言差异;
- `cmd migrate` 不变,内部按 `Engine` 分发。

## 7. 测试矩阵

- 现有 repository 测试(dbtest)参数化:`DBTEST_ENGINE=postgres|mysql|sqlite`;
- 本地默认 **SQLite**(临时文件,零依赖,`go test ./...` 开箱即跑——这是本方案的
  附带收益:测试套件不再强依赖外部数据库);
- PG 沿用现有 WSL 实例;MySQL 本地可选 docker;
- CI:三个 job(sqlite 直接跑;postgres/mysql 用 service container);
- 每引擎各加一条端到端冒烟:migrate → setup → 登录 → 用户 CRUD → 队列任务生命周期。

## 8. 长期维护约定(落地后写入 CONTRIBUTING)

新增/修改一条 SQL 的流程:

1. 在 `db/queries/{postgres,mysql,sqlite}` 三处同步修改(通常只有占位符与个别函数差异);
2. `make sqlc`(生成三个包)→ `make gen-store`(重新生成适配层);
3. 若方法签名有变,更新 `store/store.go` 接口与 `store/models.go`;
4. `go build ./...` 编译期兜底:三个适配层任何一处签名不一致都无法编译——
   **这是防「忘改某个方言」的主要机制**;
5. CI 三引擎测试全绿方可合并。

新增迁移:三个目录同版本号各写一份,遵循 §4 类型映射表。

## 9. 实施阶段

### 阶段一:解耦(行为不变,可独立合并)

- [ ] 新建 `db/store`:接口 + 中立结构体 + TxManager 接口;
- [ ] `tools/gen-store` 生成器;`db/sqlc` → `db/pg`,包一层 `pgstore`;
- [ ] 29 个 repository 依赖 `*sqlc.Queries` → `store.Store`(84 处引用,机械替换);
- [ ] 22 处 `RunInTx` 回调签名切换;
- [ ] 53 处 `NOW()` 改 Go 传参(query + repository 两侧);
- [ ] 全量测试通过,行为与 main 一致。

### 阶段二:SQLite

- [ ] `db/queries/sqlite` + `db/migrations/sqlite` 改写;
- [ ] modernc 驱动接入、DSN/PRAGMA、`sql.Tx` 版 TxManager、`sqlitestore` 适配层;
- [ ] dbtest 参数化 + 本地默认引擎切 SQLite;
- [ ] 端到端冒烟通过。

### 阶段三:MySQL

- [ ] `db/queries/mysql` + `db/migrations/mysql` 改写(含 RETURNING → execresult);
- [ ] go-sql-driver 接入、`mysqlstore` 适配层;
- [ ] CI service container + 三引擎矩阵;
- [ ] 文档:README 配置示例 ×3、版本要求、换库说明。

### 工作量与风险

| 阶段 | 规模感 | 主要风险 |
|---|---|---|
| 一 | 改动面最广但全机械;diff 数千行(多为搬移/生成) | `NOW()` 传参遗漏 → 靠测试断言时间字段兜底 |
| 二 | 17 query + 11 迁移改写,方言差异小 | SQLite 时间/并发语义 → PRAGMA + 池限 1 + 矩阵测试 |
| 三 | 同二 + RETURNING 改写 ~20 处 | LastInsertId 回读的事务一致性 → 适配层内同连接保证 |

长期成本(架构决定,非一次性):每条 SQL 三份方言、CI 三倍时长。缓解:§8 的
编译期兜底 + 派生脚本 + 测试矩阵。
