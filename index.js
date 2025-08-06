#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const prompts = require('prompts');
const { red, green, yellow, blue, cyan } = require('kolorist');
const minimist = require('minimist');
const { execSync } = require('child_process');

const argv = minimist(process.argv.slice(2));

// Detect package manager
function detectPackageManager() {
  const userAgent = process.env.npm_config_user_agent || '';
  
  if (userAgent.includes('pnpm')) {
    return 'pnpm';
  } else if (userAgent.includes('yarn')) {
    return 'yarn';
  } else if (userAgent.includes('bun')) {
    return 'bun';
  }
  
  // Fallback: check if pnpm/yarn/bun are installed
  try {
    execSync('pnpm --version', { stdio: 'ignore' });
    return 'pnpm';
  } catch {}
  
  try {
    execSync('yarn --version', { stdio: 'ignore' });
    return 'yarn';
  } catch {}
  
  try {
    execSync('bun --version', { stdio: 'ignore' });
    return 'bun';
  } catch {}
  
  return 'npm';
}

async function init() {
  // Check if running in CI/test environment
  const isCI = process.env.CI || process.env.NODE_ENV === 'test';
  
  if (!isCI) {
    console.log(cyan('\n🚀 Create Worker App\n'));
  }

  let targetDir = argv._[0];
  let projectName = targetDir;

  if (!targetDir) {
    const response = await prompts({
      type: 'text',
      name: 'projectName',
      message: 'Project name:',
      initial: 'my-worker-app'
    });
    
    if (!response.projectName) {
      console.log(yellow('\n✖ Operation cancelled'));
      process.exit(0);
    }
    
    projectName = response.projectName;
    targetDir = projectName;
  }

  const projectPath = path.join(process.cwd(), targetDir);

  if (fs.existsSync(projectPath)) {
    console.log(red(`Error: Directory ${targetDir} already exists!`));
    process.exit(1);
  }

  const questions = [
    {
      type: 'text',
      name: 'description',
      message: 'Project description:',
      initial: 'A Cloudflare Worker application'
    },
    {
      type: 'multiselect',
      name: 'features',
      message: 'Select Cloudflare services to use:',
      choices: [
        { title: 'D1 Database (SQLite)', value: 'd1', selected: false },
        { title: 'KV Storage', value: 'kv', selected: false },
        { title: 'R2 Object Storage', value: 'r2', selected: false },
        { title: 'Durable Objects', value: 'do', selected: false },
        { title: 'Queues', value: 'queues', selected: false }
      ],
      hint: 'Space to select, Enter to confirm'
    },
    {
      type: 'confirm',
      name: 'useOpenAPI',
      message: 'Include OpenAPI/Swagger documentation?',
      initial: true
    },
    {
      type: 'confirm',
      name: 'useAuth',
      message: 'Include authentication middleware?',
      initial: false
    },
    {
      type: 'confirm',
      name: 'installDeps',
      message: 'Install dependencies now?',
      initial: true
    }
  ];

  const answers = await prompts(questions);
  
  // Check if user cancelled
  if (answers.description === undefined) {
    console.log(yellow('\n✖ Operation cancelled'));
    process.exit(0);
  }

  console.log(blue('\n📁 Creating project...\n'));

  // Create project directory
  fs.mkdirSync(projectPath, { recursive: true });

  // Copy template files
  const templateDir = path.join(__dirname, 'template');
  copyDir(templateDir, projectPath);

  // Update package.json
  const packageJson = JSON.parse(fs.readFileSync(path.join(projectPath, 'package.json'), 'utf8'));
  packageJson.name = projectName;
  packageJson.description = answers.description;
  
  // Add metadata to identify the generator
  packageJson.generator = {
    name: 'create-worker-app',
    version: require('./package.json').version,
    timestamp: new Date().toISOString()
  };

  // Remove OpenAPI dependencies if not selected
  if (!answers.useOpenAPI) {
    delete packageJson.dependencies['@hono/zod-openapi'];
    delete packageJson.dependencies['@hono/swagger-ui'];
    delete packageJson.dependencies['zod'];
  }

  fs.writeFileSync(
    path.join(projectPath, 'package.json'),
    JSON.stringify(packageJson, null, 2)
  );

  // Update wrangler.toml based on selected features
  let wranglerContent = fs.readFileSync(path.join(projectPath, 'wrangler.toml'), 'utf8');
  
  // Add configuration for selected Cloudflare services
  const features = answers.features || [];
  let bindingsConfig = '\n# Cloudflare service bindings\n';
  
  if (features.includes('d1')) {
    bindingsConfig += `
# D1 Database binding
[[d1_databases]]
binding = "DB" # Available as env.DB
database_name = "${projectName}-db"
database_id = "YOUR_DATABASE_ID" # Replace with actual D1 database ID
`;
  }
  
  if (features.includes('kv')) {
    bindingsConfig += `
# KV Namespace binding
[[kv_namespaces]]
binding = "KV" # Available as env.KV
id = "YOUR_KV_NAMESPACE_ID" # Replace with actual KV namespace ID
`;
  }
  
  if (features.includes('r2')) {
    bindingsConfig += `
# R2 Bucket binding
[[r2_buckets]]
binding = "BUCKET" # Available as env.BUCKET
bucket_name = "${projectName}-bucket"
`;
  }
  
  if (features.includes('do')) {
    bindingsConfig += `
# Durable Objects binding
[durable_objects]
bindings = [
  { name = "COUNTER", class_name = "Counter", script_name = "" }
]
`;
  }
  
  if (features.includes('queues')) {
    bindingsConfig += `
# Queue binding
[[queues.producers]]
binding = "QUEUE" # Available as env.QUEUE
queue = "${projectName}-queue"
`;
  }
  
  // Always replace {{name}} placeholder
  wranglerContent = wranglerContent.replace(/\{\{name\}\}/g, projectName);
  
  if (features.length > 0) {
    wranglerContent += bindingsConfig;
  }
  
  fs.writeFileSync(path.join(projectPath, 'wrangler.toml'), wranglerContent);
  
  // Create .dev.vars.example file for local development secrets
  if (answers.useAuth) {
    const devVarsExample = `# Local development secrets (copy to .dev.vars)
# For production, use: wrangler secret put SECRET_NAME --env production

JWT_SECRET=your-jwt-secret-here
API_KEY=your-api-key-here
`;
    fs.writeFileSync(path.join(projectPath, '.dev.vars.example'), devVarsExample);
    
    // Add .dev.vars to .gitignore
    let gitignoreContent = fs.readFileSync(path.join(projectPath, '.gitignore'), 'utf8');
    if (!gitignoreContent.includes('.dev.vars')) {
      gitignoreContent += '\n# Local development secrets\n.dev.vars\n';
      fs.writeFileSync(path.join(projectPath, '.gitignore'), gitignoreContent);
    }
  }

  // Update README.md - replace placeholders
  let readmeContent = fs.readFileSync(path.join(projectPath, 'README.md'), 'utf8');
  readmeContent = readmeContent.replace(/\{\{name\}\}/g, projectName);
  readmeContent = readmeContent.replace(/\{\{description\}\}/g, answers.description || 'A Cloudflare Worker application');
  fs.writeFileSync(path.join(projectPath, 'README.md'), readmeContent);
  
  // Update src/lib/openapi.ts - replace placeholders
  const openApiPath = path.join(projectPath, 'src/lib/openapi.ts');
  if (fs.existsSync(openApiPath)) {
    let openApiContent = fs.readFileSync(openApiPath, 'utf8');
    openApiContent = openApiContent.replace(/\{\{name\}\}/g, projectName);
    openApiContent = openApiContent.replace(/\{\{description\}\}/g, answers.description || 'A Cloudflare Worker application');
    fs.writeFileSync(openApiPath, openApiContent);
  }

  // Update src/index.ts based on options
  let indexContent = fs.readFileSync(path.join(projectPath, 'src/index.ts'), 'utf8');
  
  if (!answers.useOpenAPI) {
    // Remove OpenAPI imports and routes
    indexContent = indexContent.replace(/import.*swagger.*\n/g, '');
    indexContent = indexContent.replace(/import.*openapi.*\n/g, '');
    indexContent = indexContent.replace(/\/\/ Swagger UI[\s\S]*?app\.get\('\/openapi\.json'[\s\S]*?\}\);/g, '');
  }

  fs.writeFileSync(path.join(projectPath, 'src/index.ts'), indexContent);

  // Detect package manager
  const packageManager = detectPackageManager();
  
  console.log(green('✅ Project created successfully!\n'));
  
  // Auto install dependencies if requested
  if (answers.installDeps) {
    console.log(blue('📦 Installing dependencies...\n'));
    try {
      execSync(`${packageManager} install`, {
        cwd: projectPath,
        stdio: 'inherit'
      });
      console.log(green('\n✅ Dependencies installed successfully!\n'));
    } catch (error) {
      console.log(yellow('\n⚠️  Failed to install dependencies. Please install manually.\n'));
    }
  }
  
  console.log('Next steps:\n');
  console.log(cyan(`  cd ${targetDir}`));
  if (!answers.installDeps) {
    console.log(cyan(`  ${packageManager} install`));
  }
  
  // Show service setup instructions
  if (features.length > 0) {
    console.log(cyan('\n⚙️  Configure Cloudflare services:'));
    
    if (features.includes('d1')) {
      console.log(yellow('\n  D1 Database:'));
      console.log('    1. Create database: wrangler d1 create <database-name>');
      console.log('    2. Update database_id in wrangler.toml');
      console.log('    3. Run migrations: wrangler d1 migrations apply <database-name>');
    }
    
    if (features.includes('kv')) {
      console.log(yellow('\n  KV Namespace:'));
      console.log('    1. Create namespace: wrangler kv:namespace create <namespace-name>');
      console.log('    2. Update id in wrangler.toml with the namespace ID');
    }
    
    if (features.includes('r2')) {
      console.log(yellow('\n  R2 Bucket:'));
      console.log('    1. Create bucket: wrangler r2 bucket create <bucket-name>');
      console.log('    2. Bucket name is already configured in wrangler.toml');
    }
  }
  
  if (answers.useAuth) {
    console.log(yellow('\n🔐 Authentication setup:'));
    console.log('    1. Copy .dev.vars.example to .dev.vars for local development');
    console.log('    2. For production: wrangler secret put JWT_SECRET --env production');
  }
  
  console.log(cyan(`  ${packageManager} run dev\n`));

  if (answers.useOpenAPI) {
    console.log(yellow('📚 API documentation will be available at http://localhost:8787/docs\n'));
  }
}

function copyDir(src, dest) {
  fs.mkdirSync(dest, { recursive: true });
  const entries = fs.readdirSync(src, { withFileTypes: true });

  for (const entry of entries) {
    const srcPath = path.join(src, entry.name);
    const destPath = path.join(dest, entry.name);

    if (entry.isDirectory()) {
      copyDir(srcPath, destPath);
    } else {
      fs.copyFileSync(srcPath, destPath);
    }
  }
}

init().catch((e) => {
  console.error(red('Error:'), e);
  process.exit(1);
});