<div align="center">
  <img src="docs/logo.png" width="180" height="180" alt="Light Admin Logo" />

  <h1>Light Admin</h1>

  <p>
    <strong>轻量、优雅的中后台管理系统解决方案</strong>
  </p>

  <p>
    基于 Echo + sqlc + Casbin + Uber-FX 构建的 RBAC 权限管理脚手架
  </p>

  <p>
    <a href="https://github.com/top-system/light-admin/blob/main/README.en.md">English</a> | 简体中文
  </p>

  <p>
    <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go" alt="Go Version" />
    <img src="https://img.shields.io/badge/Echo-4.11+-00ADD8?style=flat-square" alt="Echo Version" />
    <img src="https://img.shields.io/badge/sqlc-1.31+-red?style=flat-square" alt="sqlc Version" />
    <img src="https://img.shields.io/badge/Casbin-2.77+-brightgreen?style=flat-square" alt="Casbin Version" />
    <img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License" />
  </p>
</div>

---

## ✨ 特性

### 核心功能
- 🔐 **用户认证** - JWT Token 认证，支持 Token 刷新
- 👥 **用户管理** - 用户增删改查、状态管理、密码重置
- 🎭 **角色管理** - 灵活的角色配置，支持多角色
- 📋 **菜单管理** - 动态菜单配置，支持多级菜单
- 🏢 **部门管理** - 树形组织架构管理
- 🔑 **权限控制** - 基于 perm 标识的 RBAC 访问控制，支持缓存加速
- 📝 **操作日志** - 完整的操作审计日志
- 📢 **通知公告** - 系统通知与公告管理
- ⚙️ **系统配置** - 动态系统参数配置
- 📚 **字典管理** - 数据字典维护

### 扩展功能
- 📤 **文件上传** - 支持本地存储、MinIO、阿里云 OSS
- ⏰ **定时任务** - 灵活的 Cron 定时任务调度
- 📥 **任务队列** - 异步任务处理，支持重试、持久化、状态恢复
- ⬇️ **下载管理** - 集成 aria2/qBittorrent 下载器，与队列系统深度整合
- 🔌 **WebSocket** - 基于 STOMP 协议的实时通信，支持广播和点对点消息

### 技术特性
- 🚀 **高性能** - 基于 Echo 框架，高效路由匹配
- 📦 **依赖注入** - 基于 Uber-FX 的依赖注入
- 📖 **API 文档** - 集成 Swagger 自动生成 API 文档
- 🔧 **模块化** - 清晰的代码结构，易于扩展
- 🛡️ **安全性** - 完善的安全中间件支持
- 💾 **多数据库** - 支持 MySQL、PostgreSQL、SQLite
- 🗄️ **多缓存** - 支持 Redis 和内存缓存

---

## 📁 项目结构

```
light-admin/
├── api/                    # API 层
│   ├── middlewares/        # 中间件
│   ├── platform/           # 平台模块 (文件上传、WebSocket等)
│   └── system/             # 系统模块 (用户、角色、菜单等)
├── bootstrap/              # 应用启动
├── cmd/                    # 命令行入口
├── config/                 # 配置文件
├── docs/                   # 文档 & Swagger
├── errors/                 # 错误定义
├── lib/                    # 核心库
├── models/                 # 数据模型
│   ├── database/           # 数据库模型基类
│   ├── dto/                # 数据传输对象
│   ├── platform/           # 平台模块模型
│   └── system/             # 系统模块模型
├── pkg/                    # 工具包
│   ├── crontab/            # 定时任务
│   ├── downloader/         # 下载器 (aria2/qBittorrent)
│   ├── queue/              # 任务队列
│   ├── websocket/          # WebSocket (STOMP协议)
│   └── ...                 # 其他工具
└── tests/                  # 测试文件
```

---

## 🚀 快速开始

### 环境要求

- Go 1.21+
- Node.js 16+ (前端)
- 可选：MySQL 5.7+ / PostgreSQL 12+ / SQLite 3
- 可选：Redis 6.0+（不配置则使用内存缓存）

