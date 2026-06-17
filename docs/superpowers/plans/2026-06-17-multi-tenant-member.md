# 多租户 + 会员体系 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 light-admin 增加行级隔离的多租户能力与独立鉴权的 C 端会员体系（注册/登录/资料 + 后台租户/会员管理），多租户可通过配置整体关闭。

**Architecture:** 模式 A 行级隔离——业务表加 `tenant_id` 列，GORM Scope 自动过滤。现有 RBAC 后台零改动；新增独立 `api/member` 域模块（4 层 route→controller→service→repository）+ 两个路由组级中间件（租户解析、会员 JWT）。关闭多租户时所有会员归属固定 `default` 租户。

**Tech Stack:** Go 1.24 + Echo + GORM + Uber-FX + JWT(golang-jwt/v5) + bcrypt(`pkg/hash`)；前端 React 19 + UmiMax + Ant Design Pro。

**Spec:** `docs/superpowers/specs/2026-06-17-multi-tenant-member-design.md`

**关键约定（已从代码核实）:**
- 主键自动生成：`lib/db.go` 全局回调 `light:assign_uuid` 会为「字符串空主键」自动写入 32 位 UUID。模型 Create 时 **留空 ID 即可**。
- 统一响应：`echox.Response{Code, Data, Message, Page}.JSON(ctx)`。
- 错误：`errors.New(...)` + `errors.RegisterHTTPStatus(...)`；仓储层用 `errors.Wrap(errors.DatabaseInternalError, ...)`。
- 后台 admin 鉴权是全局中间件（`engine.Use`），靠 `Auth.IgnorePathPrefixes` / `Casbin.IgnorePathPrefixes` 放行；**会员路由必须加入这两个忽略前缀**，再在路由组挂会员自己的中间件。
- 路由组：`handler.RouterV1`（前缀 `/api/v1`）。
- 测试命令：`go test ./... -race`（在 `light-admin-server` 目录下）。

---

## File Structure

**后端新增:**
- `models/tenant/tenant.go` — Tenant 模型 + 查询参数/表单
- `models/member/member.go` — Member 模型 + 查询参数/表单
- `models/dto/member_auth.go` — 会员注册/登录 DTO + MemberClaims
- `pkg/scopes/tenant.go` — GORM 租户 Scope 助手
- `errors/tenant.go` — 租户/会员错误
- `api/member/module.go` — member 域 FX 聚合
- `api/member/repository/{repository.go,tenant_repository.go,member_repository.go}`
- `api/member/service/{service.go,member_auth_service.go,tenant_service.go,member_service.go}`
- `api/member/controller/{controller.go,member_auth_controller.go,tenant_controller.go,member_controller.go}`
- `api/member/route/{route.go,member_auth_route.go,tenant_route.go,member_route.go}`
- `api/middlewares/tenant_middleware.go` — 租户解析（路由组级）
- `api/middlewares/member_auth_middleware.go` — 会员 JWT（路由组级）

**后端修改:**
- `lib/config.go` — 新增 `MultiTenantConfig`
- `config/config.yaml` + `config/config.yaml.default` — 新增 `MultiTenant` 段 + member 忽略前缀
- `constants/constants.go` — context key 常量
- `api/module.go` — 注册 member 模块
- `api/middlewares/middlewares.go` — 注册两个新中间件 provider
- `cmd/migrate/migrate.go` — 追加两张表
- `cmd/setup/setup.go` — 预置 default 租户

**前端新增/修改:**
- `src/services/member/{tenant.ts,member.ts}` — API 封装
- `src/pages/member/tenant/index.tsx` — 租户管理页
- `src/pages/member/list/index.tsx` — 会员管理页
- `config/routes.ts` + 后端 `config/menu.yaml` — 路由与菜单

---

## Phase 0 — 配置、常量、错误地基

### Task 1: 多租户配置 MultiTenantConfig

**Files:**
- Modify: `lib/config.go`
- Modify: `config/config.yaml.default`
- Modify: `config/config.yaml`

- [ ] **Step 1: 在 `lib/config.go` 的 `Config` 结构体追加字段**

在 `Config` struct（约 72-88 行）`OSS` 字段后、扩展功能注释前加入：

```go
	MultiTenant *MultiTenantConfig `mapstructure:"MultiTenant"`
```

- [ ] **Step 2: 新增 `MultiTenantConfig` 类型（放在 `OSSConfig` 定义附近）**

```go
// MultiTenantConfig 多租户配置
// Enable=false 时所有会员归属 DefaultTenant 指定的固定租户
type MultiTenantConfig struct {
	Enable        bool   `mapstructure:"Enable"`        // 多租户总开关
	DefaultTenant string `mapstructure:"DefaultTenant"` // 关闭时归属的租户 code
	Resolver      string `mapstructure:"Resolver"`      // header | subdomain | path
	HeaderName    string `mapstructure:"HeaderName"`    // Resolver=header 时的请求头名
}

// IsEnabled 是否启用多租户
func (a *MultiTenantConfig) IsEnabled() bool {
	return a != nil && a.Enable
}

// ResolverHeaderName 返回租户码请求头名，默认 X-Tenant-Code
func (a *MultiTenantConfig) ResolverHeaderName() string {
	if a == nil || a.HeaderName == "" {
		return "X-Tenant-Code"
	}
	return a.HeaderName
}

// FallbackTenantCode 关闭多租户时的默认租户 code，默认 "default"
func (a *MultiTenantConfig) FallbackTenantCode() string {
	if a == nil || a.DefaultTenant == "" {
		return "default"
	}
	return a.DefaultTenant
}
```

- [ ] **Step 3: 在 `defaultConfig`（约 14-37 行）追加默认值**

在 `OSS:` 字段后加：

```go
	MultiTenant: &MultiTenantConfig{Enable: false, DefaultTenant: "default", Resolver: "header", HeaderName: "X-Tenant-Code"},
```

- [ ] **Step 4: 在 `config/config.yaml.default` 末尾追加配置段**

```yaml

# 多租户配置
# Enable: 多租户总开关；false 时所有会员归属 DefaultTenant 指定的固定租户
# Resolver: 租户解析方式 header | subdomain | path
MultiTenant:
  Enable: true
  DefaultTenant: default
  Resolver: header
  HeaderName: X-Tenant-Code
```

- [ ] **Step 5: 在 `config/config.yaml` 同步追加上面同样的 `MultiTenant` 段**（本地联调用，可设 `Enable: true`）

- [ ] **Step 6: 在 `config/config.yaml.default` 与 `config/config.yaml` 的 `Auth.IgnorePathPrefixes` 和 `Casbin.IgnorePathPrefixes` 各追加一行**

```yaml
    - /api/v1/member
```

> 这样后台全局 admin 鉴权/Casbin 会放行所有 `/api/v1/member/*`，由会员自己的中间件接管。

- [ ] **Step 7: 编译验证**

Run: `cd light-admin-server && go build ./...`
Expected: 编译通过，无报错。

- [ ] **Step 8: Commit**

```bash
git add light-admin-server/lib/config.go light-admin-server/config/config.yaml light-admin-server/config/config.yaml.default
git commit -m "feat(config): add MultiTenant config and member ignore-path prefixes"
```

---

### Task 2: 常量与错误定义

**Files:**
- Modify: `constants/constants.go`
- Create: `errors/tenant.go`

- [ ] **Step 1: 在 `constants/constants.go` 追加 context key 常量**

```go
// 多租户 / 会员 上下文 key
const CurrentTenantID = "current-tenant-id" // 解析出的租户 ID
const CurrentMember = "current-member"      // 会员 JWT claims
```

- [ ] **Step 2: 新建 `errors/tenant.go`**

```go
package errors

import "net/http"

var (
	TenantNotFound     = New("tenant not found")
	TenantDisabled     = New("tenant is disabled")
	TenantCodeRequired = New("tenant code is required")
	TenantCodeExists   = New("tenant code already exists")

	MemberRecordNotFound = New("member record not found")
	MemberAlreadyExists  = New("member already exists")
	MemberInvalidLogin   = New("invalid username or password")
	MemberIsDisabled     = New("member is disabled")

	MemberTokenInvalid = New("member token is invalid")
	MemberTokenExpired = New("member token is expired")
)

func init() {
	RegisterHTTPStatus(TenantNotFound, http.StatusNotFound)
	RegisterHTTPStatus(TenantDisabled, http.StatusForbidden)
	RegisterHTTPStatus(TenantCodeRequired, http.StatusBadRequest)
	RegisterHTTPStatus(TenantCodeExists, http.StatusConflict)

	RegisterHTTPStatus(MemberRecordNotFound, http.StatusNotFound)
	RegisterHTTPStatus(MemberAlreadyExists, http.StatusConflict)
	RegisterHTTPStatus(MemberInvalidLogin, http.StatusUnauthorized)
	RegisterHTTPStatus(MemberIsDisabled, http.StatusForbidden)

	RegisterHTTPStatus(MemberTokenInvalid, http.StatusUnauthorized)
	RegisterHTTPStatus(MemberTokenExpired, http.StatusUnauthorized)
}
```

