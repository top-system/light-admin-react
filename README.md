# Light Admin React

`Light Admin React` 是一套前后端分离的轻量级管理后台示例代码，包含基于 React 的管理端界面，以及基于 Go 的后台服务。

这个仓库当前由两个独立项目组成：

- `light-admin-ui`：管理后台前端
- `light-admin-server`：管理后台后端

项目目标不是保留通用脚手架形态，而是收敛成一套可以直接承载业务开发的后台基础模板，已经接入了动态菜单、RBAC 权限、字典、公告、文件上传、任务队列、下载器和 WebSocket 等能力。

## 项目结构

```text
light-admin-react/
├─ light-admin-ui/       # React + Umi Max + Ant Design 前端
└─ light-admin-server/   # Go + Echo + GORM 后端
```

## 技术栈

### 前端

- React 19
- `@umijs/max`
- Ant Design 6 / Pro Components
- TypeScript
- Tailwind CSS 4
- React Query
- TinyMCE

### 后端

- Go 1.24
- Echo
- GORM
- JWT
- Swagger
- Redis / Memory Cache
- MySQL / PostgreSQL / SQLite
- WebSocket

## 主要能力

### 前端页面

- 登录页与鉴权跳转
- 首页
- 用户资料页
- 系统管理：用户、角色、菜单、部门、字典、参数、日志、公告、队列、下载器
- 组件示例：CRUD、上传、富文本、拖拽、图标选择、表格选择、滚动文本等
- 功能示例：WebSocket、字典同步、单表 CRUD、路由参数、多级菜单

### 后端能力

- JWT 登录认证
- RBAC 权限控制
- 动态菜单
- 用户/角色/部门/菜单/字典/参数管理
- 文件上传
- 公告通知
- 任务队列
- 下载器集成
- WebSocket 推送
- Swagger API 文档

## 快速开始

### 1. 启动后端

进入后端目录：

```bash
cd light-admin-server
```

首次运行需要准备配置文件：

```bash
cp config/config.yaml.default config/config.yaml
```

注意：

- `config/config.yaml.default` 里的数据库示例默认是 MySQL
- 如果你想最快跑起来，建议改成 SQLite
- 前端开发代理默认请求 `http://localhost:9999`
- 后端代码默认端口也是 `9999`
- 但示例配置文件里写的是 `2222`，本地联调时建议把 `config/config.yaml` 的 `HTTP.Port` 改成 `9999`

推荐的本地最小配置示例：

```yaml
Name: light-admin

HTTP:
  Host: 0.0.0.0
  Port: 9999

SuperAdmin:
  Username: root
  Realname: Super Admin
  Password: 123123

Captcha:
  Enable: false

Cache:
  Type: memory
  KeyPrefix: app

Database:
  Engine: sqlite
  Name: ./data/app.db
  TablePrefix: t
  MaxLifetime: 7200
  MaxOpenConns: 1
  MaxIdleConns: 1

OSS:
  Type: local
  Local:
    StoragePath: ./uploads
```

初始化数据库和菜单：

```bash
go run . migrate --config=./config/config.yaml
go run . setup --config=./config/config.yaml --menu=./config/menu.yaml
```

启动服务：

```bash
go run . runserver --config=./config/config.yaml --casbin_model=./config/casbin_model.conf
```

也可以直接使用：

```bash
make start
```

如果使用 `make start`，请确认 `config/config.yaml` 已经存在且端口配置正确。

### 2. 启动前端

进入前端目录：

```bash
cd light-admin-ui
```

安装依赖：

```bash
npm install
```

启动开发环境：

```bash
npm run dev
```

默认开发代理配置位于 `light-admin-ui/config/proxy.ts`，会将以下请求转发给后端：

- `/api/v1`
- `/ws`

前端启动后，通常可通过 `http://localhost:8000` 访问。

## 默认登录信息

如果你使用上面的推荐配置并执行了初始化命令，默认超级管理员账号为：

- 用户名：`root`
- 密码：`123123`

## 常用命令

### 前端

```bash
npm run dev
npm run build
npm run lint
npm run test
npm run tsc
```

### 后端

```bash
go run .
go run . migrate
go run . setup --menu=./config/menu.yaml
go test ./...
make start
make swagger
```

## 文档位置

- 后端 Swagger 文档：`light-admin-server/docs/swagger.yaml`
- 后端 WebSocket 文档：`light-admin-server/docs/websocket.md`
- 前端路由配置：`light-admin-ui/config/routes.ts`

## 说明

- 当前工作区中的 `light-admin-ui` 和 `light-admin-server` 原本各自是独立 git 仓库
- 这个根目录 README 用来描述整套前后端代码的组合使用方式
- 如果你要把它发布成一个新的 GitHub 仓库，建议按“整套代码仓库”统一维护，而不是继续作为两个嵌套仓库直接提交
