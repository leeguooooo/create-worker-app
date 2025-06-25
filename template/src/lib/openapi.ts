import { OpenAPIHono } from '@hono/zod-openapi';
import type { Env } from '../types/env';

export function createApp() {
  return new OpenAPIHono<{ Bindings: Env }>({
    defaultHook: (result, c) => {
      if (!result.success) {
        return c.json(
          {
            error: 'Validation failed',
            details: result.error.flatten()
          },
          400
        );
      }
    }
  });
}

export const openAPISpec = {
  openapi: '3.0.0',
  info: {
    title: '{{name}} API',
    version: '1.0.0',
    description: '{{description}}'
  },
  servers: [
    {
      url: 'http://localhost:8787',
      description: 'Development server'
    }
  ]
};