- [ ] **Step 3: 编译验证**

Run: `cd light-admin-server && go build ./...`
Expected: 通过。

- [ ] **Step 4: Commit**

```bash
git add light-admin-server/constants/constants.go light-admin-server/errors/tenant.go
git commit -m "feat(errors): add tenant/member context keys and error definitions"
```

---

## Phase 1 — 模型、Scope、迁移

### Task 3: Tenant 模型

**Files:**
- Create: `models/tenant/tenant.go`

- [ ] **Step 1: 新建 `models/tenant/tenant.go`**

```go
package tenant

import "github.com/top-system/light-admin/models/dto"

// Tenant 租户模型
// Status: 1-正常 0-禁用
type Tenant struct {
	ID         string       `gorm:"primaryKey;type:char(32)" json:"id"`
	Code       string       `gorm:"column:code;size:64;uniqueIndex:uniq_tenant_code" json:"code"`
	Name       string       `gorm:"column:name;size:128" json:"name"`
	Status     int          `gorm:"column:status;default:1" json:"status"`
	CreateTime dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	CreateBy   string       `gorm:"column:create_by" json:"createBy"`
	UpdateTime dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	UpdateBy   string       `gorm:"column:update_by" json:"updateBy"`
	IsDeleted  int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (Tenant) TableName() string { return "t_tenant" }

type Tenants []*Tenant

// TenantForm 租户表单
type TenantForm struct {
	ID     string `json:"id"`
	Code   string `json:"code" validate:"required"`
	Name   string `json:"name" validate:"required"`
	Status int    `json:"status"`
}

// TenantQueryParam 租户查询参数
type TenantQueryParam struct {
	dto.PaginationParam
	dto.OrderParam
	Keywords string `query:"keywords"`
	Status   *int   `query:"status"`
}

// TenantQueryResult 租户查询结果
type TenantQueryResult struct {
	List       Tenants         `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}
```

> 注：`dto.DateTime` 与现有 `models/system/user.go` 使用的类型一致（`models/dto` 包导出）。

- [ ] **Step 2: 编译验证**

Run: `cd light-admin-server && go build ./models/...`
Expected: 通过。若 `dto.DateTime` 实际路径不同，按 `models/system/user.go` 第 4 行 import 修正为相同包。

- [ ] **Step 3: Commit**

```bash
git add light-admin-server/models/tenant/tenant.go
git commit -m "feat(models): add Tenant model"
```

---

### Task 4: Member 模型 + 会员鉴权 DTO

**Files:**
- Create: `models/member/member.go`
- Create: `models/dto/member_auth.go`

- [ ] **Step 1: 新建 `models/member/member.go`（注意 `TenantID` 用下面给出的完整组合唯一索引 tag）**

```go
package member

import "github.com/top-system/light-admin/models/dto"

// Member 会员模型（C 端用户，按 tenant_id 行级隔离）
// Status: 1-正常 0-禁用 ; Gender: 0-保密 1-男 2-女
type Member struct {
	ID            string       `gorm:"primaryKey;type:char(32)" json:"id"`
	TenantID      string       `gorm:"column:tenant_id;type:char(32);index:idx_member_tenant;uniqueIndex:uniq_tenant_username,priority:1;uniqueIndex:uniq_tenant_email,priority:1" json:"tenantId"`
	Username      string       `gorm:"column:username;size:64;uniqueIndex:uniq_tenant_username,priority:2" json:"username"`
	Email         string       `gorm:"column:email;size:128;uniqueIndex:uniq_tenant_email,priority:2" json:"email"`
	Mobile        string       `gorm:"column:mobile;size:20" json:"mobile"`
	Password      string       `gorm:"column:password;size:100" json:"-"`
	Nickname      string       `gorm:"column:nickname;size:64" json:"nickname"`
	Avatar        string       `gorm:"column:avatar;size:255" json:"avatar"`
	Gender        int          `gorm:"column:gender;default:0" json:"gender"`
	Status        int          `gorm:"column:status;default:1" json:"status"`
	LastLoginTime dto.DateTime `gorm:"column:last_login_time" json:"lastLoginTime"`
	LastLoginIP   string       `gorm:"column:last_login_ip;size:64" json:"lastLoginIp"`
	CreateTime    dto.DateTime `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime    dto.DateTime `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
	IsDeleted     int          `gorm:"column:is_deleted;default:0" json:"isDeleted"`
}

func (Member) TableName() string { return "t_member" }

type Members []*Member

// MemberProfileForm 会员资料表单（会员本人更新）
type MemberProfileForm struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   int    `json:"gender"`
	Mobile   string `json:"mobile"`
	Email    string `json:"email"`
}

// MemberQueryParam 后台会员查询参数
type MemberQueryParam struct {
	dto.PaginationParam
	dto.OrderParam
	TenantID string `query:"tenantId"`
	Keywords string `query:"keywords"`
	Status   *int   `query:"status"`
}

// MemberQueryResult 后台会员查询结果
type MemberQueryResult struct {
	List       Members         `json:"list"`
	Pagination *dto.Pagination `json:"pagination"`
}

func (a *Member) CleanSecure() *Member { a.Password = ""; return a }
```

> 索引：`(tenant_id, username)` 与 `(tenant_id, email)` 两个唯一索引使「不同租户可用相同用户名/邮箱」。

- [ ] **Step 2: 新建 `models/dto/member_auth.go`**

```go
package dto

import "github.com/golang-jwt/jwt/v5"

