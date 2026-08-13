# Let's Encrypt 证书管理平台

基于 Go 语言构建的 Let's Encrypt SSL 证书自动化管理平台，通过 Cloudflare DNS 验证方式自动签发和续期证书，支持与 Nginx 容器的联动。内置响应式 Web 管理界面（毛玻璃风格 UI），支持桌面端和移动端，开箱即用。

## 功能特性

- **用户系统**：邮箱验证码注册、bcrypt 密码加密、Token 鉴权登录、24 小时会话保持
- **证书管理**：创建、查询、编辑、刷新、删除 Let's Encrypt 证书记录
- **自动续期**：定时任务（每日 00:00）自动检测证书有效期，低于续期阈值时自动触发续期
- **邮件通知**：定时任务（每日 00:00）自动检测即将到期的证书，发送邮件提醒
- **DNS 验证**：通过 Cloudflare DNS API 完成 ACME DNS-01 验证
- **Nginx 联动**：证书更新后自动通过 Docker API 重载 Nginx 配置（SIGHUP 信号）
- **SQLite 持久化**：使用 SQLite 数据库存储账户和证书信息（WAL 模式）
- **响应式 Web 界面**：内置 SPA，毛玻璃风格 UI，支持桌面端 / 平板 / 移动端自适应
- **移动端优化**：汉堡菜单侧边栏、底部导航栏、卡片式证书列表
- **统一响应格式**：所有 API 响应均采用 `{success, timestamp, data}` 标准包装
- **权限隔离**：用户只能管理自己名下的证书记录
- **安全写文件**：证书写入采用临时文件 + 原子重命名方式，防止写入中断导致文件损坏
- **到期预警**：即将过期证书行底色标黄，已过期证书行底色标红
- **开关控件**：自动续期与邮件通知采用滑动开关，操作即时生效
- **证书下载**：一键打包下载证书链和私钥文件（ZIP 格式）

## 技术栈

| 类别        | 技术                         | 说明                   |
| ----------- | ---------------------------- | ---------------------- |
| 语言        | Go 1.25+                     | 主开发语言             |
| Web 框架    | Gin v1.12                    | HTTP API 服务框架      |
| ORM         | GORM v1.31                   | 数据库操作             |
| 数据库      | SQLite (glebarez/sqlite)     | 轻量级嵌入式数据库     |
| ACME 客户端 | go-acme/lego/v4 v4.35        | Let's Encrypt 证书签发 |
| DNS 验证    | Cloudflare DNS               | ACME DNS-01 验证       |
| 定时任务    | robfig/cron/v3               | 自动续期与通知调度     |
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
├── kouleen.china@gmail.com.key          # 用户 RSA 私钥文件（注册时自动生成）
├── static/
│   ├── embed.go                         # embed.FS 嵌入静态文件
│   ├── index.html                       # 前端 SPA 页面（毛玻璃 UI）
│   ├── notice.html                      # 邮件通知模板
│   └── code.html                        # 验证码邮件模板
├── internal/
│   ├── api/acme_api.go                  # HTTP Handler 层
│   ├── routes/acme.go                   # 路由注册
│   ├── service/acme_service.go          # 业务逻辑层
│   ├── middleware/auth.go                # Token 鉴权中间件
│   ├── modle/
│   │   ├── acme_encrypt.go              # 证书数据模型
│   │   ├── acme_account.go              # 账户数据模型
│   │   ├── acme_user.go                 # ACME 用户操作（签发/刷新）
│   │   └── base_page.go                 # 分页模型
│   ├── repository/
│   │   ├── sqlite_repository.go         # SQLite 数据访问层
│   │   ├── docker_repository.go         # Docker 操作（Nginx 重载）
│   │   └── store_repository.go          # 缓存操作（Token/验证码）
│   ├── task/deily_task.go               # 定时任务（自动续期 + 邮件通知）
│   └── validator/validator.go          # 参数校验器
├── pkg/util/
│   ├── email.go                         # 邮件发送工具
│   ├── acme_verify.go                   # 证书过期时间检测
│   ├── acme_key.go                      # RSA 密钥管理
│   └── store.go                         # 缓存工具封装
└── test/                                # 集成测试
```

## 快速开始

### 1. 环境要求

- Go 1.25+（本地编译）
- Docker & Docker Compose（容器部署）
- Cloudflare 账号（DNS 验证）
- SMTP 邮件服务（发送验证码和到期提醒）

### 2. 本地编译运行

```bash
# 克隆项目
git clone <repo-url>
cd lets-encrypt

