# Let's Encrypt 证书管理服务

基于 Go 语言构建的 Let's Encrypt SSL 证书自动化管理平台，通过 Cloudflare DNS 验证方式自动签发和续期证书，支持与 Nginx 容器的联动。内置响应式 Web 管理界面（毛玻璃风格 UI），支持桌面端和移动端，开箱即用。

## 功能特性

- **用户系统**：邮箱验证码注册、bcrypt 密码加密、Token 鉴权登录、24 小时会话保持
- **证书管理**：创建、查询、编辑、刷新、删除 Let's Encrypt 证书记录
- **自动续期**：检测证书剩余有效期，低于阈值自动触发续期
- **DNS 验证**：通过 Cloudflare DNS API 完成 ACME DNS-01 验证
- **Nginx 联动**：证书更新后自动通过 Docker API 重载 Nginx 配置（SIGHUP 信号）
- **SQLite 持久化**：使用 SQLite 数据库存储账户和证书信息（WAL 模式）
- **响应式 Web 界面**：内置 SPA，毛玻璃风格 UI，支持桌面端 / 平板 / 移动端自适应
- **移动端优化**：汉堡菜单侧边栏、底部导航栏、卡片式证书列表、底部抽屉弹窗
- **统一响应格式**：所有 API 响应均采用 `{success, timestamp, data}` 标准包装
- **权限隔离**：用户只能管理自己名下的证书记录
- **安全写文件**：证书写入采用临时文件 + 原子重命名方式，防止写入中断导致文件损坏
- **Docker Socket 集成**：通过 Docker API 与 Nginx 容器联动

## 技术栈

| 类别        | 技术                         | 说明                   |
| ----------- | ---------------------------- | ---------------------- |
| 语言        | Go 1.25+                     | 主开发语言             |
| Web 框架    | Gin v1.12                    | HTTP API 服务框架      |
| ORM         | GORM v1.31                   | 数据库操作             |
| 数据库      | SQLite (glebarez/sqlite)     | 轻量级嵌入式数据库     |
| ACME 客户端 | go-acme/lego/v4 v4.35        | Let's Encrypt 证书签发 |
| DNS 验证    | Cloudflare DNS               | ACME DNS-01 验证       |
| 容器化      | Docker / Docker Compose      | 部署与运行             |
| Docker SDK  | docker/docker v28            | Nginx 容器管理         |
| 数据校验    | go-playground/validator v10  | 请求参数校验           |
| ID 生成     | snowflake v0.3               | 分布式 ID 生成         |
| 密码加密    | bcrypt (golang.org/x/crypto) | 用户密码哈希           |
| 静态嵌入    | Go embed.FS                  | 前端页面内嵌二进制     |

## 项目结构

```
lets-encrypt/
├── main.go                              # 程序入口，HTTP 服务器启动（端口 8099）
├── go.mod / go.sum                      # Go 模块依赖
├── Dockerfile                           # Docker 镜像构建（多阶段构建）
├── docker-compose.yml                   # Docker Compose 编排
├── lets-encrypt.db                      # SQLite 数据库文件（自动生成）
├── kouleen.china@gmail.com.key          # 示例 RSA 私钥文件（注册时自动生成）
├── .github/workflows/
│   └── docker-image.yml                 # GitHub Actions: Docker 镜像自动构建与推送
├── internal/
│   ├── api/
│   │   └── acme_api.go                  # HTTP API 处理器（10 个接口）
│   ├── middleware/
│   │   └── auth.go                      # Bearer Token 鉴权中间件
│   ├── modle/
│   │   ├── acme_account.go             # 用户账户模型 + 请求结构体
│   │   ├── acme_encrypt.go             # 证书记录模型 + 请求/查询结构体
│   │   ├── acme_user.go                # ACME 核心逻辑（lego 证书签发 + 安全写文件）
│   │   └── base_page.go               # 分页响应基础结构
│   ├── repository/
│   │   ├── sqlite_repository.go        # SQLite 数据访问层（GORM CRUD）
│   │   ├── store_repository.go         # 内存缓存仓储（Token/验证码 TTL）
│   │   └── docker_repository.go        # Docker API 客户端（Nginx 重载）
│   ├── routes/
│   │   └── acme.go                      # 路由注册（公共 + 鉴权分组 + 页面路由）
│   ├── service/
│   │   └── acme_service.go             # 业务逻辑层（核心业务编排）
│   └── validator/
│       └── validator.go                 # 参数校验工具（邮箱正则 + 结构体校验）
├── pkg/util/
│   ├── acme_key.go                      # RSA 私钥 PEM 文件读写
│   ├── acme_verify.go                   # 证书过期检测（远程 TLS / 本地 PEM）
│   └── store.go                         # 线程安全自动过期 KV 缓存（ExpireMap）
├── static/
│   ├── embed.go                         # embed.FS 声明（静态文件嵌入）
│   └── index.html                       # 前端 SPA 页面（毛玻璃风格 + 响应式）
└── test/
    ├── lets-encrypt.go                  # Cloudflare DNS 证书签发独立测试
    ├── lets-encrypt_test.go             # 证书签发单元测试入口
    └── verify-encrypt_test.go            # 证书过期检测测试
```