// MemberRegister 会员注册请求
type MemberRegister struct {
	Username string `json:"username" validate:"required,min=3,max=64"`
	Password string `json:"password" validate:"required,min=6,max=64"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

// MemberLogin 会员登录请求
type MemberLogin struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// MemberClaims 会员 JWT claims（独立于后台 JwtClaims）
type MemberClaims struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
```

- [ ] **Step 3: 编译验证**

Run: `cd light-admin-server && go build ./...`
Expected: 通过。

- [ ] **Step 4: Commit**

```bash
git add light-admin-server/models/member/member.go light-admin-server/models/dto/member_auth.go
git commit -m "feat(models): add Member model and member auth DTOs"
```

---

### Task 5: GORM 租户 Scope 助手

**Files:**
- Create: `pkg/scopes/tenant.go`
- Test: `pkg/scopes/tenant_test.go`

- [ ] **Step 1: 先写失败测试 `pkg/scopes/tenant_test.go`**

```go
package scopes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type memberRow struct {
	ID       string `gorm:"primaryKey"`
	TenantID string
	Name     string
}

func (memberRow) TableName() string { return "members" }

func setupDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&memberRow{}))
	db.Create(&memberRow{ID: "1", TenantID: "tA", Name: "alice"})
	db.Create(&memberRow{ID: "2", TenantID: "tB", Name: "bob"})
	return db
}

func TestTenantScopeFiltersByTenant(t *testing.T) {
	db := setupDB(t)

	var rows []memberRow
	err := db.Model(&memberRow{}).Scopes(Tenant("tA")).Find(&rows).Error

	assert.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "alice", rows[0].Name)
}

func TestTenantScopeEmptyTenantReturnsNone(t *testing.T) {
	db := setupDB(t)

	var rows []memberRow
	err := db.Model(&memberRow{}).Scopes(Tenant("")).Find(&rows).Error

	assert.NoError(t, err)
	assert.Len(t, rows, 0) // 空租户 ID 必须查不到任何数据，防止越权
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd light-admin-server && go test ./pkg/scopes/ -run TestTenantScope -v`
Expected: FAIL（`Tenant` 未定义，编译错误）。

- [ ] **Step 3: 实现 `pkg/scopes/tenant.go`**

```go
package scopes

import "gorm.io/gorm"

// Tenant 返回一个按 tenant_id 过滤的 GORM Scope。
// tenantID 为空时强制匹配空字符串（查不到任何数据），防止越权读取全表。
func Tenant(tenantID string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd light-admin-server && go test ./pkg/scopes/ -run TestTenantScope -v`
Expected: PASS（两个用例都通过）。

- [ ] **Step 5: Commit**

```bash
git add light-admin-server/pkg/scopes/
git commit -m "feat(scopes): add GORM tenant scope helper with tests"
```

---

### Task 6: 迁移注册 + 预置 default 租户

**Files:**
- Modify: `cmd/migrate/migrate.go`
- Modify: `cmd/setup/setup.go`

- [ ] **Step 1: 在 `cmd/migrate/migrate.go` import 追加**

```go
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/models/tenant"
```

- [ ] **Step 2: 在 `AutoMigrate(...)` 列表中、扩展功能模型之前追加**

```go
		// 多租户 / 会员
		&tenant.Tenant{},
		&member.Member{},
```

- [ ] **Step 3: 在 `cmd/setup/setup.go` 末尾（Run 函数内、菜单/管理员初始化之后）追加预置 default 租户逻辑**

在 import 追加 `"github.com/top-system/light-admin/models/tenant"`，并在 Run 函数末尾加：

```go
		// Step N: 预置 default 租户（多租户关闭时的兜底归属）
		var tenantCount int64
		db.ORM.Model(&tenant.Tenant{}).Where("code = ?", "default").Count(&tenantCount)
		if tenantCount == 0 {
			if err := db.ORM.Create(&tenant.Tenant{
				Code:   "default",
				Name:   "默认租户",
				Status: 1,
			}).Error; err != nil {
				logger.Zap.Fatalf("create default tenant err: %v", err)
			}
			logger.Zap.Info("default tenant created successfully")
		}
```

> ID 留空，由 `light:assign_uuid` 回调自动生成 32 位 UUID。

- [ ] **Step 4: 运行迁移与 setup 验证（SQLite 最快）**

Run:
```bash
cd light-admin-server && go run . migrate --config=./config/config.yaml && go run . setup --config=./config/config.yaml --menu=./config/menu.yaml
```
Expected: 日志出现 "Database migration completed successfully" 与 "default tenant created successfully"；数据库中存在 `t_tenant`、`t_member` 表及一条 `code=default` 记录。

- [ ] **Step 5: Commit**

```bash
git add light-admin-server/cmd/migrate/migrate.go light-admin-server/cmd/setup/setup.go
git commit -m "feat(cmd): migrate tenant/member tables and seed default tenant"
```

---

## Phase 2 — 仓储层

### Task 7: TenantRepository + 分页 helper

**Files:**
- Create: `api/member/repository/repository.go`
- Create: `api/member/repository/tenant_repository.go`

- [ ] **Step 1: 新建 FX 聚合 + 共享 helper `api/member/repository/repository.go`**

```go
package repository

import (
	"time"

	"go.uber.org/fx"

	"github.com/top-system/light-admin/models/dto"
)

// Module exports member repositories
var Module = fx.Options(
	fx.Provide(NewTenantRepository),
	fx.Provide(NewMemberRepository),
)

// dtoPagination 构造统一分页对象（tenant/member 仓储共用）
func dtoPagination(total int64, pageNum, pageSize int) dto.Pagination {
	return dto.Pagination{Total: total, PageNum: pageNum, PageSize: pageSize}
}

// nowDateTime 返回当前时间的 dto.DateTime（用于登录信息更新）
func nowDateTime() dto.DateTime {
	return dto.DateTime(time.Now())
}
```

> 实施前先打开 `models/dto` 的 datetime 定义确认 `dto.DateTime` 是否是 `time.Time` 的具名类型。若不是，则把 `nowDateTime` 改为该类型实际的构造方式；或在 `UpdateLoginInfo` 中改为只更新 `last_login_ip`（见 Task 8 说明）。

- [ ] **Step 2: 新建 `api/member/repository/tenant_repository.go`**

```go
package repository

import (
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/tenant"
)

type TenantRepository struct {
	db     lib.Database
	logger lib.Logger
}

func NewTenantRepository(db lib.Database, logger lib.Logger) TenantRepository {
	return TenantRepository{db: db, logger: logger}
}

func (a TenantRepository) Query(param *tenant.TenantQueryParam) (*tenant.TenantQueryResult, error) {
	db := a.db.ORM.Model(&tenant.Tenant{}).Where("is_deleted = ?", 0)
	if v := param.Keywords; v != "" {
		db = db.Where("code LIKE ? OR name LIKE ?", "%"+v+"%", "%"+v+"%")
	}
	if v := param.Status; v != nil {
		db = db.Where("status = ?", *v)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	var list tenant.Tenants
	err := db.Order(param.ParseOrder()).
		Offset((param.GetPageNum() - 1) * param.GetPageSize()).
		Limit(param.GetPageSize()).
		Find(&list).Error
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	p := dtoPagination(total, param.GetPageNum(), param.GetPageSize())
	return &tenant.TenantQueryResult{List: list, Pagination: &p}, nil
}

func (a TenantRepository) GetByCode(code string) (*tenant.Tenant, error) {
	var t tenant.Tenant
	err := a.db.ORM.Where("code = ? AND is_deleted = ?", code, 0).First(&t).Error
	if err != nil {
		return nil, errors.TenantNotFound
	}
	return &t, nil
}

func (a TenantRepository) Get(id string) (*tenant.Tenant, error) {
	var t tenant.Tenant
	err := a.db.ORM.Where("id = ? AND is_deleted = ?", id, 0).First(&t).Error
	if err != nil {
		return nil, errors.TenantNotFound
	}
	return &t, nil
}

func (a TenantRepository) Create(t *tenant.Tenant) error {
	if err := a.db.ORM.Create(t).Error; err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a TenantRepository) Update(id string, t *tenant.Tenant) error {
	err := a.db.ORM.Model(&tenant.Tenant{}).Where("id = ?", id).
		Select("code", "name", "status", "update_by").Updates(t).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a TenantRepository) Delete(id string) error {
	err := a.db.ORM.Model(&tenant.Tenant{}).Where("id = ?", id).
		Update("is_deleted", 1).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}
```

> 确认 `errors.DatabaseInternalError` 存在（其他仓储已使用，见 `api/system/repository/dept_repository.go`）。

- [ ] **Step 3: 编译验证（连同 Task 8 一起编译；本步骤可能因 MemberRepository 未定义而失败，属预期）**

Run: `cd light-admin-server && go build ./api/member/repository/...`
Expected: 待 Task 8 完成后通过。

---

### Task 8: MemberRepository（租户隔离）

**Files:**
- Create: `api/member/repository/member_repository.go`

- [ ] **Step 1: 新建 `api/member/repository/member_repository.go`**

```go
package repository

import (
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/scopes"
)

type MemberRepository struct {
	db     lib.Database
	logger lib.Logger
}

func NewMemberRepository(db lib.Database, logger lib.Logger) MemberRepository {
	return MemberRepository{db: db, logger: logger}
}

func (a MemberRepository) GetByUsername(tenantID, username string) (*member.Member, error) {
	var m member.Member
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("username = ? AND is_deleted = ?", username, 0).
		First(&m).Error
	if err != nil {
		return nil, errors.MemberRecordNotFound
	}
	return &m, nil
}

func (a MemberRepository) GetByID(tenantID, id string) (*member.Member, error) {
	var m member.Member
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ? AND is_deleted = ?", id, 0).
		First(&m).Error
	if err != nil {
		return nil, errors.MemberRecordNotFound
	}
	return &m, nil
}

func (a MemberRepository) ExistsUsername(tenantID, username string) (bool, error) {
	var count int64
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("username = ? AND is_deleted = ?", username, 0).
		Count(&count).Error
	if err != nil {
		return false, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return count > 0, nil
}

func (a MemberRepository) Create(m *member.Member) error {
	if err := a.db.ORM.Create(m).Error; err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateProfile(tenantID, id string, m *member.Member) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).
		Select("nickname", "avatar", "gender", "mobile", "email").
		Updates(m).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateStatus(tenantID, id string, status int) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).Update("status", status).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdatePassword(tenantID, id, hashed string) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).Update("password", hashed).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

func (a MemberRepository) UpdateLoginInfo(tenantID, id, ip string) error {
	err := a.db.ORM.Model(&member.Member{}).
		Scopes(scopes.Tenant(tenantID)).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_ip":   ip,
			"last_login_time": nowDateTime(),
		}).Error
	if err != nil {
		return errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	return nil
}

// Query 后台分页查询（可跨租户，按 tenantId 可选过滤）
func (a MemberRepository) Query(param *member.MemberQueryParam) (*member.MemberQueryResult, error) {
	db := a.db.ORM.Model(&member.Member{}).Where("is_deleted = ?", 0)
	if v := param.TenantID; v != "" {
		db = db.Where("tenant_id = ?", v)
	}
	if v := param.Keywords; v != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", "%"+v+"%", "%"+v+"%", "%"+v+"%")
	}
	if v := param.Status; v != nil {
		db = db.Where("status = ?", *v)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}

	var list member.Members
	err := db.Order(param.ParseOrder()).
		Offset((param.GetPageNum() - 1) * param.GetPageSize()).
		Limit(param.GetPageSize()).
		Find(&list).Error
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseInternalError, err.Error())
	}
	for _, m := range list {
		m.CleanSecure()
	}

	p := dtoPagination(total, param.GetPageNum(), param.GetPageSize())
	return &member.MemberQueryResult{List: list, Pagination: &p}, nil
}
```

> 若 Step 1 中 `nowDateTime()` 因 `dto.DateTime` 非 `time.Time` 具名类型而编译失败，则把 `UpdateLoginInfo` 的 map 改为只 `{"last_login_ip": ip}`，并删除 `repository.go` 里的 `nowDateTime`/`time` import。

- [ ] **Step 2: 编译验证**

Run: `cd light-admin-server && go build ./api/member/repository/...`
Expected: 通过。

- [ ] **Step 3: Commit**

```bash
git add light-admin-server/api/member/repository/
git commit -m "feat(member): add tenant and member repositories"
```

---

## Phase 3 — 服务层

### Task 9: MemberAuthService（独立会员 JWT）

**Files:**
- Create: `api/member/service/service.go`
- Create: `api/member/service/member_auth_service.go`
- Test: `api/member/service/member_auth_service_test.go`

- [ ] **Step 1: 新建 FX 聚合 `api/member/service/service.go`**

```go
package service

import "go.uber.org/fx"

// Module exports member services
var Module = fx.Options(
	fx.Provide(NewMemberAuthService),
	fx.Provide(NewTenantService),
	fx.Provide(NewMemberService),
)
```

- [ ] **Step 2: 先写失败测试 `api/member/service/member_auth_service_test.go`**

```go
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/member"
)

func newTestAuth() MemberAuthService {
	cfg := lib.Config{Name: "test", Auth: &lib.AuthConfig{TokenExpired: 7200}}
	return NewMemberAuthService(cfg)
}

func TestMemberTokenRoundTrip(t *testing.T) {
	svc := newTestAuth()
	m := &member.Member{ID: "m1", TenantID: "tA", Username: "alice"}

	resp, err := svc.GenerateToken(m)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)

	claims, err := svc.ParseToken(resp.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, "m1", claims.ID)
	assert.Equal(t, "tA", claims.TenantID)
	assert.Equal(t, "alice", claims.Username)
}

func TestMemberParseInvalidToken(t *testing.T) {
	svc := newTestAuth()
	_, err := svc.ParseToken("not-a-token")
	assert.Error(t, err)
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `cd light-admin-server && go test ./api/member/service/ -run TestMember -v`
Expected: FAIL（编译错误，`NewMemberAuthService` 等未定义）。临时把 `service.go` Module 中 `NewTenantService`、`NewMemberService` 两行注释，避免未定义符号阻塞本任务测试。

- [ ] **Step 4: 实现 `api/member/service/member_auth_service.go`**

```go
package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	apperrors "github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
)

