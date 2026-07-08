# 基础库封装方案(防依赖锁定)— light-admin 版

> 状态:批次 0/1/3 已落地(2026-07-08),批次 2(slog 收敛)待做。
> 姊妹篇:`docs/multi-database-plan.md`(pgx 的抽离由该方案的阶段一负责,本文
> 不重复)。airdesk 有同名方案(`airdesk/admin-server/docs/infra-abstraction-plan.md`),
> 判断原则与机制完全相同;本项目是 airdesk 的上游脚手架,**这里先落地即是
> airdesk 的试点**——与多数据库方案的实施顺序逻辑一致。

## 0. 判断原则(与 airdesk 版相同,摘要)

**扩散度 × 替换概率 × 抽象成本**三者都高才封装;已收敛的维持;不为封装而封装。
**守边界(CI import 检查)比包接口更重要**:封装解决存量锁定,守门防止增量锁定。

## 1. 盘点总表(2026-07-08 实测,非测试文件引用数)

| 库 | 扩散 | 现状 | 处置 |
|---|---|---|---|
| labstack/echo v4 | 28 | **100% 在 controller/route/middleware 表现层**,service/repository 零引用 | 不抽象框架,边界固化进 CI(§4) |
| go-redis v8(+ go-redis/cache) | **1** | 只有 `lib/cache_redis.go`,已完美藏在 `Cache` 接口后(memory/redis 双实现) | ✅ 结构已达标;**但 v8 已停止维护**,做 v9 升级(§2) |
| zap(+lumberjack) | 5 个 import,**123 处 `.Zap.` 调用 / 75 个文件依赖 `lib.Logger`** | `lib.Logger` 结构体直接暴露 `*zap.SugaredLogger` 字段 → 事实上全项目锁死 zap | ★ 唯一的真锁定,收敛到 slog(§3) |
| golang-jwt v5 | 4 | auth_service / member_auth_service / models/dto 两处 claims 各自解析 | `TokenCodec` 接口收编(§5) |
| gorilla/websocket | 3 | pkg/websocket hub(已是抽象点)+ aria2 rpc 独立客户端 | 维持 |
| minio-go | 1 | file_service 内 OSS 后端选择(local/minio/aliyun) | ✅ 已收敛,维持 |
| base64Captcha / robfig-cron / viper / useragent | 各 1 | lib/captcha、pkg/crontab、lib/config、log_middleware | ✅ 已收敛,维持 |
| jackc/pgx | 17(repositories) | — | 由 `multi-database-plan.md` 阶段一(store 接口层)处理,本文不重复 |
| fx / cobra / validator / samber-lo / testify | — | 框架级或工具级 | 不封装 |

**与 airdesk 的关键差异**:airdesk 的重灾区 Redis(21 文件、五种用途)在本项目
不存在——这里 Redis 只做缓存且已有接口,**不需要拆 EventBus/StreamQueue 那套窄
接口**,只欠一次 v9 升级。本项目唯一的真锁定是 zap。

## 2. Redis:v8 → v9 升级(结构不动)

`Cache` 接口形状不变,改动应只落在:

- `lib/cache_redis.go`(import 路径 `redis/v8`→`redis/v9`,API 差异极小);
- `go-redis/cache/v8` → v9(或顺手用 redis 原生命令替掉这个薄依赖,少一个库);
- `go.mod`。

**这次升级同时是「已封装好」的验证**:除上述文件外任何地方需要改动,都说明
Cache 接口漏了,就地修补。开源项目角度还有一个理由:v8 依赖会让下游用户的
`go get -u` 和安全扫描持续报警。

## 3. 日志:zap → slog 门面(本项目的主菜)

**问题**:`lib.Logger{ Zap *zap.SugaredLogger }` 出现在 75 个文件的构造函数
签名里,123 处 `logger.Zap.Xxx` 调用——换/删 zap 等于全项目动一遍。

**方案**(与 airdesk 版相同):业务面统一依赖标准库 `*slog.Logger`,zap 经官方
`zapslog` handler 做后端,保留现有的 lumberjack 轮转与格式配置。

- 选 slog 的理由:标准库=零锁定、零学习成本;对开源脚手架尤其合适——下游用户
  接自己的日志方案只需换 handler,不用理解你的自定义 Logger;
- 迁移路径:`lib.Logger` 先内部加 `Slog *slog.Logger` 字段过渡 → 按包把
  `logger.Zap.Infof(...)` 换成 `logger.Slog.Info(...)`(123 处,纯机械,注意
  printf 风格改 key-value 风格)→ 最后构造函数签名 `lib.Logger`→`*slog.Logger`
  (75 个文件)并删掉 Zap 字段;
- fxlog、zap_middleware(echo 请求日志)同步对接 slog handler。

## 4. 边界守护(CI,最先做)

golangci-lint **depguard**:

```
api/**/service/**、api/**/repository/**、pkg/**(队列/下载器等领域包):
  禁止 import:labstack/echo、go-redis、jackc/pgx、go.uber.org/zap、golang-jwt
  白名单入口:lib、api/**/controller、api/**/route、api/middlewares、
              pkg/echox、pkg/websocket(hub 对 gorilla 的合法使用)
```

先 warn 摸底,存量收敛一项提一项到 error。对开源项目这条还有示范价值:
贡献者的 PR 想在 service 里直接摸 redis/echo,CI 直接拦下,不靠 review 记忆。

## 5. 顺手小件

- **JWT**:`TokenCodec` 接口(Sign/Parse),收编 system 与 member 两套 auth
  service 加 models/dto 里的散装 jwt/v5 用法(4 文件)。golang-jwt 历史上换过
  两次组织,值得这一层;
- 顺手统一 uuid:go.mod 里同时有 `gofrs/uuid` 和 `google/uuid` 两个库,
  `pkg/uuid` 已存在,收敛到一个(非锁定问题,纯卫生)。

## 6. 实施批次

| 批次 | 内容 | 规模感 | 状态 |
|---|---|---|---|
| 0 | depguard 上线(warn 起步) | 半天级 | ✅ `.golangci.yml`(depguard,warning 级);摸底:仅 5 处 pgx@repository(归 multi-database-plan 存量),echo/redis/zap 在业务层零直引 |
| 1 | go-redis v8→v9(1 文件 + go.mod) | 小,顺带验证 Cache 封装 | ✅ 仅动 `lib/cache_redis.go`,Cache 接口未漏——封装达标已验证 |
| 2 | slog 收敛(123 调用点 + 75 签名,脚本辅助) | 本方案主体,纯机械 | ⏳ 待做(独立会话) |
| 3 | TokenCodec + uuid 收敛 | 小 | ✅ `lib.TokenCodec`(Sign/Parse + Err* + RegisteredClaims 转出);两套 auth service + 两个 dto claims 脱 jwt;gofrs/uuid 全量并入 `pkg/uuid`(google 背书),移除一库 |

与 `multi-database-plan.md` 阶段一(pgx→store)互不阻塞,可并行;depguard 一条
规则同时护住两个方案的边界。全部落地后,本项目对第三方库的暴露面 =
lib + 表现层 + 各领域包的单文件适配点,任何库的替换/移除都不再惊天动地。