## 快速开始

### 环境要求

- Go 1.25 或更高版本
- Cloudflare 账号及 API Token（用于 DNS 验证）
- SMTP 邮件服务（用于发送注册验证码）
- Docker + Docker Compose（用于容器化部署和 Nginx 联动）

### 1. 克隆项目

```bash
git clone https://github.com/kouleen/lets-encrypt.git
cd lets-encrypt
```

### 2. 配置环境变量

在运行前需配置以下环境变量：

```bash
# 邮件服务配置（用于用户注册验证码发送，必填）
export SEND_EMAIL="your-email@example.com"       # 发件邮箱地址
export SEND_PWD="your-email-password"             # 邮箱密码或授权码
export SMTP_SERVER="smtp.example.com"             # SMTP 服务器地址
export SMTP_PORT="465"                            # SMTP 端口

# Cloudflare API Token 说明：
# 此 Token 在创建/更新证书请求中通过 cipher 字段传入，无需设为全局环境变量
```

### 3. 本地运行

```bash
go run main.go
```

服务默认监听 `8099` 端口。启动后访问：

- **API 基础路径**：`http://localhost:8099/acme`
- **Web 管理界面**：`http://localhost:8099/acme/html`

### 4. Docker Compose 部署

项目已提供完整的 `docker-compose.yml`，包含环境变量和卷挂载配置：

```yaml
services:
  lets-encrypt:
    image: kouleen/lets-encrypt:latest
    container_name: lets-encrypt
    ports:
      - "8099:8099"
    environment:
      - SMTP_SERVER=${SMTP_SERVER}
      - SMTP_PORT=${SMTP_PORT}
      - SEND_EMAIL=${SEND_EMAIL}
      - SEND_PWD=${SEND_PWD}
    volumes:
      - /usr/bin/docker:/usr/bin/docker:ro
      - /var/run/docker.sock:/var/run/docker.sock
      - /home/nginx/certs:/app/nginx/certs
```

启动服务：

```bash
# 方式一：Docker Compose（推荐）
SMTP_SERVER=smtp.example.com \
SMTP_PORT=465 \
SEND_EMAIL=you@example.com \
SEND_PWD=your-password \
docker compose up -d

# 方式二：手动 docker run
docker build -t kouleen/lets-encrypt:latest .
docker run -d \
  --name lets-encrypt \
  -p 8099:8099 \
  -e SEND_EMAIL="you@example.com" \
  -e SEND_PWD="your-password" \
  -e SMTP_SERVER="smtp.example.com" \
  -e SMTP_PORT="465" \
  -v /usr/bin/docker:/usr/bin/docker:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /home/nginx/certs:/app/nginx/certs \
  kouleen/lets-encrypt:latest
```

## Web 管理界面

服务内置了一个精美的响应式单页面管理应用（SPA），访问地址：`GET /acme/html`

### 界面功能

| 页面      | 功能说明                                                   |
| --------- | ---------------------------------------------------------- |
| 登录/注册 | 邮箱验证码注册、密码登录、Token 自动持久化（localStorage） |
| 概览面板  | 证书总数、有效/即将到期/已过期统计、最近证书预览           |
| 证书管理  | 分页列表、域名搜索、创建/编辑/刷新/删除证书记录            |

### UI 设计特点