type MemberAuthService struct {
	issuer        string
	signingMethod jwt.SigningMethod
	signingKey    []byte
	keyfunc       jwt.Keyfunc
	expired       int
	tokenType     string
}

func NewMemberAuthService(config lib.Config) MemberAuthService {
	issuer := config.Name
	signingKey := []byte(fmt.Sprintf("Jwt:member:%s", issuer)) // 独立密钥
	expired := 7200
	if config.Auth != nil && config.Auth.TokenExpired > 0 {
		expired = config.Auth.TokenExpired
	}

	return MemberAuthService{
		issuer:        issuer,
		tokenType:     "Bearer",
		expired:       expired,
		signingMethod: jwt.SigningMethodHS512,
		signingKey:    signingKey,
		keyfunc: func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, apperrors.MemberTokenInvalid
			}
			return signingKey, nil
		},
	}
}

func (a MemberAuthService) GenerateToken(m *member.Member) (*dto.LoginResponse, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(a.expired) * time.Second)
	claims := &dto.MemberClaims{
		ID:       m.ID,
		TenantID: m.TenantID,
		Username: m.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.issuer,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(a.signingMethod, claims)
	accessToken, err := token.SignedString(a.signingKey)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken: accessToken,
		TokenType:   a.tokenType,
		ExpiresIn:   a.expired,
	}, nil
}

func (a MemberAuthService) ParseToken(tokenString string) (*dto.MemberClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &dto.MemberClaims{}, a.keyfunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.MemberTokenExpired
		}
		return nil, apperrors.MemberTokenInvalid
	}
	if claims, ok := token.Claims.(*dto.MemberClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, apperrors.MemberTokenInvalid
}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd light-admin-server && go test ./api/member/service/ -run TestMember -v`
Expected: PASS。

- [ ] **Step 6: Commit**

```bash
git add light-admin-server/api/member/service/service.go light-admin-server/api/member/service/member_auth_service.go light-admin-server/api/member/service/member_auth_service_test.go
git commit -m "feat(member): add independent member auth service with JWT tests"
```

---

### Task 10: TenantService

**Files:**
- Create: `api/member/service/tenant_service.go`

- [ ] **Step 1: 新建 `api/member/service/tenant_service.go`**

```go
package service

import (
	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/tenant"
)

type TenantService struct {
	logger     lib.Logger
	tenantRepo repository.TenantRepository
}

func NewTenantService(logger lib.Logger, tenantRepo repository.TenantRepository) TenantService {
	return TenantService{logger: logger, tenantRepo: tenantRepo}
}

func (a TenantService) Query(param *tenant.TenantQueryParam) (*tenant.TenantQueryResult, error) {
	return a.tenantRepo.Query(param)
}

func (a TenantService) Get(id string) (*tenant.Tenant, error) {
	return a.tenantRepo.Get(id)
}

// ResolveByCode 校验租户码并返回启用中的租户
func (a TenantService) ResolveByCode(code string) (*tenant.Tenant, error) {
	if code == "" {
		return nil, errors.TenantCodeRequired
	}
	t, err := a.tenantRepo.GetByCode(code)
	if err != nil {
		return nil, errors.TenantNotFound
	}
	if t.Status != 1 {
		return nil, errors.TenantDisabled
	}
	return t, nil
}

func (a TenantService) Create(form *tenant.TenantForm, operator string) (string, error) {
	if _, err := a.tenantRepo.GetByCode(form.Code); err == nil {
		return "", errors.TenantCodeExists
	}
	t := &tenant.Tenant{
		Code:     form.Code,
		Name:     form.Name,
		Status:   form.Status,
		CreateBy: operator,
	}
	if err := a.tenantRepo.Create(t); err != nil {
		return "", err
	}
	return t.ID, nil
}

func (a TenantService) Update(id string, form *tenant.TenantForm, operator string) error {
	if _, err := a.tenantRepo.Get(id); err != nil {
		return err
	}
	return a.tenantRepo.Update(id, &tenant.Tenant{
		Code:     form.Code,
		Name:     form.Name,
		Status:   form.Status,
		UpdateBy: operator,
	})
}

func (a TenantService) Delete(id string) error {
	if _, err := a.tenantRepo.Get(id); err != nil {
		return err
	}
	return a.tenantRepo.Delete(id)
}
```

- [ ] **Step 2: 取消 service.go Module 中 `NewTenantService` 的注释；编译验证**

Run: `cd light-admin-server && go build ./api/member/service/...`
Expected: 通过（`NewMemberService` 仍缺则保持其注释，Task 11 恢复）。

- [ ] **Step 3: Commit**

```bash
git add light-admin-server/api/member/service/tenant_service.go light-admin-server/api/member/service/service.go
git commit -m "feat(member): add tenant service"
```

---

### Task 11: MemberService（注册/登录/资料/后台管理）

**Files:**
- Create: `api/member/service/member_service.go`
- Test: `api/member/service/member_service_test.go`

- [ ] **Step 1: 先写测试 `api/member/service/member_service_test.go`（锁定密码校验语义）**

```go
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/hash"
)

