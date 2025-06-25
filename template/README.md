# {{name}}

{{description}}

A Cloudflare Worker service built with Hono.js framework.

---

> 🚀 **Generated with [create-worker-app](https://github.com/leeguooooo/create-worker-app)**  
> The fastest way to build Cloudflare Workers applications

---

## Quick Start

```bash
# Install dependencies
npm install

# Local development
npm run dev

# Deploy to staging
npm run deploy:staging

# Deploy to production
npm run deploy:production
```

## Environment Configuration

### Development

Create a `.dev.vars` file in the root directory:

```
# Example environment variables
DB_PASSWORD=your-dev-password
API_KEY=your-dev-api-key
```

### Staging & Production

Use Wrangler secrets for sensitive data:

```bash
# Staging
wrangler secret put DB_PASSWORD --env staging

# Production
wrangler secret put DB_PASSWORD --env production
```

## API Endpoints

- `GET /` - API information
- `GET /health` - Health check
- `GET /docs` - Swagger UI documentation (development only)
- `GET /openapi.json` - OpenAPI specification

## API Documentation

This project uses `@hono/zod-openapi` to automatically generate API documentation. When writing routes, follow this pattern:

```typescript
// 1. Define Zod Schemas
const RequestSchema = z.object({
  name: z.string().min(1),
  amount: z.number().positive()
});

const ResponseSchema = z.object({
  id: z.string(),
  status: z.string()
});

// 2. Create OpenAPI route
const route = createRoute({
  method: 'post',
  path: '/api/example',
  tags: ['Example'],  // API grouping
  summary: 'Create example',
  request: {
    body: {
      content: {
        'application/json': {
          schema: RequestSchema
        }
      }
    }
  },
  responses: {
    200: {
      content: {
        'application/json': {
          schema: ResponseSchema
        }
      },
      description: 'Success response'
    }
  }
});

// 3. Implement route handler
app.openapi(route, async (c) => {
  const body = c.req.valid('json');
  // Business logic here
  return c.json({ id: '123', status: 'success' });
});
```

Documentation is automatically generated with request/response formats, type validation, and interactive testing.

### Generate API Routes

Use the scaffolding tool to quickly generate API routes:

```bash
# Interactive mode (recommended)
npm run generate:route

# Command line mode
npm run generate:route <name> <method> [path] [template]

# Examples
npm run generate:route createUser post /api/users
npm run generate:route getOrder get
npm run generate:route updateProduct put /api/products/:id auth
```

Available templates:
- **Basic** - Standard API endpoint
- **CRUD Resource** - Full RESTful resource (5 endpoints)
- **With Auth** - JWT authenticated endpoint
- **Webhook Handler** - External webhook receiver

Generated files:
- `src/schemas/<name>.ts` - Request/response schema definitions
- `src/routes/<name>.ts` - Route implementation
- Automatically updates `src/index.ts` to register the route

After generation:
1. Edit the schema file to define data structures
2. Edit the route file to implement business logic
3. Visit `/docs` to see the auto-generated documentation

## Project Structure

```
├── src/
│   ├── index.ts          # Main application entry
│   ├── types/
│   │   └── env.ts        # Environment types
│   ├── lib/
│   │   └── openapi.ts    # OpenAPI configuration
│   ├── schemas/
│   │   └── common.ts     # Common schemas
│   └── routes/
│       └── health.ts     # Health check route
├── scripts/
│   └── generate-route.js # Route generation script
├── wrangler.toml         # Cloudflare Worker config
├── tsconfig.json         # TypeScript config
└── package.json          # Dependencies
```

## Development

```bash
# Start development server
npm run dev

# The service will be available at:
# http://localhost:8787
```

## Deployment

Before deploying, ensure you have:
1. A Cloudflare account
2. Wrangler CLI authenticated (`wrangler login`)
3. Required secrets configured

```bash
# Deploy to staging
npm run deploy:staging

# Deploy to production
npm run deploy:production
```

## Contributing

1. Create a feature branch
2. Make your changes
3. Write/update tests
4. Submit a pull request

## License

ISC

---

<div align="center">
  <sub>Built with ❤️ using <a href="https://github.com/leeguooooo/create-worker-app">create-worker-app</a></sub>
</div>