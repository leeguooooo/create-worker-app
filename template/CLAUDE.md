# CLAUDE.md

本文件为 Claude Code 在处理此 Cloudflare Worker 项目时提供指导。

## 项目概述

这是一个 Cloudflare Worker 应用，使用以下技术栈：
- **Hono.js**：快速、轻量的边缘计算 Web 框架
- **TypeScript**：提供类型安全和更好的开发体验
- **Wrangler**：Cloudflare 的 Workers 开发 CLI 工具
- **OpenAPI**：自动生成 API 文档

## 开发命令

```bash
npm install                    # 安装依赖
npm run dev                    # 启动开发服务器
npm run deploy                 # 部署到 Cloudflare
npm run generate:route         # 交互式生成新的 API 路由
npm run cf-typegen             # 为绑定生成 TypeScript 类型
```

## 代码风格指南

### TypeScript 与 JavaScript
- 所有新代码使用 TypeScript
- 优先使用 `const`，其次 `let`，避免 `var`
- 尽可能使用 async/await 而非 promises
- 使用可选链操作符（`?.`）和空值合并运算符（`??`）
- 始终为函数参数和返回值定义正确的类型

### API 开发
- 将路由保存在 `src/routes/` 目录下的独立文件中
- 在 `src/schemas/` 中使用 Zod schemas 进行请求/响应验证
- 遵循 RESTful API 规范
- 始终包含适当的错误处理和有意义的错误信息

### 文件结构
```
src/
├── index.ts           # 应用程序主入口
├── routes/            # API 路由处理器
├── schemas/           # Zod 验证模式
├── lib/              # 共享工具和辅助函数
└── types/            # TypeScript 类型定义
```

### 代码组织
- **文件不超过 300 行**。接近此限制时拆分为模块
- **每个路由一个文件**，放在 routes 目录中
- **相关功能分组**在 lib 模块中
- **模式文件独立**但与使用位置保持接近

### 测试与部署
- 部署前使用 `npm run dev` 进行本地测试
- 在 `wrangler.toml` 中使用环境变量配置
- 绝不提交敏感数据或 API 密钥
- 始终使用 cf-typegen 验证环境绑定

### 文档
- 添加新功能或更改设置时更新 README.md
- 用清晰的示例记录 API 端点
- 为复杂函数使用 JSDoc 注释
- 保持内联注释简洁有意义

### 最佳实践
- 使用路由生成器（`npm run generate:route`）保持一致性
- 需要时实现适当的 CORS 头
- 使用 Cloudflare KV、D1 或 R2 进行数据持久化
- 利用 Workers 的全局 fetch 调用外部 API
- 使用适当的状态码优雅地处理错误

### 常见模式

#### 创建新路由
```typescript
// src/routes/example.ts
import { Hono } from 'hono';
import { zValidator } from '@hono/zod-validator';
import { exampleSchema } from '../schemas/example';

const app = new Hono();

app.post('/example', zValidator('json', exampleSchema), async (c) => {
  const data = c.req.valid('json');
  // 实现逻辑
  return c.json({ success: true, data });
});

export default app;
```

#### 错误处理
```typescript
app.onError((err, c) => {
  console.error(`错误: ${err.message}`);
  return c.json({ error: err.message }, 500);
});
```

## 环境变量
在 `wrangler.toml` 中定义：
- KV 命名空间
- D1 数据库
- R2 存储桶
- 环境密钥

## 重要说明
- 这是 Cloudflare Worker，不是 Node.js 应用
- 全局 API 可能与 Node.js 不同（使用 Web API）
- 保持小包体积以获得最佳性能
- 使用 Wrangler 进行本地开发和部署