- **深色毛玻璃风格**：`backdrop-filter: blur(20px)` 玻璃质感
- **动态背景**：多层径向渐变 + 缓动动画（20 秒循环）
- **渐变主色调**：紫蓝色系 `#667eea → #764ba2`
- **响应式布局**：桌面端（侧边栏）/ 平板（图标模式）/ 移动端（底部导航 + 汉堡菜单）
- **移动端适配**：
  - ≤1024px：侧边栏收缩为 72px 图标模式
  - ≤768px：侧边栏隐藏，底部导航栏激活，证书列表转为卡片视图
  - ≤380px：统计卡片单列展示
- **Toast 通知**：右上角滑入动画，支持 success/error/info/warning
- **弹窗动画**：弹性缩放 + 遮罩淡入，移动端转为底部抽屉样式
- **状态徽章**：有效/异常带发光动画
- **剩余天数标签**：安全绿/警告黄/危险红三级预警

## API 文档

### 基础信息

- **Base URL**: `http://localhost:8099`
- **数据格式**: JSON
- **认证方式**: Bearer Token（登录后获取，有效期 24 小时）
- **传参方式**:
  - Header: `Authorization: Bearer <token>`
  - Query: `access_token=<token>`

### 统一响应格式

所有接口均采用统一的 JSON 响应包装格式：

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": <响应数据>
}
```

| 字段      | 类型    | 说明                                 |
| --------- | ------- | ------------------------------------ |
| success   | boolean | 请求是否成功                         |
| timestamp | int64   | 服务器响应时间戳（毫秒级 Unix 时间） |
| data      | any     | 业务数据，类型由具体接口决定         |

**错误响应示例**：

```json
{
  "success": false,
  "timestamp": 1700000000000,
  "data": "错误描述信息"
}
```

**HTTP 状态码说明**：

| 状态码 | 说明                                                             |
| ------ | ---------------------------------------------------------------- |
| 200    | 请求成功                                                         |
| 400    | 参数校验失败或业务逻辑错误                                       |
| 401    | Token 缺失、无效或已过期（`Unauthorized` / `Invalid Token`） |
| 500    | 服务器内部错误                                                   |

---

### 接口列表

#### 1. 发送验证码

向指定邮箱发送 6 位数字验证码。

- 验证码有效期 10 分钟
- 60 秒内不可重复发送（防刷限流）

```
GET /acme/sendCode?email=user@example.com
```

**请求参数**

| 参数  | 类型   | 必填 | 说明         |
| ----- | ------ | ---- | ------------ |
| email | string | 是   | 用户邮箱地址 |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

#### 2. 检查用户名是否存在

```
GET /acme/exist?username=user@example.com
```

**请求参数**

| 参数     | 类型   | 必填 | 说明                 |
| -------- | ------ | ---- | -------------------- |
| username | string | 是   | 用户名（通常为邮箱） |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

> `data` 为 `true` 表示用户名已存在，`false` 表示不存在。

#### 3. 用户注册

```
POST /acme/register
```

**请求体**

```json
{
  "username": "user@example.com",
  "password": "your-password",
  "code": "123456"
}
```

| 参数     | 类型   | 必填 | 说明           |
| -------- | ------ | ---- | -------------- |
| username | string | 是   | 用户名（邮箱） |
| password | string | 是   | 密码           |
| code     | string | 是   | 邮箱验证码     |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": {
    "id": "1234567890",
    "username": "user@example.com",
    "privateKey": "user@example.com.key",
    "remark": "SUCCESS",
    "createTime": "2026-08-11T10:00:00+08:00",
    "updateTime": "2026-08-11T10:00:00+08:00"
  }
}
```

> 注册成功后，系统会在项目根目录生成 `<username>.key` RSA 2048 私钥文件，用于后续 ACME 证书签发。密码通过 bcrypt 哈希存储，不会对外暴露。

#### 4. 用户登录

```
POST /acme/login
```

**请求体**

```json
{
  "username": "user@example.com",
  "password": "your-password"
}
```

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

> `data` 为 UUID 格式 Token 字符串，有效期 24 小时。Token 对应的账户信息序列化后存储在内存缓存（ExpireMap）中。

#### 5. 获取 Web 管理界面

```
GET /acme/html
```

**成功响应**：返回 `static/index.html` 的 HTML 内容，Content-Type 为 `text/html; charset=utf-8`。