# 设置邮件服务环境变量（必填）
export SMTP_SERVER=smtp.example.com
export SMTP_PORT=465
export SEND_EMAIL=your@email.com
export SEND_PWD=your-password

# 编译并运行
go build -o lets-encrypt .
./lets-encrypt
```

服务启动后访问 `http://localhost:8099/acme/html` 进入管理界面。

### 3. Docker Compose 部署

```bash
# 克隆项目
git clone <repo-url>
cd lets-encrypt

# 创建环境变量文件
cat > .env << 'EOF'
SMTP_SERVER=smtp.example.com
SMTP_PORT=465
SEND_EMAIL=your@email.com
SEND_PWD=your-password
EOF

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 4. 首次使用

1. 浏览器访问 `http://<server-ip>:8099/acme/html`
2. 点击注册，输入邮箱获取验证码完成注册
3. 登录后进入"证书管理"页面
4. 点击"创建证书"，填写域名和 Cloudflare API Token
5. 系统自动签发证书并存储到指定路径

## 配置说明

### 环境变量

| 变量名      | 必填 | 说明                | 示例                  |
| ----------- | ---- | ------------------- | --------------------- |
| SMTP_SERVER | 是   | SMTP 邮件服务器地址 | `smtp.qq.com`       |
| SMTP_PORT   | 是   | SMTP 邮件服务器端口 | `465`               |
| SEND_EMAIL  | 是   | 发件邮箱地址        | `admin@example.com` |
| SEND_PWD    | 是   | 发件邮箱密码/授权码 | `auth-code`         |

### Docker Volume 说明

| 宿主机路径               | 容器路径                 | 说明                              |
| ------------------------ | ------------------------ | --------------------------------- |
| `/usr/bin/docker`      | `/usr/bin/docker`      | Docker 二进制（用于容器管理）     |
| `/var/run/docker.sock` | `/var/run/docker.sock` | Docker Socket（用于 Nginx 联动）  |
| `/home/nginx/certs`    | `/app/nginx/certs`     | 证书存储目录（与 Nginx 容器共享） |

### Cloudflare API Token 配置

1. 登录 Cloudflare 控制台
2. 进入 "My Profile" → "API Tokens"
3. 创建 API Token，权限选择：
   - `Zone:DNS:Edit`（DNS 记录编辑权限）
   - `Zone:Read`（区域读取权限）
4. 将 Token 填入证书创建/编辑的"加密 API"字段

## UI 界面说明

### 页面结构

| 页面      | 功能说明                                                                    |
| --------- | --------------------------------------------------------------------------- |
| 登录/注册 | 邮箱验证码注册、密码登录、Token 自动持久化（localStorage）                  |
| 概览面板  | 证书总数、有效/即将到期/已过期统计、最近 5 条证书预览                       |
| 证书管理  | 分页列表、域名搜索、创建/编辑/刷新/删除证书记录、自动续期开关、邮件通知开关 |

### 表格列说明

桌面端表格展示完整信息，共 9 列：

| 列名     | 说明                                             |
| -------- | ------------------------------------------------ |
| 域名     | 证书绑定的域名                                   |
| 状态     | 有效 / 异常（带状态徽章）                        |
| 过期时间 | 证书过期日期（即将过期显示黄色，已过期显示红色） |
| 剩余天数 | 距过期的剩余天数（安全绿/警告黄/危险红）         |
| 续期阈值 | 触发自动续期的剩余天数阈值                       |
| 自动续期 | 滑动开关，开启后每日定时任务自动检测续期         |
| 邮件通知 | 滑动开关，开启后即将到期时发送邮件提醒           |
| 证书路径 | 证书文件存储目录                                 |
| 操作     | 编辑、刷新、下载、删除                           |

### 移动端适配

- ≤1024px：侧边栏收缩为 72px 图标模式
- ≤768px：侧边栏隐藏，底部导航栏激活，证书列表转为卡片视图
- 卡片信息完整展示：过期时间、剩余天数、续期阈值、自动续期开关、邮件通知开关、证书路径

### UI 设计特点

