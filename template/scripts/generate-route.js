#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const prompts = require('prompts');

async function generateRoute() {
  const args = process.argv.slice(2);
  
  // 如果没有参数，使用交互式模式
  if (args.length === 0) {
    const answers = await prompts([
      {
        type: 'text',
        name: 'name',
        message: 'Route name (e.g., getUserById):',
        validate: value => value.length > 0 || 'Name is required'
      },
      {
        type: 'select',
        name: 'method',
        message: 'HTTP method:',
        choices: [
          { title: 'GET', value: 'get' },
          { title: 'POST', value: 'post' },
          { title: 'PUT', value: 'put' },
          { title: 'PATCH', value: 'patch' },
          { title: 'DELETE', value: 'delete' }
        ]
      },
      {
        type: 'text',
        name: 'path',
        message: 'Route path:',
        initial: (prev, values) => {
          const fileName = values.name.replace(/([A-Z])/g, '-$1').toLowerCase().replace(/^-/, '');
          return `/api/${fileName}`;
        }
      },
      {
        type: 'select',
        name: 'template',
        message: 'Template:',
        choices: [
          { title: 'Basic', value: 'basic' },
          { title: 'CRUD Resource', value: 'crud' },
          { title: 'With Auth', value: 'auth' },
          { title: 'Webhook Handler', value: 'webhook' }
        ]
      }
    ]);
    
    if (!answers.name) {
      process.exit(0);
    }
    
    if (answers.template === 'crud') {
      await generateCRUD(answers.name, answers.path);
    } else {
      await generateSingleRoute(answers.name, answers.method, answers.path, answers.template);
    }
  } else {
    // 命令行模式
    if (args.length < 2) {
      console.error('Usage: npm run generate:route <name> <method> [path] [template]');
      console.error('Example: npm run generate:route createPayment post /api/payment basic');
      console.error('Or run without arguments for interactive mode');
      process.exit(1);
    }
    
    const [name, method, routePath, template = 'basic'] = args;
    await generateSingleRoute(name, method, routePath, template);
  }
}

