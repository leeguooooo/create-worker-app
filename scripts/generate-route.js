#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const args = process.argv.slice(2);
if (args.length < 2) {
  console.error('Usage: npm run generate:route <name> <method> [path]');
  console.error('Example: npm run generate:route createPayment post /api/payment');
  process.exit(1);
}

const [name, method, routePath] = args;
const fileName = name.replace(/([A-Z])/g, '-$1').toLowerCase().replace(/^-/, '');
const pascalCase = name.charAt(0).toUpperCase() + name.slice(1);
const tag = pascalCase.replace(/([A-Z])/g, ' $1').trim().split(' ')[0];
const actualPath = routePath || `/api/${fileName}`;

const schemaContent = `import { z } from 'zod';

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

const routeContent = `import { createRoute } from '@hono/zod-openapi';
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

// Create schema file
const schemaPath = path.join(__dirname, '..', 'src', 'schemas', `${fileName}.ts`);
fs.writeFileSync(schemaPath, schemaContent);
console.log(`✅ Created schema: ${schemaPath}`);

// Create route file
const routePath = path.join(__dirname, '..', 'src', 'routes', `${fileName}.ts`);
fs.writeFileSync(routePath, routeContent);
console.log(`✅ Created route: ${routePath}`);

// Update index.ts
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

console.log(`
🎉 Route generated successfully!

Next steps:
1. Edit src/schemas/${fileName}.ts to define request/response schemas
2. Edit src/routes/${fileName}.ts to implement business logic
3. Run 'npm run dev' and visit http://localhost:8787/docs
`);