> 前端页面通过 Go `embed.FS` 嵌入到二进制文件中，部署时无需单独拷贝静态文件。

---

以下接口需要 Token 认证：

#### 6. 查询证书列表

分页查询当前用户的证书记录。

```
GET /acme/domain/page?current=1&size=20&domain=example.com
```

**请求参数**

| 参数    | 类型   | 必填 | 说明              |
| ------- | ------ | ---- | ----------------- |
| current | int    | 否   | 当前页码，默认 1  |
| size    | int    | 否   | 每页数量，默认 20 |
| domain  | string | 否   | 按域名模糊筛选    |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": {
    "total": 1,
    "records": [
      {
        "id": "1234567890",
        "username": "user@example.com",
        "encrypt": "/app/nginx/certs",
        "domain": "example.com",
        "remainDay": 30,
        "expireTime": "2027-01-15T08:00:00Z",
        "status": 1,
        "remark": "",
        "createTime": "2026-08-11T10:00:00+08:00",
        "updateTime": "2026-08-11T10:00:00+08:00"
      }
    ]
  }
}
```

#### 7. 创建证书

创建一条新的证书记录并异步签发证书。

```
POST /acme/domain/create
```

**请求体**

```json
{
  "domain": "example.com",
  "cipher": "cloudflare-api-token",
  "encrypt": "/app/nginx/certs",
  "remainDay": 30
}
```

| 参数          | 类型   | 必填 | 说明                                             |
|--------------| ------ | ---- | ------------------------------------------------ |
| domain       | string | 是   | 要签发证书的域名                                 |
| cipher       | string | 是   | Cloudflare API Token（用于 DNS-01 验证）         |
| encrypt      | string | 否   | 证书存储目录路径（留空则仅远程检测现有证书状态） |
| remainDay    | int    | 是   | 证书剩余天数阈值，低于此值时自动触发续期         |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": {
    "id": "1234567890",
    "username": "user@example.com",
    "encrypt": "/app/nginx/certs",
    "domain": "example.com",
    "remainDay": 30,
    "expireTime": null,
    "status": 1,
    "remark": "SUCCESS",
    "createTime": "2026-08-11T10:00:00+08:00",
    "updateTime": "2026-08-11T10:00:00+08:00"
  }
}
```

> 证书签发为 **异步操作**：接口立即返回，后台 goroutine 执行签发流程（最长 60 秒超时）。
>
> - 若 `encrypt` 路径非空：执行 Let's Encrypt 签发 → 保存证书（safeWriteFile 原子写入） → 读取过期时间 → 触发 Nginx 重载
> - 若 `encrypt` 路径为空：仅通过 TLS 远程检测现有证书的过期时间
> - 签发成功后，证书保存为 `<encrypt>/<domain>_bundle.pem`（证书链）和 `<encrypt>/<domain>.key`（私钥）

#### 8. 更新证书配置

更新指定域名的证书配置并异步重新签发。

```
POST /acme/domain/put
```

**请求体**

```json
{
  "domain": "example.com",
  "cipher": "new-cloudflare-api-token",
  "encrypt": "/app/nginx/certs",
  "remainDay": 30
}
```

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": {
    "id": "1234567890",
    "username": "user@example.com",
    "encrypt": "/app/nginx/certs",
    "domain": "example.com",
    "remainDay": 30,
    "expireTime": "2027-01-15T08:00:00Z",
    "status": 1,
    "remark": "",
    "createTime": "2026-08-11T10:00:00+08:00",
    "updateTime": "2026-08-11T10:05:00+08:00"
  }
}
```

> 更新操作会自动触发异步证书重新签发流程，流程与创建时相同。

#### 9. 手动刷新证书

手动触发指定域名的证书续期流程。

```
GET /acme/domain/refresh?domain=example.com
```

**请求参数**

| 参数   | 类型   | 必填 | 说明             |
| ------ | ------ | ---- | ---------------- |
| domain | string | 是   | 要刷新证书的域名 |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": {
    "id": "1234567890",
    "username": "user@example.com",
    "encrypt": "/app/nginx/certs",
    "domain": "example.com",
    "remainDay": 30,
    "expireTime": "2027-01-15T08:00:00Z",
    "status": 1,
    "remark": "",
    "createTime": "2026-08-11T10:00:00+08:00",
    "updateTime": "2026-08-11T10:05:00+08:00"
  }
}
```

