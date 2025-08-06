# create-worker-app

🚀 The fastest way to create Cloudflare Workers applications with Hono.js

[简体中文](./README.zh-CN.md)

## ✨ Features

- ⚡️ **Lightning Fast** - Built on [Hono.js](https://hono.dev/), optimized for edge computing
- 📝 **TypeScript First** - Full type support and IntelliSense
- 📚 **Auto API Documentation** - OpenAPI/Swagger integration, visit `/docs`
- 🛠️ **Smart Route Generator** - Interactive CLI for creating CRUD, Auth, Webhook templates
- 🎯 **Production Ready** - Built-in error handling, CORS, logging middleware
- 🚀 **One-Click Deploy** - Pre-configured Wrangler for multi-environment deployment
- 🤖 **Claude Code Integration** - AI-powered development with context awareness (CLAUDE.md)
- ☁️ **Cloudflare Services** - Built-in support for D1, KV, R2, Durable Objects, Queues

## 🏃‍♂️ Quick Start

Using npx (recommended):

```bash
npx create-worker-app my-app
cd my-app
npm install
npm run dev
```

Or install globally:

```bash
npm install -g create-worker-app
create-worker-app my-app
```

### Alternative: Install from GitHub Packages

This package is also available on [GitHub Packages](./docs/GITHUB_PACKAGES.md):

```bash
npx @leeguooooo/create-worker-app@latest my-app
```

## 🎮 Interactive Setup

The CLI will guide you through the setup:

```
🚀 Create Worker App

✔ Project name: my-awesome-api
✔ Project description: A high-performance API service
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

## 🆕 What's New in v1.2.0

### 🤖 Claude Code Integration
Each generated project now includes:
- **CLAUDE.md** - AI context file for Claude Code assistance
- **INITIAL.md** - Project requirements template
- Smart code generation with AI awareness

### ☁️ Cloudflare Services Support
- **D1 Database** - SQLite at the edge
- **KV Storage** - Key-value store
- **R2 Storage** - S3-compatible object storage
- **Durable Objects** - Stateful serverless
- **Queues** - Message queuing

### 🔧 Improvements
- Fixed template placeholder replacement
- Better dependency management
- Use `.dev.vars` instead of `.env`
- Graceful cancellation handling
- Detailed service setup instructions

## 🏗️ Project Structure

```
my-app/
├── src/
│   ├── index.ts          # Application entry
│   ├── types/            # TypeScript type definitions
│   │   └── env.ts        # Environment types
│   ├── routes/           # API routes
│   │   └── health.ts     # Health check example
│   ├── schemas/          # Zod validation schemas
│   │   └── common.ts     # Common schemas
│   └── lib/              # Utilities
│       └── openapi.ts    # OpenAPI configuration
├── scripts/
│   └── generate-route.js # Route generator
├── wrangler.toml         # Cloudflare Workers config
├── tsconfig.json         # TypeScript config
├── package.json
└── README.md
```

## 🔥 Powerful Route Generator

### Interactive Mode (Recommended)

```bash
npm run generate:route
```

Choose from templates:
- **Basic** - Standard API route
- **CRUD Resource** - Full REST endpoints
- **With Auth** - JWT authenticated route
- **Webhook Handler** - External webhook receiver

### CLI Mode

```bash
# Generate basic route
npm run generate:route createUser post /api/users

# Generate authenticated route
npm run generate:route getProfile get /api/profile auth

# Generate CRUD resource (creates 5 endpoints)
npm run generate:route -- # Then select CRUD Resource
```

### CRUD Generation Example

When selecting CRUD Resource:

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

## 🤖 Working with Claude Code

Generated projects include AI-powered development support:

### CLAUDE.md
Provides context for Claude Code to understand your project:
- Project structure guidelines
- Code style conventions  
- Cloudflare Workers best practices
- Development commands

### INITIAL.md
Template for defining project requirements:
- Feature specifications
- API design
- Data models
- Environment variables
- External APIs and documentation

Simply open your project in Claude Code and it will automatically understand your codebase structure and requirements!

## 🚀 Development & Deployment

### Local Development

```bash
npm run dev
# Visit http://localhost:8787
# API docs at http://localhost:8787/docs
```

### Deploy to Cloudflare

```bash
# Deploy to development
npm run deploy

# Deploy to staging
npm run deploy:staging

# Deploy to production
npm run deploy:production
```

## 📋 Template Comparison

| Template | Use Case | Features |
|----------|----------|----------|
| Basic | Standard API endpoints | Request validation, error handling |
| CRUD Resource | RESTful resources | Full CRUD operations, pagination |
| With Auth | Protected APIs | JWT validation, user context |
| Webhook Handler | External callbacks | Signature verification, event handling |

## 🔧 Configuration

### Cloudflare Services

When you select Cloudflare services during setup, the `wrangler.toml` will be automatically configured:

```toml
# D1 Database
[[d1_databases]]
binding = "DB"
database_name = "my-app-db"
database_id = "YOUR_DATABASE_ID"

# KV Namespace
[[kv_namespaces]]
binding = "KV"
id = "YOUR_KV_NAMESPACE_ID"

# R2 Bucket
[[r2_buckets]]
binding = "BUCKET"
bucket_name = "my-app-bucket"
```

### Environment Variables

Local development secrets go in `.dev.vars`:

```bash
# Copy the example file
cp .dev.vars.example .dev.vars

# For production
wrangler secret put JWT_SECRET --env production
```

### Type Safety

All bindings and environment variables are fully typed:

```typescript
// src/types/env.ts
export interface Env {
  // Cloudflare Bindings
  DB?: D1Database;
  KV?: KVNamespace;
  BUCKET?: R2Bucket;
  
  // Environment Variables
  JWT_SECRET?: string;
  API_KEY?: string;
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT

---

<div align="center">
  <sub>Built with ❤️ for Edge Computing</sub>
</div>