func TestVerifyPasswordMatch(t *testing.T) {
	hashed, _ := hash.BcryptHash("secret123")
	m := &member.Member{Password: hashed, Status: 1}

	assert.True(t, hash.BcryptCheck("secret123", m.Password))
	assert.False(t, hash.BcryptCheck("wrong", m.Password))
}
```

> 注册/登录的全链路（含 DB）由 Phase 6 端到端冒烟覆盖；此处用纯函数测试保证 TDD 红绿，避免在单测中起 DB。

- [ ] **Step 2: 运行测试确认通过（hash 已存在）**

Run: `cd light-admin-server && go test ./api/member/service/ -run TestVerifyPasswordMatch -v`
Expected: 编译失败（MemberService 未定义导致包不编译）→ 先做 Step 3，再回到本步骤验证 PASS。

- [ ] **Step 3: 实现 `api/member/service/member_service.go`**

```go
package service

import (
	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/hash"
)

type MemberService struct {
	logger     lib.Logger
	memberRepo repository.MemberRepository
}

func NewMemberService(logger lib.Logger, memberRepo repository.MemberRepository) MemberService {
	return MemberService{logger: logger, memberRepo: memberRepo}
}

// Register 在指定租户内注册会员
func (a MemberService) Register(tenantID string, form *dto.MemberRegister) (*member.Member, error) {
	exists, err := a.memberRepo.ExistsUsername(tenantID, form.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.MemberAlreadyExists
	}

	hashed, err := hash.BcryptHash(form.Password)
	if err != nil {
		return nil, err
	}

	m := &member.Member{
		TenantID: tenantID, // 服务端强制注入，不信任客户端
		Username: form.Username,
		Password: hashed,
		Nickname: form.Nickname,
		Email:    form.Email,
		Mobile:   form.Mobile,
		Status:   1,
	}
	if err := a.memberRepo.Create(m); err != nil {
		return nil, err
	}
	return m.CleanSecure(), nil
}

// Verify 校验会员账号密码（登录用）
func (a MemberService) Verify(tenantID, username, password string) (*member.Member, error) {
	m, err := a.memberRepo.GetByUsername(tenantID, username)
	if err != nil {
		return nil, errors.MemberInvalidLogin // 不区分"不存在/密码错"，防枚举
	}
	if !hash.BcryptCheck(password, m.Password) {
		return nil, errors.MemberInvalidLogin
	}
	if m.Status != 1 {
		return nil, errors.MemberIsDisabled
	}
	return m, nil
}

func (a MemberService) RecordLogin(tenantID, id, ip string) {
	_ = a.memberRepo.UpdateLoginInfo(tenantID, id, ip)
}

func (a MemberService) GetProfile(tenantID, id string) (*member.Member, error) {
	m, err := a.memberRepo.GetByID(tenantID, id)
	if err != nil {
		return nil, err
	}
	return m.CleanSecure(), nil
}

func (a MemberService) UpdateProfile(tenantID, id string, form *dto.MemberProfileForm) error {
	if _, err := a.memberRepo.GetByID(tenantID, id); err != nil {
		return err
	}
	return a.memberRepo.UpdateProfile(tenantID, id, &member.Member{
		Nickname: form.Nickname,
		Avatar:   form.Avatar,
		Gender:   form.Gender,
		Mobile:   form.Mobile,
		Email:    form.Email,
	})
}

// ---- 后台管理 ----

func (a MemberService) Query(param *member.MemberQueryParam) (*member.MemberQueryResult, error) {
	return a.memberRepo.Query(param)
}

func (a MemberService) SetStatus(tenantID, id string, status int) error {
	return a.memberRepo.UpdateStatus(tenantID, id, status)
}

func (a MemberService) ResetPassword(tenantID, id, newPassword string) error {
	hashed, err := hash.BcryptHash(newPassword)
	if err != nil {
		return err
	}
	return a.memberRepo.UpdatePassword(tenantID, id, hashed)
}
```

- [ ] **Step 4: 恢复 service.go Module 中 `NewMemberService`，编译 + 全量服务测试**

Run: `cd light-admin-server && go build ./api/member/... && go test ./api/member/service/ -v`
Expected: 编译通过，所有测试 PASS。

- [ ] **Step 5: Commit**

```bash
git add light-admin-server/api/member/service/member_service.go light-admin-server/api/member/service/member_service_test.go light-admin-server/api/member/service/service.go
git commit -m "feat(member): add member service (register/login/profile/admin)"
```

---

## Phase 4 — 中间件

### Task 12: 租户解析中间件

**Files:**
- Create: `api/middlewares/tenant_middleware.go`
- Modify: `api/middlewares/middlewares.go`

- [ ] **Step 1: 新建 `api/middlewares/tenant_middleware.go`**

```go
package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	memberService "github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/pkg/echox"
)

// TenantMiddleware 解析请求所属租户，注入 constants.CurrentTenantID
type TenantMiddleware struct {
	config        lib.Config
	logger        lib.Logger
	tenantService memberService.TenantService
}

func NewTenantMiddleware(config lib.Config, logger lib.Logger, tenantService memberService.TenantService) TenantMiddleware {
	return TenantMiddleware{config: config, logger: logger, tenantService: tenantService}
}

// Resolve 返回路由组级中间件
func (a TenantMiddleware) Resolve() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			code := a.extractCode(ctx)

			// 多租户关闭：统一落到默认租户
			if !a.config.MultiTenant.IsEnabled() {
				code = a.config.MultiTenant.FallbackTenantCode()
			}

			t, err := a.tenantService.ResolveByCode(code)
			if err != nil {
				return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
			}

			ctx.Set(constants.CurrentTenantID, t.ID)
			return next(ctx)
		}
	}
}

func (a TenantMiddleware) extractCode(ctx echo.Context) string {
	switch a.config.MultiTenant.Resolver {
	case "subdomain":
		host := ctx.Request().Host
		if i := strings.IndexByte(host, '.'); i > 0 {
			return host[:i]
		}
		return ""
	case "path":
		return ctx.Param("tenantCode")
	default: // header
		return ctx.Request().Header.Get(a.config.MultiTenant.ResolverHeaderName())
	}
}
```

- [ ] **Step 2: 在 `api/middlewares/middlewares.go` 的 `Module` 中追加 provider（不进入全局 `NewMiddlewares` 列表）**

```go
	fx.Provide(NewTenantMiddleware),
	fx.Provide(NewMemberAuthMiddleware),
```

- [ ] **Step 3: 编译验证（依赖 Task 13，可一起编译）**

Run: `cd light-admin-server && go build ./api/middlewares/...`
Expected: 待 Task 13 完成后通过。

---

### Task 13: 会员鉴权中间件

**Files:**
- Create: `api/middlewares/member_auth_middleware.go`

- [ ] **Step 1: 新建 `api/middlewares/member_auth_middleware.go`**

```go
package middlewares

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	memberService "github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/pkg/echox"
)

// MemberAuthMiddleware 校验会员 JWT，注入 constants.CurrentMember
type MemberAuthMiddleware struct {
	logger      lib.Logger
	authService memberService.MemberAuthService
}

func NewMemberAuthMiddleware(logger lib.Logger, authService memberService.MemberAuthService) MemberAuthMiddleware {
	return MemberAuthMiddleware{logger: logger, authService: authService}
}

// Require 返回需要会员登录的路由组级中间件
func (a MemberAuthMiddleware) Require() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			auth := ctx.Request().Header.Get("Authorization")
			prefix := "Bearer "
			var token string
			if auth != "" && strings.HasPrefix(auth, prefix) {
				token = auth[len(prefix):]
			}

			claims, err := a.authService.ParseToken(token)
			if err != nil {
				return echox.Response{Code: http.StatusUnauthorized, Message: err}.JSON(ctx)
			}

			ctx.Set(constants.CurrentMember, claims)
			// 以 token 内租户为准
			ctx.Set(constants.CurrentTenantID, claims.TenantID)
			return next(ctx)
		}
	}
}
```

- [ ] **Step 2: 编译验证**

Run: `cd light-admin-server && go build ./api/middlewares/...`
Expected: 通过。

- [ ] **Step 3: Commit**

```bash
git add light-admin-server/api/middlewares/tenant_middleware.go light-admin-server/api/middlewares/member_auth_middleware.go light-admin-server/api/middlewares/middlewares.go
git commit -m "feat(middleware): add tenant resolution and member auth middlewares"
```

---

## Phase 5 — 控制器、路由、FX 装配

### Task 14: 会员鉴权控制器

**Files:**
- Create: `api/member/controller/controller.go`
- Create: `api/member/controller/member_auth_controller.go`

- [ ] **Step 1: 新建 FX 聚合 `api/member/controller/controller.go`**

```go
package controller

