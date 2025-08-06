# create-worker-app

🚀 快速创建 Cloudflare Workers 应用的脚手架工具，基于超快的 Hono.js 框架。

[English](./README.md)

## ✨ 特性

- ⚡️ **超快性能** - 基于 [Hono.js](https://hono.dev/)，专为 Edge 环境优化
- 📝 **TypeScript 优先** - 完整的类型支持和智能提示
- 📚 **API 文档自动生成** - 集成 OpenAPI/Swagger，访问 `/docs` 查看
- 🛠️ **智能路由生成器** - 交互式 CLI 快速生成 CRUD、Auth、Webhook 等模板
- 🎯 **生产就绪** - 内置错误处理、CORS、日志等中间件
- 🚀 **一键部署** - 预配置 Wrangler，支持多环境部署
- 🤖 **Claude Code 集成** - AI 驱动开发，上下文感知（CLAUDE.md）
- ☁️ **Cloudflare 服务** - 内置支持 D1、KV、R2、Durable Objects、Queues

## 🏃‍♂️ 快速开始

使用 npx（推荐）：

```bash
npx create-worker-app my-app
cd my-app
npm install
npm run dev
```

或全局安装：

```bash
npm install -g create-worker-app
create-worker-app my-app
```

### 备选方案：从 GitHub Packages 安装

本包也发布在 [GitHub Packages](./docs/GITHUB_PACKAGES.md)：

```bash
npx @leeguooooo/create-worker-app@latest my-app
```

## 🎮 交互式创建

运行命令后，CLI 会引导你完成项目配置：

```
🚀 Create Worker App

✔ Project name: my-awesome-api
✔ Project description: 高性能 API 服务
✔ Select Cloudflare services to use: 
  ◯ D1 Database (SQLite)
  ◯ KV Storage
  ◯ R2 Object Storage
  ◯ Durable Objects
  ◯ Queues
✔ Include OpenAPI/Swagger documentation? … Yes
✔ Include authentication middleware? … No

📁 Creating project...

✅ Project created successfully!
```

## 🆕 v1.2.0 新功能

### 🤖 Claude Code 集成
每个生成的项目现在包含：
- **CLAUDE.md** - AI 上下文文件，让 Claude Code 更好地理解你的项目
- **INITIAL.md** - 项目需求模板
- 智能代码生成，AI 感知

### ☁️ Cloudflare 服务支持
- **D1 Database** - 边缘 SQLite 数据库
- **KV Storage** - 键值对存储
- **R2 Storage** - S3 兼容对象存储
- **Durable Objects** - 有状态无服务器
- **Queues** - 消息队列

### 🔧 改进
- 修复模板占位符替换
- 更好的依赖管理
- 使用 `.dev.vars` 代替 `.env`
- 优雅的取消操作处理
- 详细的服务设置说明

## 🏗️ 项目结构

```
my-app/
├── src/
│   ├── index.ts          # 应用入口
│   ├── types/            # TypeScript 类型定义
│   │   └── env.ts        # 环境变量类型
│   ├── routes/           # API 路由
│   │   └── health.ts     # 健康检查路由示例
│   ├── schemas/          # Zod schemas 验证
│   │   └── common.ts     # 通用 schema 定义
│   └── lib/              # 工具库
│       └── openapi.ts    # OpenAPI 配置
├── scripts/
│   └── generate-route.js # 路由生成器
├── wrangler.toml         # Cloudflare Workers 配置
├── tsconfig.json         # TypeScript 配置
├── package.json
└── README.md
```

## 🔥 强大的路由生成器

### 交互式模式（推荐）

```bash
npm run generate:route
```

选择你需要的模板：
- **Basic** - 基础 API 路由
- **CRUD Resource** - 完整的增删改查
- **With Auth** - 带认证的路由
- **Webhook Handler** - Webhook 处理器

### 命令行模式

```bash
# 生成基础路由
npm run generate:route createUser post /api/users

# 生成带认证的路由
npm run generate:route getProfile get /api/profile auth

# 生成 CRUD 资源（会创建 5 个端点）
npm run generate:route -- # 然后选择 CRUD Resource
```

### CRUD 生成示例

选择 CRUD Resource 后，会自动生成：

```
✅ Created schema: src/schemas/product.ts
✅ Created CRUD routes: src/routes/product.ts
✅ Updated index.ts

Created endpoints:
- GET    /api/products     - List all products
- GET    /api/products/{id} - Get single product
- POST   /api/products     - Create new product
- PATCH  /api/products/{id} - Update product
- DELETE /api/products/{id} - Delete product
```

## 🤖 与 Claude Code 协作

生成的项目包含 AI 驱动的开发支持：

### CLAUDE.md
为 Claude Code 提供项目上下文：
- 项目结构指南
- 代码风格约定
- Cloudflare Workers 最佳实践
- 开发命令

### INITIAL.md
定义项目需求的模板：
- 功能规格
- API 设计
- 数据模型
- 环境变量
- 外部 API 和文档

只需在 Claude Code 中打开你的项目，它就会自动理解你的代码库结构和需求！

## 🚀 开发和部署

### 本地开发

```bash
npm run dev
# 访问 http://localhost:8787
# API 文档 http://localhost:8787/docs
```

### 部署到 Cloudflare

```bash
# 部署到开发环境
npm run deploy

# 部署到预发布环境
npm run deploy:staging

# 部署到生产环境
npm run deploy:production
```

## 📋 预设模板对比

| 模板 | 用途 | 包含功能 |
|------|------|----------|
| Basic | 标准 API 端点 | 请求验证、错误处理 |
| CRUD Resource | RESTful 资源 | 完整增删改查、分页 |
| With Auth | 需要认证的 API | JWT 验证、用户上下文 |
| Webhook Handler | 接收外部回调 | 签名验证、事件处理 |

## 🔧 配置选项

### Cloudflare 服务

选择 Cloudflare 服务后，`wrangler.toml` 会自动配置：

```toml
# D1 数据库
[[d1_databases]]
binding = "DB"
database_name = "my-app-db"
database_id = "YOUR_DATABASE_ID"

# KV 命名空间
[[kv_namespaces]]
binding = "KV"
id = "YOUR_KV_NAMESPACE_ID"

# R2 存储桶
[[r2_buckets]]
binding = "BUCKET"
bucket_name = "my-app-bucket"
```

### 环境变量

本地开发密钥存储在 `.dev.vars`：

```bash
# 复制示例文件
cp .dev.vars.example .dev.vars

# 生产环境
wrangler secret put JWT_SECRET --env production
```

### 类型安全

所有绑定和环境变量都有完整类型：

```typescript
// src/types/env.ts
export interface Env {
  // Cloudflare 绑定
  DB?: D1Database;
  KV?: KVNamespace;
  BUCKET?: R2Bucket;
  
  // 环境变量
  JWT_SECRET?: string;
  API_KEY?: string;
}
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 License

MIT

---

用 ❤️ 构建，为 Edge Computing 而生。