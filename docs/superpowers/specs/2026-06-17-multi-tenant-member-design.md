# 多租户 + 会员体系 设计文档

**日期**: 2026-06-17
**作者**: avsdk (协同 Claude Code)
**状态**: 已确认，待生成实施计划

## 1. 背景与目标

为 `light-admin` 增加两项能力：

1. **多租户（Multi-Tenant）**：采用行级隔离（共享库共享表 + `tenant_id` 列），可通过配置整体关闭。
2. **会员体系（Member）**：C 端自助注册/登录，独立鉴权，按租户隔离。

约束（用户已拍板）：

- **隔离模式**：模式 A —— 共享表 + `tenant_id` 列（行级隔离），GORM Scope 自动注入过滤。
- **后台范围**：现有 RBAC 后台（`t_user`/角色/菜单/部门/字典等）**保持平台全局单租户，完全不改动**。只有新增的 `member` 会员体系按租户隔离。
- **会员鉴权**：**独立体系** —— 独立的 `/api/v1/member/auth/*` 注册登录 + 独立 JWT + 独立中间件，与后台 admin 鉴权互不影响。
- **租户归属**：会员注册接口携带租户标识（子域名 / 租户码 / 请求头）解析出 `tenant_id`。
- **可关闭**：配置 `MultiTenant.Enable=false` 时，所有会员归属固定 `default` 租户，业务零改动。

非目标（YAGNI，本期不做）：

- 套餐/计费、会员配额（`max_members`）、租户到期（`expire_time`）——**已明确移除**。
- 模式 B/C（独立 schema/库）。
- C 端会员前端 UI 页面（本期只交付会员 API；后台新增「租户管理」「会员管理」两个页面）。

## 2. 技术栈与现有架构对齐

- 后端：Go 1.24 + Echo + GORM + Casbin + Uber-FX；4 层域模块 `route → controller → service → repository`。
- 主键风格：`char(32)` UUID（与最近一次迁移一致）。
- 密码哈希：复用 `pkg/hash` 的 `BcryptHash` / `BcryptCheck`（与 `t_user` 一致）。
- 配置：Viper YAML，`lib/config.go` 的 `Config` 结构体。
- 迁移：`cmd/migrate/migrate.go` 显式注册模型；`cmd/setup` 做种子数据。

## 3. 数据模型

### 3.1 `t_tenant` 租户表（平台管理员维护，全局可见）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | char(32) PK | UUID |
| code | varchar(64) | 租户码，**唯一索引**，用于子域名/请求头解析 |
| name | varchar(128) | 租户名称 |
| status | int default 1 | 1 正常 / 0 禁用 |
| create_time | datetime | autoCreateTime |
| create_by | varchar | 创建人（平台管理员 id） |
| update_time | datetime | autoUpdateTime |
| update_by | varchar | |
| is_deleted | int default 0 | 软删除标志，对齐现有风格 |

> 不包含 `expire_time`、`max_members`（已移除）。

### 3.2 `t_member` 会员表（C 端用户，租户隔离）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | char(32) PK | UUID |
| tenant_id | char(32) | **隔离列**，索引 |
| username | varchar(64) | 与 tenant_id 组合唯一 `uniq_tenant_username (tenant_id, username)` |
| email | varchar(128) | 与 tenant_id 组合唯一 `uniq_tenant_email (tenant_id, email)`（email 为空时不冲突） |
| mobile | varchar(20) | |
| password | varchar(100) | bcrypt 哈希，`json:"-"`，序列化时不输出 |
| nickname | varchar(64) | |
| avatar | varchar(255) | |
| gender | int default 0 | 0 保密 / 1 男 / 2 女 |
| status | int default 1 | 1 正常 / 0 禁用 |
| last_login_time | datetime | 登录审计 |
| last_login_ip | varchar(64) | 登录审计 |
| create_time | datetime | autoCreateTime |
| update_time | datetime | autoUpdateTime |
| is_deleted | int default 0 | 软删除 |

设计要点：不同租户可用相同 `username`/`email`（组合唯一），这是主流 SaaS 标准做法。

模型位置：`models/tenant/tenant.go`、`models/member/member.go`（与 `models/system` 同级）。

## 4. 配置开关

`config.yaml` / `config.yaml.default` 新增段：

```yaml
MultiTenant:
  Enable: true            # 多租户总开关
  DefaultTenant: default  # 关闭时所有会员归属此固定租户 code
  Resolver: header        # header | subdomain | path
  HeaderName: X-Tenant-Code
```

`lib/config.go`：

- 新增 `MultiTenantConfig` 结构体（字段 `Enable bool`、`DefaultTenant string`、`Resolver string`、`HeaderName string`）。
- 加入 `Config` 结构体与 `defaultConfig` 默认值（`Enable:false` 作为安全默认，`DefaultTenant:"default"`，`Resolver:"header"`，`HeaderName:"X-Tenant-Code"`）。
- 辅助方法：`(*MultiTenantConfig) IsEnabled() bool`。

**关闭多租户的退化行为**：

- 租户解析中间件直接返回 `DefaultTenant` 对应的 tenant_id。
- 会员注册/登录正常工作，全部落到 default 租户。
- 后台「租户管理」接口仍可调用但无实际意义（前端菜单可按开关隐藏，后端不强制）。
- `cmd/setup` 始终预置一个 `code=default` 的租户兜底。

## 5. 后端模块结构

新增域模块 `api/member/`，沿用 4 层 + 各自 FX module，在 `api/module.go` 注册。

### 5.1 资源与端点