- **深色毛玻璃风格**：`backdrop-filter: blur(20px)` 玻璃质感
- **动态背景**：多层径向渐变 + 缓动动画（20 秒循环）
- **渐变主色调**：紫蓝色系 `#667eea → #764ba2`
- **Toast 通知**：右上角滑入动画，支持 success/error/info/warning
- **弹窗动画**：弹性缩放 + 遮罩淡入，移动端转为底部抽屉样式
- **状态徽章**：有效/异常带发光动画
- **剩余天数标签**：安全绿/警告黄/危险红三级预警
- **行底色预警**：即将过期证书行底色标黄，已过期证书行底色标红

### 自动续期与邮件通知

每个证书记录支持两个独立配置项：

| 配置项   | 字段       | 默认值 | 说明                                                                |
| -------- | ---------- | ------ | ------------------------------------------------------------------- |
| 自动续期 | `auto`   | `0`  | 开启后，每日 00:00 定时任务检测到证书剩余天数 ≤ 续期阈值时自动续期 |
| 邮件通知 | `notice` | `1`  | 开启后，每日 00:00 定时任务检测到证书即将到期时发送邮件提醒         |

通过表格中的滑动开关即时切换，无需进入编辑页面。

> 注意：自动续期需要证书已配置 `encrypt` 路径（证书存储目录），否则无法执行签发流程。

## API 文档

### 基础信息

- **Base URL**: `http://localhost:8099`
- **数据格式**: JSON
- **认证方式**: Bearer Token（登录后获取，有效期 24 小时）
- **传参方式**:
  - Header: `Authorization: Bearer <token>`
  - Query: `access_token=<token>`

### 统一响应格式

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": {}
}
```

---

以下接口无需认证：

#### 1. 发送验证码

```
GET /acme/sendCode?email=user@example.com
```

**请求参数**

| 参数  | 类型   | 必填 | 说明     |
| ----- | ------ | ---- | -------- |
| email | string | 是   | 邮箱地址 |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

> 验证码 60 秒内不可重复发送，有效期 10 分钟。

#### 2. 用户注册

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

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

#### 3. 用户登录

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
  "data": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

> `data` 为 UUID 格式 Token 字符串，有效期 24 小时。

#### 4. 检测用户名是否存在

```
GET /acme/exist?username=user@example.com
```

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

#### 5. 获取 Web 管理界面

```
GET /acme/html
```

返回 `static/index.html` 的 HTML 内容。

---

以下接口需要 Token 认证：

#### 6. 查询证书列表

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
        "auto": 1,
        "notice": 1,
        "remark": "",
        "createTime": "2026-08-11T10:00:00+08:00",
        "updateTime": "2026-08-11T10:00:00+08:00"
      }
    ]
  }
}
```

> 返回数据中 `auto`（自动续期开关，0=关闭，1=开启）和 `notice`（邮件通知开关，0=关闭，1=开启）字段。

#### 7. 创建证书

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

| 参数      | 类型   | 必填 | 说明                                             |
| --------- | ------ | ---- | ------------------------------------------------ |
| domain    | string | 是   | 要签发证书的域名                                 |
| cipher    | string | 是   | Cloudflare API Token（用于 DNS-01 验证）         |
| encrypt   | string | 否   | 证书存储目录路径（留空则仅远程检测现有证书状态） |
| remainDay | int    | 是   | 证书剩余天数阈值，低于此值时自动触发续期         |

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
    "auto": 0,
    "notice": 1,
    "remark": "SUCCESS",
    "createTime": "2026-08-11T10:00:00+08:00",
    "updateTime": "2026-08-11T10:00:00+08:00"
  }
}
```

> 证书签发为 **异步操作**：接口立即返回，后台 goroutine 执行签发流程（最长 60 秒超时）。
>
> - 若 `encrypt` 路径非空：执行 Let's Encrypt 签发 → 保存证书 → 读取过期时间 → 触发 Nginx 重载
> - 若 `encrypt` 路径为空：仅通过 TLS 远程检测现有证书的过期时间
> - 签发成功后，证书保存为 `<encrypt>/<domain>_bundle.pem`（证书链）和 `<encrypt>/<domain>.key`（私钥）
> - 新建证书记录默认 `auto=0`（自动续期关闭）、`notice=1`（邮件通知开启）

#### 8. 更新证书配置

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
    "auto": 1,
    "notice": 1,
    "remark": "",
    "createTime": "2026-08-11T10:00:00+08:00",
    "updateTime": "2026-08-11T10:05:00+08:00"
  }
}
```

