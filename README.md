# 🔐 CF SSL Manager — CloudFlare SSL 证书管理工具

> 基于 **Wails + Go + Vue3** 构建的跨平台桌面应用，助你轻松管理 CloudFlare 域名的 SSL/TLS 证书。

---

## 📖 项目简介

**CF SSL Manager** 是一款管理 CloudFlare 域名的 Windows 桌面工具，通过 CloudFlare API 管理你的域名，并使用 **ACME 协议**（Let's Encrypt）自动申请免费 SSL 证书。

所有敏感数据（API Key、Token）均经过 **AES 加密** 后存入本地 SQLite 数据库，安全可靠。

---

## ✨ 核心功能

### 🔑 账户管理
- 添加 / 编辑 / 删除 CloudFlare 账户
- 支持 **API Key** + 邮箱 或 **API Token** 两种认证方式
- 一键验证账户凭证是否有效
- API 凭证本地加密存储，拒绝明文泄露

### 🌐 域名区域管理
- 自动拉取账户下所有 CloudFlare 域名区域
- 查看域名的套餐类型、名称服务器（NS）等信息
- 按账户筛选管理

### 📜 SSL 证书申请
- 通过 ACME 协议自动申请 Let's Encrypt 免费证书
- 支持 **SAN（多域名）证书**，一次申请覆盖多个域名
- 支持多种密钥算法：**ECDSA P-256 / P-384**、**RSA 2048 / 4096**
- 自定义证书存储路径
- 实时日志输出，申请过程透明可见

### 📋 证书列表 & 筛选
- 查看所有已申请的证书，按到期时间排序
- 多维度筛选：域名搜索、CA 品牌、账户、证书状态
- 自动计算剩余天数，🚦 **绿 / 黄 / 红** 三色标识证书状态：
  - 🟢 **有效** — 剩余 > 30 天
  - 🟡 **即将到期** — 剩余 ≤ 30 天
  - 🔴 **已过期**

### 🔍 证书详情
- 查看证书完整信息：主题、颁发者、序列号、指纹等
- 预览 PEM 格式的 **证书、私钥、证书链、完整链**
- 支持一键复制 / 导出证书内容

### 🛠️ 实用工具
- **远程证书检查** — 输入域名 + 端口，查看服务器当前部署的证书信息
- **PEM 导入解析** — 粘贴 PEM 文本即可解析证书详情
- **打开证书目录** — 一键打开证书文件所在文件夹

### ⚙️ 系统设置
- 自定义证书存储目录
- 到期提醒天数配置
- 自动续期开关
- 代理设置（HTTP / SOCKS5）
- 调试面板：数据库状态、CloudFlare API 连通性检测

---

## 🏗️ 技术栈