**会员自助鉴权（公开 + 会员 JWT，经租户解析中间件）**

| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | /api/v1/member/auth/register | 公开 + 租户解析 | 会员注册（绑定解析出的 tenant_id） |
| POST | /api/v1/member/auth/login | 公开 + 租户解析 | 会员登录，返回会员 JWT |
| DELETE | /api/v1/member/auth/logout | 会员 JWT | 注销 |
| GET | /api/v1/member/profile | 会员 JWT | 获取当前会员资料 |
| PUT | /api/v1/member/profile | 会员 JWT | 更新当前会员资料 |

**平台管理员管理（走现有 admin JWT + Casbin）**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/tenants | 租户分页列表 |
| POST | /api/v1/tenants | 创建租户 |
| PUT | /api/v1/tenants/:id | 更新租户 |
| DELETE | /api/v1/tenants/:id | 删除租户（软删） |
| GET | /api/v1/members | 会员分页列表（按 tenant 过滤查询） |
| PUT | /api/v1/members/:id/status | 启用/禁用会员 |
| PUT | /api/v1/members/:id/password | 重置会员密码 |

### 5.2 三个横切件

1. **租户解析中间件** `api/middlewares/tenant_middleware.go`
   - 按 `config.MultiTenant.Resolver` 从 header（默认 `X-Tenant-Code`）/子域名/path 取租户码。
   - 查 `t_tenant` by code，校验存在且 `status==1`；不存在/禁用则返回明确错误。
   - 将 `tenant_id` 注入 echo context（key 常量放 `constants`）。
   - `Enable==false`：跳过解析，直接注入 DefaultTenant 的 tenant_id。
   - 仅挂在 `/api/v1/member/*` 路由组，不影响后台。

2. **会员鉴权中间件** `api/middlewares/member_auth_middleware.go`
   - 独立签名密钥（`fmt.Sprintf("Jwt:member:%s", config.Name)`），与 admin auth 区分。
   - 会员 JWT claims：`MemberID + TenantID + Username`（新 `dto.MemberClaims`）。
   - 校验通过后注入会员身份 + tenant_id 到 context。
   - 独立的 IgnorePath（register/login 公开）。

3. **GORM 租户 Scope** `pkg/scopes/tenant.go`（或 `models/database`）
   - `func Tenant(tenantID string) func(*gorm.DB) *gorm.DB`，会员 repository 所有查询统一挂载 `.Scopes(scopes.Tenant(tenantID))`。
   - 写入时 service 层强制 set `tenant_id`，不信任客户端传入。
   - 备注：主流做法可升级为注册 GORM 全局 callback 对实现 `TenantModel` 接口的模型自动注入；本期先用显式 Scope（简单、透明、易测）。

### 5.3 会员鉴权 service

新增 `api/member/service/member_auth_service.go`：

- `GenerateToken(member)` → 会员 JWT（独立密钥、`MemberClaims`）。
- `ParseToken` / `DestroyToken`，结构对照现有 `AuthService`，但完全独立，避免耦合。

## 6. 迁移与种子

- `cmd/migrate/migrate.go`：`AutoMigrate` 追加 `&tenant.Tenant{}`、`&member.Member{}`。
- `cmd/setup/setup.go`：若不存在则插入 `code=default, name=默认租户, status=1` 的租户。
- 前端菜单：`config/menu.yaml` 新增「租户管理」「会员管理」节点（受多租户开关影响时前端可隐藏租户管理）。

## 7. 前端（light-admin-ui）

- 新增后台页面：**租户管理**、**会员管理**，复用现有 ProTable/CRUD 范式与 services 封装。
- C 端会员注册/登录页本期**不做 UI**，只交付后端 API。
- 路由与菜单：`config/routes.ts` + `menu.yaml` 对应新增。

## 8. 安全要点

- 会员密码 bcrypt 存储，`json:"-"` 不外泄；登录失败信息不区分"用户不存在/密码错误"。
- 写操作的 `tenant_id` 一律由服务端从 context 注入，**禁止信任请求体中的 tenant_id**（防越权跨租户写）。
- 会员查询强制 tenant scope，防止跨租户读。
- 注册接口需校验租户存在且启用；对未知租户码返回统一错误，避免枚举。
- 注册/登录端点加入限流（项目已有 `ratelimit_middleware.go`，复用）。

## 9. 测试

- 单元：`scopes.Tenant` 过滤、会员注册（组合唯一冲突）、密码哈希校验、会员 JWT 生成/解析。
- 集成：注册→登录→profile 全链路；多租户开关 on/off 两种模式；跨租户隔离（A 租户会员不可见 B 租户数据）。
- 表驱动 + `-race`，对齐 `golang/testing` 规则。

## 10. 影响面（现有代码改动清单）

| 文件 | 改动 |
|------|------|
| `lib/config.go` | 新增 `MultiTenantConfig` + 默认值 |
| `config/config.yaml(.default)` | 新增 `MultiTenant` 段 |
| `api/module.go` | 注册新 `member` 模块 |
| `api/middlewares/*` | 新增 tenant + member_auth 中间件，并在中间件链/路由组装载 |
| `cmd/migrate/migrate.go` | 追加两张表 |
| `cmd/setup/setup.go` | 预置 default 租户 |
| `constants/*` | 新增 context key 常量 |
| 新增 `models/tenant/`、`models/member/`、`api/member/`、`pkg/scopes/` | 全新文件 |
| 前端 `config/routes.ts`、`menu.yaml`、新增页面与 services | 后台两页面 + API 封装 |

现有 `api/system/`、`models/system/`、RBAC、Casbin **零改动**。
