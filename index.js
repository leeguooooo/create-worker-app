#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const prompts = require('prompts');
const { red, green, yellow, blue, cyan } = require('kolorist');
const minimist = require('minimist');

const argv = minimist(process.argv.slice(2));

async function init() {
  console.log(cyan('\n🚀 Create Worker App\n'));

  let targetDir = argv._[0];
  let projectName = targetDir;

  if (!targetDir) {
    const response = await prompts({
      type: 'text',
      name: 'projectName',
      message: 'Project name:',
      initial: 'my-worker-app'
    });
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
      type: 'confirm',
      name: 'useDatabase',
      message: 'Will you need database configuration?',
      initial: false
    },
    {
      type: 'confirm',
      name: 'useOpenAPI',
      message: 'Include OpenAPI/Swagger documentation?',
      initial: true
    }
  ];

  const answers = await prompts(questions);

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

  // Add OpenAPI dependencies if selected
  if (answers.useOpenAPI) {
    packageJson.dependencies['@hono/zod-openapi'] = '^0.19.8';
    packageJson.dependencies['@hono/swagger-ui'] = '^0.5.2';
    packageJson.dependencies['zod'] = '^3.25.67';
  }

  fs.writeFileSync(
    path.join(projectPath, 'package.json'),
    JSON.stringify(packageJson, null, 2)
  );

  // Create .env.example if database is needed
  if (answers.useDatabase) {
    const envExample = `# Database configuration
DB_HOST=
DB_PORT=
DB_NAME=
DB_USER=
DB_PASSWORD=`;
    fs.writeFileSync(path.join(projectPath, '.env.example'), envExample);
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

  console.log(green('✅ Project created successfully!\n'));
  console.log('Next steps:\n');
  console.log(cyan(`  cd ${targetDir}`));
  console.log(cyan('  npm install'));
  console.log(cyan('  npm run dev\n'));

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