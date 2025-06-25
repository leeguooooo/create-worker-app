const { exec } = require('child_process');
const { promisify } = require('util');
const fs = require('fs').promises;
const path = require('path');
const os = require('os');

const execAsync = promisify(exec);

describe('create-worker-app', () => {
  let tempDir;

  beforeEach(async () => {
    // Create a temp directory for each test
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'create-worker-app-test-'));
  });

  afterEach(async () => {
    // Clean up temp directory
    await fs.rm(tempDir, { recursive: true, force: true });
  });

  test('creates a basic project with all required files', async () => {
    const projectName = 'test-app';
    const projectPath = path.join(tempDir, projectName);
    
    // Run the CLI
    const { stdout } = await execAsync(
      `node ${path.join(__dirname, '..', 'index.js')} ${projectName}`,
      { 
        cwd: tempDir,
        // Simulate interactive inputs
        input: 'Test description\nn\ny\n'
      }
    );

    // Check if project was created
    expect(stdout).toContain('Project created successfully');
    
    // Check project structure
    const files = [
      'package.json',
      'tsconfig.json',
      'wrangler.toml',
      'README.md',
      'src/index.ts',
      'src/routes/health.ts',
      'src/schemas/common.ts',
      'src/lib/openapi.ts',
      'src/types/env.ts',
      'scripts/generate-route.js'
    ];

    for (const file of files) {
      const filePath = path.join(projectPath, file);
      await expect(fs.access(filePath)).resolves.not.toThrow();
    }

    // Check package.json content
    const packageJson = JSON.parse(await fs.readFile(path.join(projectPath, 'package.json'), 'utf8'));
    expect(packageJson.name).toBe(projectName);
    expect(packageJson.description).toBe('Test description');
    expect(packageJson.dependencies['@hono/zod-openapi']).toBeDefined();
    expect(packageJson.dependencies['hono']).toBeDefined();
  }, 30000);

  test('creates a project without OpenAPI when declined', async () => {
    const projectName = 'test-app-no-openapi';
    const projectPath = path.join(tempDir, projectName);
    
    const { stdout } = await execAsync(
      `node ${path.join(__dirname, '..', 'index.js')} ${projectName}`,
      { 
        cwd: tempDir,
        input: 'Test description\nn\nn\n' // No database, no OpenAPI
      }
    );

    expect(stdout).toContain('Project created successfully');
    
    // Check that OpenAPI dependencies are not included
    const packageJson = JSON.parse(await fs.readFile(path.join(projectPath, 'package.json'), 'utf8'));
    expect(packageJson.dependencies['@hono/zod-openapi']).toBeUndefined();
    expect(packageJson.dependencies['@hono/swagger-ui']).toBeUndefined();
    expect(packageJson.dependencies['zod']).toBeUndefined();
    
    // Check that index.ts doesn't contain OpenAPI imports
    const indexContent = await fs.readFile(path.join(projectPath, 'src/index.ts'), 'utf8');
    expect(indexContent).not.toContain('swagger');
    expect(indexContent).not.toContain('openapi');
  }, 30000);

  test('creates .env.example when database option is selected', async () => {
    const projectName = 'test-app-with-db';
    const projectPath = path.join(tempDir, projectName);
    
    const { stdout } = await execAsync(
      `node ${path.join(__dirname, '..', 'index.js')} ${projectName}`,
      { 
        cwd: tempDir,
        input: 'Test description\ny\ny\n' // Yes to database
      }
    );

    expect(stdout).toContain('Project created successfully');
    
    // Check .env.example exists and has correct content
    const envPath = path.join(projectPath, '.env.example');
    await expect(fs.access(envPath)).resolves.not.toThrow();
    
    const envContent = await fs.readFile(envPath, 'utf8');
    expect(envContent).toContain('DB_HOST=');
    expect(envContent).toContain('DB_PASSWORD=');
  }, 30000);

  test('fails when directory already exists', async () => {
    const projectName = 'existing-dir';
    const projectPath = path.join(tempDir, projectName);
    
    // Create directory first
    await fs.mkdir(projectPath);
    
    // Try to create project in existing directory
    await expect(execAsync(
      `node ${path.join(__dirname, '..', 'index.js')} ${projectName}`,
      { cwd: tempDir }
    )).rejects.toThrow('already exists');
  });

  test('template files are copied correctly', async () => {
    const projectName = 'test-template-copy';
    const projectPath = path.join(tempDir, projectName);
    
    await execAsync(
      `node ${path.join(__dirname, '..', 'index.js')} ${projectName}`,
      { 
        cwd: tempDir,
        input: 'Test\nn\ny\n'
      }
    );

    // Check that all template directories are copied
    const dirs = ['src', 'src/routes', 'src/schemas', 'src/lib', 'src/types', 'scripts'];
    for (const dir of dirs) {
      const dirPath = path.join(projectPath, dir);
      const stat = await fs.stat(dirPath);
      expect(stat.isDirectory()).toBe(true);
    }
  }, 30000);
});