> 仅对已存在且配置了 `encrypt` 路径的记录有效。异步执行重新签发 + Nginx 重载流程。

#### 10. 删除证书记录

删除指定域名的证书记录（不会删除服务器上的证书文件）。

```
DELETE /acme/domain/delete?domain=example.com
```

**请求参数**

| 参数   | 类型   | 必填 | 说明             |
| ------ | ------ | ---- | ---------------- |
| domain | string | 是   | 要删除的证书域名 |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

> **权限校验**：系统会验证当前用户是否为该记录的所有者，非本人记录将返回 `无权操作此记录` 错误。
>
> **注意**：此操作仅删除数据库中的证书记录，不会删除服务器文件系统上的证书文件（`.pem`）。如需清理证书文件，请手动处理。

#### 认证失败响应

当 Token 缺失或无效时，中间件直接拦截并返回统一格式错误响应：

```json
{
  "success": false,
  "timestamp": 1700000000000,
  "data": "Unauthorized"
}
```

## 数据模型

### AcmeAccount（用户账户）

| 字段        | 类型     | 说明                                |
| ----------- | -------- | ----------------------------------- |
| id          | int64    | 主键（Snowflake 生成）              |
| username    | string   | 用户名，唯一（通常为邮箱）          |
| password    | string   | 密码（bcrypt 哈希存储，不对外暴露） |
| private_key | string   | RSA 私钥文件路径                    |
| remark      | string   | 备注                                |
| create_time | datetime | 创建时间（自动填充）                |
| update_time | datetime | 更新时间（自动填充）                |

### AcmeEncrypt（证书记录）

| 字段        | 类型       | JSON 标签    | 说明                                           |
| ----------- | ---------- |------------| ---------------------------------------------- |
| id          | int64      | id,string  | 主键（Snowflake 生成，序列化为字符串）         |
| username    | string     | username   | 所属用户                                       |
| cipher      | string     | -          | Cloudflare API Token（不对外暴露 JSON 序列化） |
| encrypt     | string     | encrypt    | 证书存储目录路径                               |
| domain      | string     | domain     | 证书对应的域名                                 |
| remain_day  | int        | remainDay  | 续期阈值（剩余天数 ≤ 此值时触发续期）         |
| expire_time | *time.Time | expireTime | 证书过期时间（指针，可为空）                   |
| status      | uint8      | status     | 状态：1=处理中/成功，0=失败                    |
| remark      | string     | remark     | 备注或错误信息                                 |
| create_time | *time.Time | createTime | 创建时间（自动填充）                           |
| update_time | *time.Time | updateTime | 更新时间（自动填充）                           |

### BasePage（分页响应）

| 字段    | 类型  | 说明               |
| ------- | ----- | ------------------ |
| total   | int64 | 符合条件的总记录数 |
| records | array | 当前页的记录列表   |

### 请求结构体

**AcmeAccountRegister（注册请求）**

| 字段     | 类型   | JSON 标签 | 校验规则 |
| -------- | ------ | --------- | -------- |
| username | string | username  | required |
| password | string | password  | required |
| code     | string | code      | required |

**AcmeAccountLogin（登录请求）**

| 字段     | 类型   | JSON 标签 | 校验规则 |
| -------- | ------ | --------- | -------- |
| username | string | username  | required |
| password | string | password  | required |

**AcmeEncryptRequest（证书创建/更新请求）**

| 字段        | 类型   | JSON 标签   | 校验规则 |
|-----------| ------ |-----------| -------- |
| domain    | string | domain    | required |
| encrypt   | string | encrypt   | -        |
| cipher    | string | cipher    | required |
| remainDay | int    | remainDay | required |

**AcmeEncryptQuery（分页查询请求）**

| 字段    | 类型   | 表单标签 | 说明               |
| ------- | ------ | -------- | ------------------ |
| current | int    | current  | 当前页码，默认 1   |
| size    | int    | size     | 每页数量，默认 20  |
| domain  | string | domain   | 域名模糊搜索关键词 |

## 核心工作流程

### 用户注册流程

