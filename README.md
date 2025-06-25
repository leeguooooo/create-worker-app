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

## 🎮 Interactive Setup

The CLI will guide you through the setup:

```
🚀 Create Worker App

✔ Project name: my-awesome-api
✔ Project description: A high-performance API service
✔ Will you need database configuration? … No
✔ Include OpenAPI/Swagger documentation? … Yes

📁 Creating project...

✅ Project created successfully!
```

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

### Database Support

If you enable database configuration, a `.env.example` file will be created:

```env
DB_HOST=
DB_PORT=
DB_NAME=
DB_USER=
DB_PASSWORD=
```

### Environment Types

All environment variables are fully typed:

```typescript
// src/types/env.ts
export interface Env {
  // Your environment variables
  API_KEY: string;
  DB_URL?: string;
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