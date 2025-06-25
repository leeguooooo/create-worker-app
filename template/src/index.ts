/**
 * Generated with create-worker-app
 * https://github.com/leeguooooo/create-worker-app
 * 
 * A high-performance Cloudflare Worker application built with Hono.js
 */

import { swaggerUI } from '@hono/swagger-ui';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { createApp, openAPISpec } from './lib/openapi';
import healthRoutes from './routes/health';

const app = createApp();

// Middleware
app.use('*', logger());
app.use('*', cors());

// Routes
app.route('/', healthRoutes);

// API Documentation
app.get('/', (c) => {
  return c.json({ 
    message: '{{name}} API',
    version: '1.0.0',
    docs: '/docs'
  });
});

// Swagger UI - only in development
app.get('/docs', swaggerUI({ url: '/openapi.json' }));

// OpenAPI spec
app.get('/openapi.json', (c) => {
  return c.json({
    ...openAPISpec,
    ...app.getOpenAPI31Document(openAPISpec)
  });
});

// 404 handler
app.notFound((c) => {
  return c.json({ error: 'Not Found' }, 404);
});

// Error handler
app.onError((err, c) => {
  console.error(`${err}`);
  return c.json({ error: 'Internal Server Error' }, 500);
});

export default app;