import "go.uber.org/fx"

// Module exports member controllers
var Module = fx.Options(
	fx.Provide(NewMemberAuthController),
	fx.Provide(NewTenantController),
	fx.Provide(NewMemberController),
)
```

- [ ] **Step 2: 新建 `api/member/controller/member_auth_controller.go`**

```go
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/pkg/echox"
)

type MemberAuthController struct {
	memberService service.MemberService
	authService   service.MemberAuthService
	logger        lib.Logger
}

func NewMemberAuthController(memberService service.MemberService, authService service.MemberAuthService, logger lib.Logger) MemberAuthController {
	return MemberAuthController{memberService: memberService, authService: authService, logger: logger}
}

func currentTenantID(ctx echo.Context) string {
	v, _ := ctx.Get(constants.CurrentTenantID).(string)
	return v
}

func currentMember(ctx echo.Context) *dto.MemberClaims {
	v, _ := ctx.Get(constants.CurrentMember).(*dto.MemberClaims)
	return v
}

// Register @router /api/v1/member/auth/register [post]
func (a MemberAuthController) Register(ctx echo.Context) error {
	form := new(dto.MemberRegister)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	m, err := a.memberService.Register(currentTenantID(ctx), form)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK, Data: m}.JSON(ctx)
}

// Login @router /api/v1/member/auth/login [post]
func (a MemberAuthController) Login(ctx echo.Context) error {
	form := new(dto.MemberLogin)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	tenantID := currentTenantID(ctx)
	m, err := a.memberService.Verify(tenantID, form.Username, form.Password)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	resp, err := a.authService.GenerateToken(m)
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}
	a.memberService.RecordLogin(tenantID, m.ID, ctx.RealIP())
	return echox.Response{Code: http.StatusOK, Data: resp}.JSON(ctx)
}

// Profile @router /api/v1/member/profile [get]
func (a MemberAuthController) Profile(ctx echo.Context) error {
	claims := currentMember(ctx)
	if claims == nil {
		return echox.Response{Code: http.StatusUnauthorized, Message: errors.MemberTokenInvalid}.JSON(ctx)
	}
	m, err := a.memberService.GetProfile(claims.TenantID, claims.ID)
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK, Data: m}.JSON(ctx)
}