```
客户端请求 → GET /acme/sendCode → SMTP 发送 6 位验证码到邮箱（10 分钟有效）
    ↓
客户端提交 → POST /acme/register → bcrypt 校验密码强度
    ↓
验证码比对 → 校验通过后删除验证码（一次性使用）
    ↓
Snowflake 生成 ID → 生成 RSA 2048 私钥 → safeWriteFile 写入 PEM 文件
    ↓
账户信息持久化到 SQLite → 返回账户信息（不含密码）
```

### 用户登录流程

```
客户端请求 → POST /acme/login → bcrypt.CompareHashAndPassword 校验密码
    ↓
生成 UUID v4 格式 Token（xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx）
    ↓
账户信息 JSON 序列化 → 存入 ExpireMap 内存缓存（24h TTL）
    ↓
返回 Token → 客户端保存到 localStorage 用于后续认证
```

### Token 鉴权流程

```
请求到达 → Auth() 中间件
    ↓
从 Authorization Header 或 access_token Query 参数提取 Token
    ↓
Bearer 格式校验 → ExpireMap 查询 Token 对应的账户信息
    ↓
JSON 反序列化为 AcmeAccount → 将 username 注入 Request Context
    ↓
后续 Service 层通过 ctx.Value("username") 获取当前操作用户
```

### 证书签发流程

```
创建/更新请求 → 写入证书记录（status=1, remark="SUCCESS"）
    ↓
异步 goroutine（60s 超时 context）
    ↓
加载用户 RSA 私钥 → 配置 lego ACME 客户端
    ↓
Cloudflare DNS-01 验证（cfConfig.AuthToken = cipher）
    ↓
首次注册（Register）或解析已有账户（ResolveAccountByKey）
    ↓
申请证书（Certificate.Obtain）→ Bundle=true 获取证书链
    ↓
安全写入证书文件（safeWriteFile：写临时文件 → 原子重命名）
    ↓
  encrypt 非空？                encrypt 为空？
  → 读取本地 PEM 过期时间      → TLS 远程检测 443 端口
  → 更新 expire_time            → 更新 expire_time
  → Docker Exec: nginx -t       → 结束
  → SIGHUP 重载 Nginx
```

### 证书自动续期判断

系统在创建、更新、刷新证书时均会检测证书有效期：

- **本地检测**：`GetLocalCertExpire()` — 读取 PEM 文件中的 `NotAfter` 字段
- **远程检测**：`GetRemoteCertExpire()` — TLS 直连域名 443 端口获取证书
- **续期条件**：`remainDay <= remainDay` 时视为需要续期

## Docker 集成说明

本服务通过 Docker API（`docker.sock`）与 Nginx 容器联动，实现证书更新后自动重载配置。

### 必要的 Docker 挂载

| 挂载路径                 | 说明                                           |
| ------------------------ | ---------------------------------------------- |
| `/var/run/docker.sock` | Docker Socket（必需，用于 Exec 和 Kill 操作）  |
| `/usr/bin/docker`      | Docker CLI（可选，用于容器内调用 docker 命令） |
| `/home/nginx/certs`    | 证书存储目录（需与 Nginx 容器共享同一目录）    |

### Nginx 容器要求

- 容器名需为 `nginx`（可在 `docker_repository.go` 的 `ReloadConfig` 函数中修改）
- 支持 `nginx -t` 配置语法检测
- 支持 `SIGHUP` 信号触发 `nginx -s reload`
- 需挂载与本服务相同的证书目录

## 测试

项目包含以下测试：

```bash
# 进入测试目录
cd test

# 运行 Cloudflare DNS 证书签发完整测试（需 CLOUDFLARE_ACCESS_TOKEN 环境变量）
CLOUDFLARE_ACCESS_TOKEN=your-token go test -v -run TestLetsEncrypt

# 远程检测证书过期时间
go test -v -run TestCheckRemoteCertExpire

# 本地检测证书过期时间
go test -v -run TestCheckLocalCertExpire
```

## CI/CD

项目使用 GitHub Actions 实现自动化构建和部署。

### 触发条件

- 向 `release` 分支推送代码
- 向 `release` 分支提交 Pull Request

### 构建流程

