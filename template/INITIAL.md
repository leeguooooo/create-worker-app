## 功能需求

[在此描述你要构建的 Cloudflare Worker 功能，例如：
- 构建一个 REST API 服务
- 实现 webhook 处理器
- 创建代理服务
- 开发边缘函数
- 实现认证网关等]

## 示例代码

[说明 `examples/` 文件夹中的示例代码及其用途，例如：
- API 路由示例
- 中间件使用
- 数据库连接
- 外部 API 调用等]

## 技术栈与依赖

### 核心框架
- **Hono.js** - 轻量级 Web 框架
- **TypeScript** - 类型安全
- **Wrangler** - Cloudflare CLI 工具

### 可选服务
- **KV 存储** - 键值对存储
- **D1 数据库** - SQLite 数据库
- **R2 存储** - 对象存储
- **Durable Objects** - 有状态服务
- **Queues** - 消息队列

## API 设计

[列出计划的 API 端点，例如：]
```
GET    /api/users          # 获取用户列表
POST   /api/users          # 创建用户
GET    /api/users/:id      # 获取单个用户
PUT    /api/users/:id      # 更新用户
DELETE /api/users/:id      # 删除用户
```

## 数据模型

[定义主要的数据结构，使用 TypeScript 接口或 Zod schemas：]
```typescript
interface User {
  id: string;
  name: string;
  email: string;
  createdAt: Date;
}
```

## 环境变量

[列出需要的环境变量：]
```
# 必需
DATABASE_URL=          # D1 数据库连接
API_KEY=              # API 密钥

# 可选
ENVIRONMENT=          # development/staging/production
LOG_LEVEL=           # debug/info/warn/error
```

## 外部 API 与文档

[列出开发时需要参考的文档和 API：]
- Cloudflare Workers 文档：https://developers.cloudflare.com/workers/
- Hono.js 文档：https://hono.dev/
- Wrangler 配置：https://developers.cloudflare.com/workers/wrangler/configuration/
- [其他相关 API 文档]

## 特殊要求与注意事项

### 性能要求
- 响应时间 < 100ms
- 包大小 < 1MB
- 支持并发请求

### 安全要求
- JWT 认证
- 速率限制
- CORS 配置
- 输入验证

### 部署要求
- 多环境支持（开发/预发/生产）
- 自动化部署流程
- 监控和日志

### 开发规范
- 使用 TypeScript 严格模式
- 所有 API 需要 OpenAPI 文档
- 错误处理规范化
- 单元测试覆盖

## 项目结构规划

```
src/
├── routes/          # API 路由
│   ├── users.ts
│   └── auth.ts
├── schemas/         # 数据验证模式
│   ├── user.ts
│   └── auth.ts
├── middleware/      # 中间件
│   ├── auth.ts
│   └── cors.ts
├── services/        # 业务逻辑
│   └── userService.ts
├── utils/           # 工具函数
│   └── crypto.ts
└── types/          # 类型定义
    └── custom.ts
```

## 开发步骤建议

1. **初始化项目**
   - 配置 wrangler.toml
   - 设置环境变量
   - 配置 TypeScript

2. **基础架构**
   - 设置路由结构
   - 配置中间件
   - 错误处理

3. **核心功能**
   - 实现数据模型
   - 开发 API 端点
   - 添加验证逻辑

4. **集成与测试**
   - 连接外部服务
   - 编写测试用例
   - 性能优化

5. **部署**
   - 配置 CI/CD
   - 设置监控
   - 文档完善

## 常见问题与解决方案

[记录开发中可能遇到的问题：]
- Worker 大小限制：使用代码分割和懒加载
- 冷启动优化：预热策略
- 跨域问题：正确配置 CORS 中间件
- 环境变量管理：使用 wrangler secrets