### 安装

```bash
# 克隆项目
git clone https://github.com/top-system/light-admin.git
cd light-admin

# 复制配置文件
cp config/config.yaml.default config/config.yaml

# 修改配置文件（默认使用 SQLite，开箱即用）
vim config/config.yaml

# 初始化数据库
go run . migrate

# 初始化菜单数据
go run . setup

# 启动服务
go run .
```

### 使用 Docker

```bash
# 构建镜像
docker build -t light-admin .

# 运行容器
docker run -d -p 2222:2222 \
  -v ./config:/app/config \
  -v ./data:/app/data \
  light-admin
```

---

## 📖 文档

| 文档 | 说明 |
|------|------|
| [API 文档](docs/swagger.yaml) | Swagger API 文档 |
| [任务队列](docs/queue.md) | 异步任务队列使用指南 |
| [定时任务](docs/crontab.md) | 定时任务配置指南 |
| [下载器](docs/downloader.md) | aria2/qBittorrent 集成指南 |
| [WebSocket](docs/websocket.md) | 实时通信使用指南 |

---

## ⚙️ 配置说明

### 基础配置（SQLite + 内存缓存，零依赖）

```yaml
Name: light-admin

HTTP:
  Host: 0.0.0.0
  Port: 2222

# SQLite 数据库（开箱即用）
Database:
  Engine: sqlite
  Name: ./data/app.db
  TablePrefix: t
  MaxLifetime: 7200
  MaxOpenConns: 1
  MaxIdleConns: 1

# 内存缓存（无需 Redis）
Cache:
  Type: memory
  KeyPrefix: app
```

### MySQL + Redis 配置

```yaml
# 要求 MySQL 8.0.13+（迁移使用表达式默认值）
Database:
  Engine: mysql
  Host: 127.0.0.1
  Port: 3306
  Name: light_admin
  Username: root
  Password: your_password

Cache:
  Type: redis
  Host: 127.0.0.1
  Port: 6379
  Password: ""
```

### PostgreSQL 配置

```yaml
# 要求 PostgreSQL 13+
Database:
  Engine: postgres
  Host: 127.0.0.1
  Port: 5432
  Name: light_admin
  Username: postgres
  Password: your_password
```

> 三种引擎共享同一套业务代码与 API:切换引擎只需改 `Database.Engine` 并重新
> `migrate` + `setup`(跨库不迁移数据)。SQLite 内嵌零依赖,适合本地开发与轻量部署。

### 扩展功能配置

```yaml
# 任务队列
Queue:
  Enable: true
  Name: "default"
  WorkerNum: 2
  MaxRetry: 3

# 下载器
Downloader:
  Enable: true
  Type: aria2  # aria2 或 qbittorrent
  Aria2:
    Server: "http://localhost:6800/jsonrpc"
    Token: ""
    TempPath: "./downloads"
```

---

## 🛠️ 开发命令

```bash
# 编译
go build -o light-admin .

# 运行
go run .

# 数据库迁移
go run . migrate

# 初始化数据
go run . setup

# 生成 Swagger 文档
swag init

# 运行测试
go test ./...
```

---

## 🗺️ 路线图

- [x] 用户认证与权限管理
- [x] 动态菜单与角色管理
- [x] 部门与组织架构
- [x] 系统配置与字典
- [x] 文件上传 (本地/OSS)
- [x] 异步任务队列
- [x] 定时任务调度
- [x] 下载器集成 (与队列深度整合)
- [x] WebSocket 实时通信
- [x] 权限缓存优化
- [x] SQLite 支持
- [x] 操作日志审计完善
- [ ] 工作流引擎
- [ ] 消息推送
- [ ] 数据导入导出

---

## 🤝 贡献指南

欢迎提交 PR 和 Issue！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

---

## 📄 开源协议

本项目采用 [MIT](LICENSE) 开源协议。

---

## 🔗 相关链接

- [前端项目](https://github.com/top-system/light-admin-ui)
