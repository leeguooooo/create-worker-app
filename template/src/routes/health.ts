import { createRoute } from '@hono/zod-openapi';
import { createApp } from '../lib/openapi';
import { HealthResponseSchema } from '../schemas/common';

const route = createRoute({
  method: 'get',
  path: '/health',
  tags: ['System'],
  summary: 'Health check endpoint',
  responses: {
    200: {
      content: {
        'application/json': {
          schema: HealthResponseSchema
        }
      },
      description: 'Service health status'
    }
  }
});

const app = createApp();

app.openapi(route, (c) => {
  return c.json({
    status: 'healthy',
    timestamp: new Date().toISOString()
  });
});

export default app;