async function generateSingleRoute(name, method, routePath, template) {
  const fileName = name.replace(/([A-Z])/g, '-$1').toLowerCase().replace(/^-/, '');
  const pascalCase = name.charAt(0).toUpperCase() + name.slice(1);
  const tag = pascalCase.replace(/([A-Z])/g, ' $1').trim().split(' ')[0];
  const actualPath = routePath || `/api/${fileName}`;
  
  // Schema content based on template
  let schemaContent = '';
  let routeContent = '';
  
  if (template === 'auth') {
    schemaContent = `import { z } from 'zod';

export const ${pascalCase}RequestSchema = z.object({
  // TODO: Define request schema
  data: z.object({
    example: z.string().min(1)
  })
});

export const ${pascalCase}ResponseSchema = z.object({
  success: z.boolean(),
  data: z.object({
    id: z.string(),
    message: z.string()
  }),
  timestamp: z.string()
});
`;

    routeContent = `import { createRoute } from '@hono/zod-openapi';
import { createApp } from '../lib/openapi';
import { ${pascalCase}RequestSchema, ${pascalCase}ResponseSchema } from '../schemas/${fileName}';
import { ErrorResponseSchema } from '../schemas/common';
import type { Env } from '../types/env';

const route = createRoute({
  method: '${method}',
  path: '${actualPath}',
  tags: ['${tag}'],
  summary: 'TODO: Add summary',
  description: 'TODO: Add detailed description',
  security: [{ bearerAuth: [] }],
  request: {
    body: {
      content: {
        'application/json': {
          schema: ${pascalCase}RequestSchema
        }
      }
    }
  },
  responses: {
    200: {
      content: {
        'application/json': {
          schema: ${pascalCase}ResponseSchema
        }
      },
      description: 'Successful response'
    },
    401: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: 'Unauthorized'
    },
    400: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: 'Bad request'
    },
    500: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: 'Internal server error'
    }
  }
});

const app = createApp<Env>();

// Auth middleware
app.use('${actualPath}', async (c, next) => {
  const token = c.req.header('Authorization')?.replace('Bearer ', '');
  
  if (!token) {
    return c.json({ error: 'Unauthorized' }, 401);
  }
  
  // TODO: Verify token
  // const payload = await verifyJWT(token, c.env.JWT_SECRET);
  // c.set('user', payload);
  
  await next();
});

app.openapi(route, async (c) => {
  const body = c.req.valid('json');
  // const user = c.get('user');
  
  // TODO: Implement business logic with auth context
  
  return c.json({
    success: true,
    data: {
      id: crypto.randomUUID(),
      message: 'Success'
    },
    timestamp: new Date().toISOString()
  });
});

export default app;
`;
  } else if (template === 'webhook') {
    schemaContent = `import { z } from 'zod';

export const ${pascalCase}RequestSchema = z.object({
  event: z.string(),
  timestamp: z.string(),
  data: z.record(z.any())
});

export const ${pascalCase}ResponseSchema = z.object({
  received: z.boolean()
});
`;

    routeContent = `import { createRoute } from '@hono/zod-openapi';
import { createApp } from '../lib/openapi';
import { ${pascalCase}RequestSchema, ${pascalCase}ResponseSchema } from '../schemas/${fileName}';
import { ErrorResponseSchema } from '../schemas/common';
import type { Env } from '../types/env';

const route = createRoute({
  method: 'post',
  path: '${actualPath}',
  tags: ['Webhooks'],
  summary: 'Handle ${tag} webhook',
  description: 'Receives and processes ${tag} webhook events',
  request: {
    body: {
      content: {
        'application/json': {
          schema: ${pascalCase}RequestSchema
        }
      }
    },
    headers: z.object({
      'x-webhook-signature': z.string().optional()
    })
  },
  responses: {
    200: {
      content: {
        'application/json': {
          schema: ${pascalCase}ResponseSchema
        }
      },
      description: 'Webhook received'
    },
    400: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: 'Invalid webhook payload'
    }
  }
});

const app = createApp<Env>();

app.openapi(route, async (c) => {
  const body = c.req.valid('json');
  const signature = c.req.header('x-webhook-signature');
  
  // TODO: Verify webhook signature
  // if (signature && !verifyWebhookSignature(body, signature, c.env.WEBHOOK_SECRET)) {
  //   return c.json({ error: 'Invalid signature' }, 400);
  // }
  
  console.log('Webhook received:', body.event, body.timestamp);
  
  // TODO: Process webhook event
  // await processWebhookEvent(body);
  
  return c.json({ received: true });
});

export default app;
`;
  } else {
    // Basic template
    schemaContent = `import { z } from 'zod';

export const ${pascalCase}RequestSchema = z.object({
  // TODO: Define request schema
  example: z.string().min(1)
});

export const ${pascalCase}ResponseSchema = z.object({
  // TODO: Define response schema
  id: z.string(),
  message: z.string()
});
`;

    routeContent = `import { createRoute } from '@hono/zod-openapi';
import { createApp } from '../lib/openapi';
import { ${pascalCase}RequestSchema, ${pascalCase}ResponseSchema } from '../schemas/${fileName}';
import { ErrorResponseSchema } from '../schemas/common';

const route = createRoute({
  method: '${method}',
  path: '${actualPath}',
  tags: ['${tag}'],
  summary: 'TODO: Add summary',
  description: 'TODO: Add detailed description',
  request: {
    body: {
      content: {
        'application/json': {
          schema: ${pascalCase}RequestSchema
        }
      }
    }
  },
  responses: {
    200: {
      content: {
        'application/json': {
          schema: ${pascalCase}ResponseSchema
        }
      },
      description: 'Successful response'
    },
    400: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: 'Bad request'
    },
    500: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: 'Internal server error'
    }
  }
});

const app = createApp();

app.openapi(route, async (c) => {
  const body = c.req.valid('json');
  
  // TODO: Implement business logic
  
  return c.json({
    id: crypto.randomUUID(),
    message: 'Success'
  });
});

export default app;
`;
  }
  
  // Create files
  const schemaPath = path.join(__dirname, '..', 'src', 'schemas', `${fileName}.ts`);
  fs.writeFileSync(schemaPath, schemaContent);
  console.log(`✅ Created schema: ${schemaPath}`);
  
  const routeFilePath = path.join(__dirname, '..', 'src', 'routes', `${fileName}.ts`);
  fs.writeFileSync(routeFilePath, routeContent);
  console.log(`✅ Created route: ${routeFilePath}`);
  
  // Update index.ts
  updateIndex(name, fileName);
  
  console.log(`
🎉 Route generated successfully!

Next steps:
1. Edit src/schemas/${fileName}.ts to define request/response schemas
2. Edit src/routes/${fileName}.ts to implement business logic
3. Run 'npm run dev' and visit http://localhost:8787/docs
`);
}