> 更新操作会自动触发异步证书重新签发流程。

#### 9. 手动刷新证书

```
GET /acme/domain/refresh?domain=example.com
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
    "auto": 1,
    "notice": 1,
    "remark": "",
    "createTime": "2026-08-11T10:00:00+08:00",
    "updateTime": "2026-08-11T10:05:00+08:00"
  }
}
```

> 仅对已存在且配置了 `encrypt` 路径的记录有效。

#### 10. 更新自动续期配置

```
PUT /acme/domain/updateAuto
```

**请求体**

```json
{
  "domain": "example.com",
  "auto": 1
}
```

| 参数   | 类型   | 必填 | 说明                         |
| ------ | ------ | ---- | ---------------------------- |
| domain | string | 是   | 证书域名                     |
| auto   | uint8  | 是   | 自动续期开关：0=关闭，1=开启 |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

> 开启时若证书路径为空会返回错误。开启后，每日 00:00 定时任务检测到证书剩余天数 ≤ 续期阈值时自动续期。

#### 11. 更新邮件通知配置

```
PUT /acme/domain/updateNotice
```

**请求体**

```json
{
  "domain": "example.com",
  "notice": 1
}
```

| 参数   | 类型   | 必填 | 说明                         |
| ------ | ------ | ---- | ---------------------------- |
| domain | string | 是   | 证书域名                     |
| notice | uint8  | 是   | 邮件通知开关：0=关闭，1=开启 |

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

#### 12. 删除证书

```
DELETE /acme/domain/delete?domain=example.com
```

**成功响应**

```json
{
  "success": true,
  "timestamp": 1700000000000,
  "data": true
}
```

#### 13. 下载证书

```
GET /acme/domain/download?domain=example.com
```

**成功响应**：返回 ZIP 文件，包含：

- `<domain>_bundle.pem`（证书链）
- `<domain>.key`（私钥）

## 数据模型

### AcmeEncrypt（证书记录）

| 字段       | 类型       | 说明                                          |
| ---------- | ---------- | --------------------------------------------- |
| id         | int64      | 雪花算法 ID                                   |
| username   | string     | 所属用户邮箱                                  |
| domain     | string     | 证书域名（唯一）                              |
| cipher     | string     | Cloudflare API Token（加密存储，JSON 不返回） |
| encrypt    | string     | 证书存储目录                                  |
| remainDay  | int        | 续期阈值（天）                                |
| expireTime | *time.Time | 过期时间                                      |
| status     | *uint8     | 状态（1=有效，0=异常）                        |
| auto       | *uint8     | 自动续期（0=关闭，1=开启）                    |
| notice     | *uint8     | 邮件通知（0=关闭，1=开启）                    |
| remark     | string     | 备注/错误信息                                 |
| createTime | *time.Time | 创建时间                                      |
| updateTime | *time.Time | 更新时间                                      |

### AcmeAccount（用户账户）

| 字段       | 类型       | 说明             |
| ---------- | ---------- | ---------------- |
| id         | int64      | 雪花算法 ID      |
| username   | string     | 用户邮箱（唯一） |
| password   | string     | bcrypt 加密密码  |
| privateKey | string     | RSA 私钥文件路径 |
| remark     | string     | 备注             |
| createTime | *time.Time | 创建时间         |
| updateTime | *time.Time | 更新时间         |

### 请求结构体

**AcmeEncryptRequest（创建/更新请求）**

| 字段      | 类型   | 必填 | 说明                 |
| --------- | ------ | ---- | -------------------- |
| domain    | string | 是   | 证书域名             |
| cipher    | string | 是   | Cloudflare API Token |
| encrypt   | string | 否   | 证书存储目录         |
| remainDay | int    | 是   | 续期阈值（天）       |

**AcmeEncryptAuto（自动续期配置请求）**

| 字段   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| domain | string | 是   | 证书域名 |
| auto   | uint8  | 是   | 0/1      |

**AcmeEncryptNotice（邮件通知配置请求）**

| 字段   | 类型   | 必填 | 说明     |
| ------ | ------ | ---- | -------- |
| domain | string | 是   | 证书域名 |
| notice | uint8  | 是   | 0/1      |