// UpdateProfile @router /api/v1/member/profile [put]
func (a MemberAuthController) UpdateProfile(ctx echo.Context) error {
	claims := currentMember(ctx)
	if claims == nil {
		return echox.Response{Code: http.StatusUnauthorized, Message: errors.MemberTokenInvalid}.JSON(ctx)
	}
	form := new(dto.MemberProfileForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.memberService.UpdateProfile(claims.TenantID, claims.ID, form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

// Logout @router /api/v1/member/auth/logout [delete]
func (a MemberAuthController) Logout(ctx echo.Context) error {
	// 无服务端会话，前端丢弃 token 即可
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
```

- [ ] **Step 3: 编译验证（依赖 Task 15）**

Run: `cd light-admin-server && go build ./api/member/controller/...`
Expected: 待 Task 15 完成后通过。

---

### Task 15: 后台租户/会员管理控制器

**Files:**
- Create: `api/member/controller/tenant_controller.go`
- Create: `api/member/controller/member_controller.go`

- [ ] **Step 1: 新建 `api/member/controller/tenant_controller.go`**

```go
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/constants"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/tenant"
	"github.com/top-system/light-admin/pkg/echox"
)

type TenantController struct {
	tenantService service.TenantService
	logger        lib.Logger
}

func NewTenantController(tenantService service.TenantService, logger lib.Logger) TenantController {
	return TenantController{tenantService: tenantService, logger: logger}
}

func operatorID(ctx echo.Context) string {
	if claims, ok := ctx.Get(constants.CurrentUser).(*dto.JwtClaims); ok && claims != nil {
		return claims.ID
	}
	return ""
}

func (a TenantController) Query(ctx echo.Context) error {
	param := new(tenant.TenantQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	result, err := a.tenantService.Query(param)
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}
	return echox.Response{
		Code: http.StatusOK,
		Data: result.List,
		Page: &echox.PageInfo{Total: result.Pagination.Total, PageNum: result.Pagination.PageNum, PageSize: result.Pagination.PageSize},
	}.JSON(ctx)
}

func (a TenantController) Create(ctx echo.Context) error {
	form := new(tenant.TenantForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	id, err := a.tenantService.Create(form, operatorID(ctx))
	if err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK, Data: map[string]string{"id": id}}.JSON(ctx)
}

func (a TenantController) Update(ctx echo.Context) error {
	id := ctx.Param("id")
	form := new(tenant.TenantForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.tenantService.Update(id, form, operatorID(ctx)); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

func (a TenantController) Delete(ctx echo.Context) error {
	if err := a.tenantService.Delete(ctx.Param("id")); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
```

> 已核实 `dto.JwtClaims` 含 `ID` 字段（`models/dto`）。

- [ ] **Step 2: 新建 `api/member/controller/member_controller.go`**

```go
package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/top-system/light-admin/api/member/service"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/member"
	"github.com/top-system/light-admin/pkg/echox"
)

type MemberController struct {
	memberService service.MemberService
	logger        lib.Logger
}

func NewMemberController(memberService service.MemberService, logger lib.Logger) MemberController {
	return MemberController{memberService: memberService, logger: logger}
}

func (a MemberController) Query(ctx echo.Context) error {
	param := new(member.MemberQueryParam)
	if err := ctx.Bind(param); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	result, err := a.memberService.Query(param)
	if err != nil {
		return echox.Response{Code: http.StatusInternalServerError, Message: err}.JSON(ctx)
	}
	return echox.Response{
		Code: http.StatusOK,
		Data: result.List,
		Page: &echox.PageInfo{Total: result.Pagination.Total, PageNum: result.Pagination.PageNum, PageSize: result.Pagination.PageSize},
	}.JSON(ctx)
}

type statusForm struct {
	TenantID string `json:"tenantId" validate:"required"`
	Status   int    `json:"status"`
}

func (a MemberController) SetStatus(ctx echo.Context) error {
	id := ctx.Param("id")
	form := new(statusForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.memberService.SetStatus(form.TenantID, id, form.Status); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}

type resetPwdForm struct {
	TenantID string `json:"tenantId" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

func (a MemberController) ResetPassword(ctx echo.Context) error {
	id := ctx.Param("id")
	form := new(resetPwdForm)
	if err := ctx.Bind(form); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	if err := a.memberService.ResetPassword(form.TenantID, id, form.Password); err != nil {
		return echox.Response{Code: http.StatusBadRequest, Message: err}.JSON(ctx)
	}
	return echox.Response{Code: http.StatusOK}.JSON(ctx)
}
```

- [ ] **Step 3: 编译验证**

Run: `cd light-admin-server && go build ./api/member/...`
Expected: 通过。

- [ ] **Step 4: Commit**

```bash
git add light-admin-server/api/member/controller/
git commit -m "feat(member): add member auth/tenant/member admin controllers"
```

---

### Task 16: 路由、域模块聚合、全局注册

**Files:**
- Create: `api/member/route/route.go`
- Create: `api/member/route/member_auth_route.go`
- Create: `api/member/route/tenant_route.go`
- Create: `api/member/route/member_route.go`
- Create: `api/member/module.go`
- Modify: `api/module.go`

- [ ] **Step 1: 新建 `api/member/route/member_auth_route.go`**

```go
package route

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type MemberAuthRoutes struct {
	logger     lib.Logger
	handler    lib.HttpHandler
	controller controller.MemberAuthController
	tenantMw   middlewares.TenantMiddleware
	memberMw   middlewares.MemberAuthMiddleware
}

func NewMemberAuthRoutes(
	logger lib.Logger,
	handler lib.HttpHandler,
	c controller.MemberAuthController,
	tenantMw middlewares.TenantMiddleware,
	memberMw middlewares.MemberAuthMiddleware,
) MemberAuthRoutes {
	return MemberAuthRoutes{logger: logger, handler: handler, controller: c, tenantMw: tenantMw, memberMw: memberMw}
}

func (a MemberAuthRoutes) Setup() {
	// 公开（仅需租户解析）
	auth := a.handler.RouterV1.Group("/member/auth", a.tenantMw.Resolve())
	{
		auth.POST("/register", a.controller.Register)
		auth.POST("/login", a.controller.Login)
		auth.DELETE("/logout", a.controller.Logout, a.memberMw.Require())
	}

	// 需要会员登录
	profile := a.handler.RouterV1.Group("/member/profile", a.memberMw.Require())
	{
		profile.GET("", a.controller.Profile)
		profile.PUT("", a.controller.UpdateProfile)
	}
}
```

- [ ] **Step 2: 新建 `api/member/route/tenant_route.go`（后台管理，走全局 admin 鉴权 + Casbin 权限）**

```go
package route

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type TenantRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	controller     controller.TenantController
	permMiddleware middlewares.PermissionMiddleware
}

func NewTenantRoutes(logger lib.Logger, handler lib.HttpHandler, c controller.TenantController, perm middlewares.PermissionMiddleware) TenantRoutes {
	return TenantRoutes{logger: logger, handler: handler, controller: c, permMiddleware: perm}
}

func (a TenantRoutes) Setup() {
	api := a.handler.RouterV1.Group("/tenants")
	{
		api.GET("", a.controller.Query, a.permMiddleware.RequirePerm("member:tenant:query"))
		api.POST("", a.controller.Create, a.permMiddleware.RequirePerm("member:tenant:add"))
		api.PUT("/:id", a.controller.Update, a.permMiddleware.RequirePerm("member:tenant:edit"))
		api.DELETE("/:id", a.controller.Delete, a.permMiddleware.RequirePerm("member:tenant:delete"))
	}
}
```

> 注意：`/api/v1/tenants` 与 `/api/v1/members` 不在 member 忽略前缀内（忽略的是 `/api/v1/member`，无复数 s），因此仍受全局 admin 鉴权与 Casbin 保护——符合"平台管理员管理"的定位。

- [ ] **Step 3: 新建 `api/member/route/member_route.go`**

```go
package route

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/lib"
)

type MemberAdminRoutes struct {
	logger         lib.Logger
	handler        lib.HttpHandler
	controller     controller.MemberController
	permMiddleware middlewares.PermissionMiddleware
}

func NewMemberAdminRoutes(logger lib.Logger, handler lib.HttpHandler, c controller.MemberController, perm middlewares.PermissionMiddleware) MemberAdminRoutes {
	return MemberAdminRoutes{logger: logger, handler: handler, controller: c, permMiddleware: perm}
}

func (a MemberAdminRoutes) Setup() {
	api := a.handler.RouterV1.Group("/members")
	{
		api.GET("", a.controller.Query, a.permMiddleware.RequirePerm("member:member:query"))
		api.PUT("/:id/status", a.controller.SetStatus, a.permMiddleware.RequirePerm("member:member:edit"))
		api.PUT("/:id/password", a.controller.ResetPassword, a.permMiddleware.RequirePerm("member:member:edit"))
	}
}
```

- [ ] **Step 4: 新建路由聚合 `api/member/route/route.go`**

```go
package route

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewMemberAuthRoutes),
	fx.Provide(NewTenantRoutes),
	fx.Provide(NewMemberAdminRoutes),
	fx.Provide(NewRoutes),
)

type Routes []Route

type Route interface{ Setup() }

func NewRoutes(
	memberAuth MemberAuthRoutes,
	tenant TenantRoutes,
	memberAdmin MemberAdminRoutes,
) Routes {
	return Routes{memberAuth, tenant, memberAdmin}
}

func (a Routes) Setup() {
	for _, r := range a {
		r.Setup()
	}
}
```

- [ ] **Step 5: 新建域模块 `api/member/module.go`**

```go
package member

import (
	"github.com/top-system/light-admin/api/member/controller"
	"github.com/top-system/light-admin/api/member/repository"
	"github.com/top-system/light-admin/api/member/route"
	"github.com/top-system/light-admin/api/member/service"

	"go.uber.org/fx"
)

// Module exports member domain module
var Module = fx.Options(
	controller.Module,
	service.Module,
	repository.Module,
	route.Module,
)
```

- [ ] **Step 6: 修改 `api/module.go` 注册 member 模块（完整目标文件）**

```go
package api

import (
	"github.com/top-system/light-admin/api/member"
	memberRoute "github.com/top-system/light-admin/api/member/route"
	"github.com/top-system/light-admin/api/middlewares"
	"github.com/top-system/light-admin/api/platform"
	platformRoute "github.com/top-system/light-admin/api/platform/route"
	"github.com/top-system/light-admin/api/system"
	systemRoute "github.com/top-system/light-admin/api/system/route"

	"go.uber.org/fx"
)

var Module = fx.Options(
	middlewares.Module,
	system.Module,
	platform.Module,
	member.Module,
	fx.Provide(NewRoutes),
)

type Routes struct {
	System   systemRoute.Routes
	Platform platformRoute.Routes
	Member   memberRoute.Routes
}

func NewRoutes(
	system systemRoute.Routes,
	platform platformRoute.Routes,
	member memberRoute.Routes,
) Routes {
	return Routes{System: system, Platform: platform, Member: member}
}

func (r Routes) Setup() {
	r.System.Setup()
	r.Platform.Setup()
	r.Member.Setup()
}
```

- [ ] **Step 7: 全量编译 + vet + 启动验证**

Run: `cd light-admin-server && go build ./... && go vet ./...`
Expected: 编译通过。

Run（确保已 migrate+setup）: `cd light-admin-server && go run . runserver --config=./config/config.yaml --casbin_model=./config/casbin_model.conf`
Expected: 服务启动无 FX 依赖报错（日志显示监听端口）。

- [ ] **Step 8: Commit**

```bash
git add light-admin-server/api/member/route/ light-admin-server/api/member/module.go light-admin-server/api/module.go
git commit -m "feat(member): wire member routes and register domain module"
```

---

## Phase 6 — 端到端 API 冒烟

### Task 17: 端到端验证

**Files:** 无（验证步骤；端口按 config，示例用 9999）

- [ ] **Step 1: 注册会员（多租户开启，租户码走 header）**

Run:
```bash
curl -s -X POST http://localhost:9999/api/v1/member/auth/register \
  -H 'Content-Type: application/json' -H 'X-Tenant-Code: default' \
  -d '{"username":"alice","password":"secret123","nickname":"Alice"}'
```
Expected: `code=00000`，返回会员对象（无 password 字段）。

- [ ] **Step 2: 重复注册应冲突**

Run: 同上再执行一次。
Expected: `member already exists`，HTTP 409。

- [ ] **Step 3: 登录拿 token**

Run:
```bash
curl -s -X POST http://localhost:9999/api/v1/member/auth/login \
  -H 'Content-Type: application/json' -H 'X-Tenant-Code: default' \
  -d '{"username":"alice","password":"secret123"}'
```
Expected: 返回 `accessToken`。

- [ ] **Step 4: 带 token 取资料**

Run:
```bash
curl -s http://localhost:9999/api/v1/member/profile -H "Authorization: Bearer <token>"
```
Expected: 返回 alice 资料。

- [ ] **Step 5: 未知租户码注册应失败**

Run: 把 `X-Tenant-Code` 换成 `nope` 再注册。
Expected: `tenant not found`，HTTP 404。

- [ ] **Step 6: 关闭多租户回归**

把 `config.yaml` 的 `MultiTenant.Enable` 改为 `false`，重启服务，不带 `X-Tenant-Code` 注册/登录 `bob`。
Expected: 成功，且数据落到 default 租户。验证后改回 `Enable: true` 并重启。

- [ ] **Step 7: 全量测试**

Run: `cd light-admin-server && go test ./... -race`
Expected: 全绿。

---

## Phase 7 — 菜单与权限种子

### Task 18: 菜单与 Casbin 权限

**Files:**
- Modify: `config/menu.yaml`

- [ ] **Step 1: 打开 `config/menu.yaml`，定位「系统管理」块作为结构范式**

Run: `grep -n "系统管理\|catalog\|button" light-admin-server/config/menu.yaml | head`
Expected: 找到目录/菜单/按钮的层级写法。

- [ ] **Step 2: 在 `config/menu.yaml` 仿照系统管理块，新增「会员中心」目录及两个菜单 + 按钮权限**

逐字段对齐既有缩进，新增：
- 目录：会员中心（path `/member`）
- 菜单：租户管理（path `/member/tenant`，组件 `member/tenant`），按钮：`member:tenant:query` / `member:tenant:add` / `member:tenant:edit` / `member:tenant:delete`
- 菜单：会员管理（path `/member/list`，组件 `member/list`），按钮：`member:member:query` / `member:member:edit`

- [ ] **Step 3: 重新 setup 导入菜单**

Run: `cd light-admin-server && go run . setup --config=./config/config.yaml --menu=./config/menu.yaml`
Expected: 菜单导入成功。先确认 setup 的菜单导入是覆盖式还是追加式（读 `cmd/setup/setup.go` 与 `menuService.CreateMenus`）；若覆盖式则整份 menu.yaml 都会重建，确保未破坏既有节点。

- [ ] **Step 4: 给默认角色/超级管理员分配新权限**（后台「角色管理」勾选，或在 setup 默认角色种子里包含 `member:*`）。超级管理员若是 `*:*:*` 则自动拥有。

- [ ] **Step 5: Commit**

```bash
git add light-admin-server/config/menu.yaml
git commit -m "feat(menu): add member center menus and permissions"
```

---

## Phase 8 — 前端（后台两页面 + API）

### Task 19: 前端 API 封装

**Files:**
- Create: `light-admin-ui/src/services/member/tenant.ts`
- Create: `light-admin-ui/src/services/member/member.ts`

- [ ] **Step 1: 先读现有 service 范式**

Run: `ls light-admin-ui/src/services && grep -rn "request(" light-admin-ui/src/services | head`
Expected: 确认 `request` 导入来源与返回包装（是否统一解包 `data`、是否有泛型 `BaseResponse<T>`）。按其风格对齐下面两个文件。

- [ ] **Step 2: 新建 `src/services/member/tenant.ts`**

```ts
import { request } from '@umijs/max';

export interface TenantItem {
  id: string;
  code: string;
  name: string;
  status: number;
  createTime?: string;
}

export async function queryTenants(params: any) {
  return request('/api/v1/tenants', { method: 'GET', params });
}
export async function createTenant(data: Partial<TenantItem>) {
  return request('/api/v1/tenants', { method: 'POST', data });
}
export async function updateTenant(id: string, data: Partial<TenantItem>) {
  return request(`/api/v1/tenants/${id}`, { method: 'PUT', data });
}
export async function deleteTenant(id: string) {
  return request(`/api/v1/tenants/${id}`, { method: 'DELETE' });
}
```

- [ ] **Step 3: 新建 `src/services/member/member.ts`**

```ts
import { request } from '@umijs/max';

export interface MemberItem {
  id: string;
  tenantId: string;
  username: string;
  nickname: string;
  email: string;
  mobile: string;
  status: number;
  createTime?: string;
}

export async function queryMembers(params: any) {
  return request('/api/v1/members', { method: 'GET', params });
}
export async function setMemberStatus(id: string, data: { tenantId: string; status: number }) {
  return request(`/api/v1/members/${id}/status`, { method: 'PUT', data });
}
export async function resetMemberPassword(id: string, data: { tenantId: string; password: string }) {
  return request(`/api/v1/members/${id}/password`, { method: 'PUT', data });
}
```

- [ ] **Step 4: 类型检查**

Run: `cd light-admin-ui && npm run tsc`
Expected: 无新增类型错误。

- [ ] **Step 5: Commit**

```bash
git add light-admin-ui/src/services/member/
git commit -m "feat(ui): add tenant/member service APIs"
```

---

### Task 20: 租户管理 + 会员管理页面 + 路由

**Files:**
- Create: `light-admin-ui/src/pages/member/tenant/index.tsx`
- Create: `light-admin-ui/src/pages/member/list/index.tsx`
- Modify: `light-admin-ui/config/routes.ts`

- [ ] **Step 1: 读现有 CRUD 页面范式**

Run: `ls light-admin-ui/src/pages/system && sed -n '1,80p' light-admin-ui/src/pages/system/dept/index.tsx`
Expected: 找到 `ProTable` / `ModalForm` / `ProColumns` 与分页字段（请求返回如何映射 `{ data, total, success }`）的实际写法，作为复制模板。

- [ ] **Step 2: 新建 `src/pages/member/tenant/index.tsx`（基于范式改造）**

要点（严格复制范式结构，再替换字段/接口）：
- `ProTable<TenantItem>` 列：`code`、`name`、`status`（valueEnum：1=正常,0=禁用）、`createTime`、操作（编辑/删除）。
- 工具栏「新建」→ `ModalForm`，字段 `code`、`name`、`status`。
- `request` 调 `queryTenants(params)`；把后端 `{ data: list, page: { total } }` 映射为 ProTable 需要的 `{ data: list, total, success: true }`（按 Step 1 实测的响应结构调整取值路径）。
- 新建/编辑提交分别调 `createTenant` / `updateTenant(id, ...)`；删除调 `deleteTenant(id)`，成功后 `actionRef.current?.reload()`。

- [ ] **Step 3: 新建 `src/pages/member/list/index.tsx`**

要点：
- `ProTable<MemberItem>` 列：`tenantId`、`username`、`nickname`、`email`、`mobile`、`status`、`createTime`、操作（禁用/启用、重置密码）。
- 筛选项：`tenantId`、`keywords`、`status`。
- 「禁用/启用」调 `setMemberStatus(id, { tenantId: row.tenantId, status: nextStatus })`。
- 「重置密码」用 `ModalForm` 收集 `password`，调 `resetMemberPassword(id, { tenantId: row.tenantId, password })`。

- [ ] **Step 4: 在 `config/routes.ts` 注册路由（与 menu 对齐）**

```ts
{
  path: '/member',
  name: 'member',
  routes: [
    { path: '/member/tenant', name: 'tenant', component: './member/tenant' },
    { path: '/member/list', name: 'list', component: './member/list' },
  ],
},
```

> 若本项目走「动态菜单 + 约定式组件路径」由后端菜单 `component` 字段驱动，则确认 `routes.ts` 是否仍需手工登记，按 system 页面现有做法一致处理（读一个 system 路由项确认）。

- [ ] **Step 5: 类型检查 + 构建**

Run: `cd light-admin-ui && npm run tsc && npm run build`
Expected: 通过。

- [ ] **Step 6: 前后端联调**

启动后端与前端，登录后台 → 「会员中心 / 租户管理」新建租户 → 「会员管理」查看/筛选/禁用/重置密码。
Expected: 列表、分页、增删改、禁用、重置密码均正常。

- [ ] **Step 7: Commit**

```bash
git add light-admin-ui/src/pages/member/ light-admin-ui/config/routes.ts
git commit -m "feat(ui): add tenant and member management pages"
```

---

## Self-Review 记录

- **Spec 覆盖**：隔离模式 A（Task 5/8 scope）、配置开关（Task 1）、关闭退化（Task 12 Resolve 分支 + Task 6 default 种子）、独立会员 JWT（Task 9/13）、注册带租户标识（Task 12 extractCode）、后台不动（仅新增 member 模块，未改 system）、租户/会员管理（Task 15/16/20）、安全（Task 11 服务端强制 tenant_id、防枚举登录、密码 bcrypt + json:"-"）、迁移与种子（Task 6）、前端两页面（Task 19/20）—— 均有对应任务。
- **类型一致性**：`MemberClaims`(Task 4) 在 Task 9/13/14 一致；`scopes.Tenant`(Task 5) 在 Task 8 一致；分页 helper 统一为 `dtoPagination`（Task 7 定义、Task 8 复用，无 `dtoPagination2`）；`MultiTenantConfig` 方法 `IsEnabled/ResolverHeaderName/FallbackTenantCode`(Task 1) 在 Task 12 一致调用。
- **需实施者现场确认的点（已在步骤内给出"参照现有 X 文件对齐"的明确指引，非占位符）**：①`dto.DateTime` 底层类型（影响 `nowDateTime` 与 `UpdateLoginInfo`，已给降级方案）；②`models/dto` 分页/排序导出名；③前端 `request` 返回包装与 ProTable 分页映射；④menu.yaml setup 覆盖式/追加式；⑤前端路由是否由动态菜单驱动。