async function generateCRUD(resourceName, basePath) {
  const singular = resourceName.toLowerCase();
  const plural = singular.endsWith('s') ? singular : singular + 's';
  const pascalCase = resourceName.charAt(0).toUpperCase() + resourceName.slice(1);
  const fileName = singular;
  
  // Generate CRUD schema
  const schemaContent = `import { z } from 'zod';

export const ${pascalCase}Schema = z.object({
  id: z.string(),
  name: z.string().min(1),
  description: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string()
});

export const Create${pascalCase}Schema = z.object({
  name: z.string().min(1),
  description: z.string().optional()
});

export const Update${pascalCase}Schema = z.object({
  name: z.string().min(1).optional(),
  description: z.string().optional()
});

export const ${pascalCase}ListSchema = z.object({
  items: z.array(${pascalCase}Schema),
  total: z.number(),
  page: z.number(),
  limit: z.number()
});
`;

  // Generate CRUD routes
  const routeContent = `import { createRoute } from '@hono/zod-openapi';
import { createApp } from '../lib/openapi';
import { z } from 'zod';
import { 
  ${pascalCase}Schema, 
  Create${pascalCase}Schema, 
  Update${pascalCase}Schema,
  ${pascalCase}ListSchema 
} from '../schemas/${fileName}';
import { ErrorResponseSchema } from '../schemas/common';

const app = createApp();
const basePath = '${basePath || `/api/${plural}`}';

// List all ${plural}
const listRoute = createRoute({
  method: 'get',
  path: basePath,
  tags: ['${pascalCase}'],
  summary: 'List all ${plural}',
  request: {
    query: z.object({
      page: z.string().optional(),
      limit: z.string().optional()
    })
  },
  responses: {
    200: {
      content: {
        'application/json': {
          schema: ${pascalCase}ListSchema
        }
      },
      description: 'List of ${plural}'
    }
  }
});

// Get single ${singular}
const getRoute = createRoute({
  method: 'get',
  path: \`\${basePath}/{id}\`,
  tags: ['${pascalCase}'],
  summary: 'Get ${singular} by ID',
  request: {
    params: z.object({
      id: z.string()
    })
  },
  responses: {
    200: {
      content: {
        'application/json': {
          schema: ${pascalCase}Schema
        }
      },
      description: '${pascalCase} details'
    },
    404: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: '${pascalCase} not found'
    }
  }
});

// Create ${singular}
const createRoute = createRoute({
  method: 'post',
  path: basePath,
  tags: ['${pascalCase}'],
  summary: 'Create new ${singular}',
  request: {
    body: {
      content: {
        'application/json': {
          schema: Create${pascalCase}Schema
        }
      }
    }
  },
  responses: {
    201: {
      content: {
        'application/json': {
          schema: ${pascalCase}Schema
        }
      },
      description: '${pascalCase} created'
    },
    400: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: 'Bad request'
    }
  }
});

// Update ${singular}
const updateRoute = createRoute({
  method: 'patch',
  path: \`\${basePath}/{id}\`,
  tags: ['${pascalCase}'],
  summary: 'Update ${singular}',
  request: {
    params: z.object({
      id: z.string()
    }),
    body: {
      content: {
        'application/json': {
          schema: Update${pascalCase}Schema
        }
      }
    }
  },
  responses: {
    200: {
      content: {
        'application/json': {
          schema: ${pascalCase}Schema
        }
      },
      description: '${pascalCase} updated'
    },
    404: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: '${pascalCase} not found'
    }
  }
});

// Delete ${singular}
const deleteRoute = createRoute({
  method: 'delete',
  path: \`\${basePath}/{id}\`,
  tags: ['${pascalCase}'],
  summary: 'Delete ${singular}',
  request: {
    params: z.object({
      id: z.string()
    })
  },
  responses: {
    204: {
      description: '${pascalCase} deleted'
    },
    404: {
      content: {
        'application/json': {
          schema: ErrorResponseSchema
        }
      },
      description: '${pascalCase} not found'
    }
  }
});

// Implement routes
app.openapi(listRoute, async (c) => {
  const { page = '1', limit = '10' } = c.req.query();
  
  // TODO: Implement list logic
  return c.json({
    items: [],
    total: 0,
    page: parseInt(page),
    limit: parseInt(limit)
  });
});

app.openapi(getRoute, async (c) => {
  const { id } = c.req.param();
  
  // TODO: Implement get logic
  return c.json({
    id,
    name: 'Example',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  });
});

app.openapi(createRoute, async (c) => {
  const body = c.req.valid('json');
  
  // TODO: Implement create logic
  return c.json({
    id: crypto.randomUUID(),
    ...body,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  }, 201);
});

app.openapi(updateRoute, async (c) => {
  const { id } = c.req.param();
  const body = c.req.valid('json');
  
  // TODO: Implement update logic
  return c.json({
    id,
    name: body.name || 'Updated',
    description: body.description,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  });
});

app.openapi(deleteRoute, async (c) => {
  const { id } = c.req.param();
  
  // TODO: Implement delete logic
  return c.status(204);
});

export default app;
`;

  // Create files
  const schemaPath = path.join(__dirname, '..', 'src', 'schemas', `${fileName}.ts`);
  fs.writeFileSync(schemaPath, schemaContent);
  console.log(`✅ Created schema: ${schemaPath}`);
  
  const routeFilePath = path.join(__dirname, '..', 'src', 'routes', `${fileName}.ts`);
  fs.writeFileSync(routeFilePath, routeContent);
  console.log(`✅ Created CRUD routes: ${routeFilePath}`);
  
  // Update index.ts
  updateIndex(resourceName, fileName);
  
  console.log(`
🎉 CRUD routes generated successfully!

Created endpoints:
- GET    ${basePath}     - List all ${plural}
- GET    ${basePath}/{id} - Get single ${singular}
- POST   ${basePath}     - Create new ${singular}
- PATCH  ${basePath}/{id} - Update ${singular}
- DELETE ${basePath}/{id} - Delete ${singular}

Next steps:
1. Implement the TODO sections in src/routes/${fileName}.ts
2. Run 'npm run dev' and visit http://localhost:8787/docs
`);
}

function updateIndex(name, fileName) {
  const indexPath = path.join(__dirname, '..', 'src', 'index.ts');
  let indexContent = fs.readFileSync(indexPath, 'utf8');
  
  // Add import
  const importLine = `import ${name}Routes from './routes/${fileName}';`;
  const lastImportIndex = indexContent.lastIndexOf('import');
  const nextLineIndex = indexContent.indexOf('\n', lastImportIndex);
  indexContent = indexContent.slice(0, nextLineIndex + 1) + importLine + '\n' + indexContent.slice(nextLineIndex + 1);
  
  // Add route
  const routeLine = `app.route('/', ${name}Routes);`;
  const routesComment = '// Routes';
  const routesIndex = indexContent.indexOf(routesComment);
  const routesNextLine = indexContent.indexOf('\n', routesIndex);
  const nextRouteIndex = indexContent.indexOf('app.route', routesNextLine);
  indexContent = indexContent.slice(0, nextRouteIndex) + routeLine + '\n' + indexContent.slice(nextRouteIndex);
  
  fs.writeFileSync(indexPath, indexContent);
  console.log(`✅ Updated index.ts`);
}

// Run
generateRoute().catch(console.error);