## 定时任务流程

```
Cron 定时任务（每日 00:00:00，本地时区）
    ↓
┌─ autoRefreshReload() 自动续期 ──────────────────────┐
│ 查询 auto=1 且 status=1 的证书记录                    │
│     ↓                                               │
│ 遍历每条记录：                                         │
│   → 跳过无 encrypt 路径的记录                          │
│   → 计算剩余天数 = expireTime - now                    │
│   → 剩余天数 > remainDay？ → 跳过                      │
│   → 调用 RefreshAcmeEncrypt() 触发异步续期             │
└──────────────────────────────────────────────────────┘

┌─ noticeEmail() 邮件通知 ────────────────────────────┐
│ 查询 notice=1 且 status=1 的证书记录                  │
│     ↓                                              │
│ 遍历每条记录：                                        │
│   → 剩余天数 ≤ 0（已过期）？ → 跳过                    │
│   → 剩余天数 > remainDay？ → 跳过                     │
│   → 按用户名分组，构建通知列表                          │
│     ↓                                               │
│ 按用户发送邮件（HTML 模板）                            │
└─────────────────────────────────────────────────────┘
```

## 证书签发流程

```
用户请求 → CreateAcmeEncrypt / UpdateAcmeEncrypt
    ↓
写入数据库（状态=有效）→ 异步 goroutine 启动
    ↓
┌─ encrypt 路径非空 ──────────────────────────────────┐
│ 1. 初始化 ACME 客户端（lego/v4）                      │
│ 2. Cloudflare DNS-01 验证                             │
│ 3. Let's Encrypt 签发证书                             │
│ 4. 原子写入文件（临时文件 + rename）                   │
│    → <encrypt>/<domain>_bundle.pem                    │
│    → <encrypt>/<domain>.key                           │
│ 5. 读取证书过期时间                                   │
│ 6. Docker API → Nginx SIGHUP 重载配置                 │
│ 7. 更新数据库（expireTime, status, remark）           │
└──────────────────────────────────────────────────────┘

┌─ encrypt 路径为空 ──────────────────────────────────┐
│ 1. TLS 远程检测域名证书过期时间                       │
│ 2. 更新数据库（expireTime）                           │
└──────────────────────────────────────────────────────┘
```

## 数据库说明

SQLite 数据库文件 `lets-encrypt.db`，使用 WAL 模式以支持并发读写。

### 表结构

**acme_account（用户账户表）**

| 字段        | 类型     | 约束        |
| ----------- | -------- | ----------- |
| id          | INTEGER  | PRIMARY KEY |
| username    | TEXT     | UNIQUE      |
| password    | TEXT     | NOT NULL    |
| private_key | TEXT     |             |
| remark      | TEXT     |             |
| create_time | DATETIME |             |
| update_time | DATETIME |             |

**acme_encrypt（证书记录表）**

| 字段        | 类型     | 约束                      |
| ----------- | -------- | ------------------------- |
| id          | INTEGER  | PRIMARY KEY               |
| username    | TEXT     | NOT NULL, INDEX           |
| domain      | TEXT     | NOT NULL, UNIQUE          |
| cipher      | TEXT     | NOT NULL                  |
| encrypt     | TEXT     |                           |
| remain_day  | INTEGER  | NOT NULL                  |
| expire_time | DATETIME |                           |
| status      | INTEGER  | NOT NULL, DEFAULT 1       |
| auto        | INTEGER  | NOT NULL, DEFAULT 0       |
| notice      | INTEGER  | NOT NULL, DEFAULT 1       |
| remark      | TEXT     | DEFAULT ''                |
| create_time | DATETIME | DEFAULT CURRENT_TIMESTAMP |
| update_time | DATETIME | DEFAULT CURRENT_TIMESTAMP |

## 安全说明

- **密码加密**：用户密码使用 bcrypt 加密存储
- **Token 机制**：随机 UUID Token 存储在服务端缓存，24 小时过期
- **权限隔离**：每个用户只能访问自己名下的证书记录
- **安全写文件**：证书写入采用临时文件 + 原子重命名
- **参数校验**：所有请求参数经过 validator 校验
- **HTTPS**：建议在生产环境前置 Nginx 反向代理并启用 HTTPS

## License

MIT License