1. 拉取代码（actions/checkout@v4）
2. 初始化 Docker Buildx（docker/setup-buildx-action@v3）
3. 登录 Docker Hub（docker/login-action@v3）
4. 构建并推送镜像（docker/build-push-action@v6），镜像标签：`kouleen/lets-encrypt:latest`
5. 调用自定义 Webhook API（`https://www.kouleen.cn/api/docker/webhook/put`）完成远程部署
   - 自动传递环境变量、端口映射、卷挂载等容器配置
   - 支持远程移除旧容器并启动新容器
   - 包含失败阶段检测与错误信息上报

### GitHub Secrets 配置

在 GitHub 仓库 **Settings → Secrets and variables → Actions** 中配置以下 Secrets：

| Secret                 | 说明              |
| ---------------------- | ----------------- |
| `DOCKER_USERNAME`    | Docker Hub 用户名 |
| `DOCKER_PASSWORD`    | Docker Hub 密码   |
| `DOCKER_SEND_EMAIL`  | 发件邮箱          |
| `DOCKER_SEND_PWD`    | 邮箱密码          |
| `DOCKER_SMTP_SERVER` | SMTP 服务器地址   |
| `DOCKER_SMTP_PORT`   | SMTP 端口         |

## 配置说明

### ACME 环境切换

代码默认使用 Let's Encrypt **生产环境**，可通过环境变量切换到测试环境：

```go
// internal/modle/acme_user.go
cfg.CADirURL = lego.LEDirectoryProduction  // 生产环境（默认）
// cfg.CADirURL = lego.LEDirectoryStaging   // 测试环境（调试用）
```

或设置 `ENV=dev` 环境变量自动切换：

```bash
ENV=dev go run main.go
```

> ⚠️ 注意：Let's Encrypt 生产环境有速率限制，请先在 Staging 环境充分调试。

### 端口修改

服务默认监听 `8099` 端口，可在 `main.go` 中修改：

```go
srv := &http.Server{
    Addr: ":8099",  // 修改此处
}
```

### Snowflake 节点 ID

`snowflake.NewNode(1)` 在多处硬编码为节点 1，分布式部署时需修改为不同节点 ID。

### SQLite 数据库配置

数据库使用文件 `lets-encrypt.db`，自动创建于项目根目录。WAL 模式提升并发性能。连接池配置：

- 最大打开连接：1
- 最大空闲连接：1
- 连接永不过期

## 安全说明

1. **RSA 私钥保管**：注册后在项目根目录生成的 `<username>.key` 文件为 ACME 账户私钥，丢失后将无法续期已有证书
2. **Cloudflare Token 权限**：Token 需具备对应域名的 DNS 编辑权限（推荐使用 API Token，范围限定到特定 Zone）
3. **SMTP 可用性**：SMTP 服务不可用时用户无法完成注册
4. **Docker Socket 安全**：挂载 `docker.sock` 存在容器逃逸风险，生产环境建议使用 rootless Docker、Podman 或单独的 sidecar 容器
5. **证书路径一致性**：`encrypt` 路径需确保容器内的 Nginx 可访问（挂载同一主机目录）
6. **数据库备份**：`lets-encrypt.db` 包含用户信息和密钥路径映射，需定期备份
7. **Token 安全**：Token 存储在进程内存中（ExpireMap），服务重启后所有 Token 失效，用户需重新登录
8. **删除操作**：删除证书记录仅影响数据库，不会清理磁盘上的证书文件
9. **密码安全**：密码通过 bcrypt 哈希存储，`cipher` 字段（Cloudflare API Token）在 JSON 序列化时使用 `-` 标签排除，不对外暴露

## 注意事项

1. **证书有效期**：Let's Encrypt 证书有效期为 90 天，系统默认在剩余 30 天时触发续期
2. **速率限制**：Let's Encrypt 生产环境有严格的速率限制，建议先在 Staging 环境充分调试
3. **异步签发**：证书签发为后台 goroutine 执行，最长 60 秒超时。前端应通过轮询或刷新查看最终状态
4. **安全写文件**：证书写入采用临时文件 + 原子重命名方式（safeWriteFile），避免写入中断导致文件损坏
5. **响应格式**：所有接口响应已统一包装为 `{success, timestamp, data}` 格式，前端需适配此格式
6. **移动端兼容性**：移动端采用卡片式证书列表和底部导航，确保在小屏幕设备上的良好体验

## 许可证

本项目作者：[Kouleen](mailto:Kouleen.china@gmail.com)