| 层级 | 技术 |
|------|------|
| 🖥️ 桌面框架 | [Wails v2](https://wails.io/) |
| ⚙️ 后端 | Go (GORM + SQLite + ACME) |
| 🎨 前端 | Vue 3 + Vite + Element Plus |
| 🔐 加密 | AES-GCM |
| 🗄️ 数据库 | SQLite (本地) |

---

## 🚀 快速开始

### 环境要求

- **Go** 1.23+
- **Node.js** 18+
- **Wails CLI** v2
- **GCC** 编译器（TDM-GCC 或 MinGW）

### 开发调试

```bash
# 1. 进入项目目录
cd cf-ssl-manager

# 2. 安装前端依赖
cd frontend && npm install && cd ..

# 3. 启动开发模式（热重载）
wails dev
```

---

## 📦 构建 Windows exe 教程

### 第一步：环境准备

#### 1.1 安装 Go (1.23+)

下载地址：[https://go.dev/dl/](https://go.dev/dl/)

```bash
go version
# ✅ 应输出: go version go1.23+ windows/amd64
```

#### 1.2 安装 Node.js (18+)

下载地址：[https://nodejs.org/](https://nodejs.org/)

```bash
node -v
npm -v
```

#### 1.3 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

```bash
wails version
# ✅ 应输出: v2.12.0 或更高
```

#### 1.4 环境检查 & 依赖安装

```bash
wails doctor
```

确保所有检查项通过，特别关注：
- ✅ **WebView2 Runtime**（Win10/11 已预装）
- ✅ **GCC 编译器**（用于 SQLite CGO 编译）

> 💡 **缺少 GCC？** 安装 [TDM-GCC](https://jmeubank.github.io/tdm-gcc/)，安装时勾选 "Add to PATH"

---

### 第二步：构建 exe

```bash
# 1. 进入项目目录
cd cf-ssl-manager

# 2. 安装前端依赖
cd frontend && npm install && cd ..

# 3. 构建！
wails build
```

构建成功后输出类似：
```
Built 'cf-ssl-manager\build\bin\cf-ssl-manager.exe' in 13s.
```

**产物位置：** `build/bin/cf-ssl-manager.exe`（约 17MB，单文件绿色版）

---

### 第三步：构建选项（可选）

```bash
# UPX 压缩构建（进一步减小体积）
wails build -upx
# 需先安装 UPX: https://github.com/upx/upx/releases

# 跳过前端重新构建（仅改后端时使用）
wails build -skipfrontend

# 自定义输出文件名
wails build -o "CF-SSL-Manager.exe"

# 生产构建（嵌入版本号）
wails build -ldflags "-X main.version=1.0.0"
```

---

### 常见问题 Q&A

<details>
<summary><b>❓ 构建时报 "gcc not found"</b></summary>

安装 [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) 并确保加入系统 PATH，然后运行 `wails doctor` 检查。
</details>

<details>
<summary><b>❓ 构建时报 "WebView2 not found"</b></summary>

下载安装 [WebView2 Runtime](https://developer.microsoft.com/en-us/microsoft-edge/webview2/)
</details>

<details>
<summary><b>❓ exe 运行后白屏</b></summary>

检查杀毒软件是否拦截，将 exe 加入白名单。
</details>

<details>
<summary><b>❓ 数据存储在哪里？</b></summary>

- 🗄️ 数据库：`%APPDATA%/cf-ssl-manager/certs.db`
- 📁 证书文件：`C:/Users/你的用户名/cf-ssl-certs/`（可在设置中自定义）
文件	作用
certs.db	主数据库文件，存放所有表和实际数据
certs.db-wal	WAL 日志文件，写入操作先记到这里，再批量合并回主文件
certs.db-shm	共享内存索引，用于多连接并发访问 WAL 文件的快速查找
</details>

---

## 📦 分发说明

生成的 `cf-ssl-manager.exe` 是**独立单文件**，无需安装：

- ✅ 复制到任意 Windows 10/11 电脑直接运行
- ✅ 放入 U 盘随身携带
- ✅ 可用 NSIS / Inno Setup 打包成安装程序

> ⚠️ 目标机器需预装 **WebView2 Runtime**（Win10/11 通常已内置）

---

## 📂 项目结构

```
cf-ssl-manager/
├── app/                    # Go 后端逻辑
│   ├── acme.go             # ACME 证书申请
│   ├── cert.go             # 证书解析 & 远程检查
│   ├── cf.go               # CloudFlare API 对接
│   ├── config.go           # 配置管理
│   ├── crypto.go           # AES 加解密
│   ├── db.go               # SQLite 数据库
│   └── models.go           # 数据模型
├── app.go                  # Wails 主入口 & API 绑定
├── frontend/               # Vue3 前端
│   └── src/
│       ├── views/          # 页面组件
│       │   ├── ApplyView.vue       # 证书申请
│       │   ├── CertListView.vue    # 证书列表
│       │   ├── CertDetailView.vue  # 证书详情
│       │   ├── ConfigView.vue      # 账户 & 域名管理
│       │   └── SettingsView.vue    # 系统设置
│       ├── components/     # 通用组件
│       └── stores/         # Pinia 状态管理
└── build/                  # 构建配置 & 产物
    └── bin/                # exe 输出目录
```

---

## 📄 License

MIT

---

<p align="center">Made with ❤️ using Wails + Go + Vue3</p>
