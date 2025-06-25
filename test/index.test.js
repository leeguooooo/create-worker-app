const fs = require('fs');
const path = require('path');
const os = require('os');

describe('create-worker-app', () => {
  let tempDir;

  beforeEach(() => {
    // Create a temp directory for each test
    tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'create-worker-app-test-'));
  });

  afterEach(() => {
    // Clean up temp directory
    if (fs.existsSync(tempDir)) {
      fs.rmSync(tempDir, { recursive: true, force: true });
    }
  });

  test('project structure is complete', () => {
    const templateDir = path.join(__dirname, '..', 'template');
    
    // Check template files exist
    const requiredFiles = [
      'package.json',
      'tsconfig.json',
      'wrangler.toml',
      'README.md',
      '.create-worker-app',
      'src/index.ts',
      'src/routes/health.ts',
      'src/schemas/common.ts',
      'src/lib/openapi.ts',
      'src/types/env.ts',
      'scripts/generate-route.js'
    ];

    for (const file of requiredFiles) {
      const filePath = path.join(templateDir, file);
      expect(fs.existsSync(filePath)).toBe(true);
    }
  });

  test('template package.json is valid', () => {
    const packageJsonPath = path.join(__dirname, '..', 'template', 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    
    expect(packageJson.name).toBe('{{name}}');
    expect(packageJson.description).toBe('{{description}}');
    expect(packageJson.dependencies.hono).toBeDefined();
    expect(packageJson.dependencies.prompts).toBeDefined();
    expect(packageJson.devDependencies.typescript).toBeDefined();
    expect(packageJson.devDependencies.wrangler).toBeDefined();
  });

  test('main index.js exports are correct', () => {
    const mainPath = path.join(__dirname, '..', 'index.js');
    expect(fs.existsSync(mainPath)).toBe(true);
    
    // Check shebang
    const content = fs.readFileSync(mainPath, 'utf8');
    expect(content.startsWith('#!/usr/bin/env node')).toBe(true);
  });

  test('all template directories exist', () => {
    const templateDir = path.join(__dirname, '..', 'template');
    const dirs = ['src', 'src/routes', 'src/schemas', 'src/lib', 'src/types', 'scripts'];
    
    for (const dir of dirs) {
      const dirPath = path.join(templateDir, dir);
      expect(fs.existsSync(dirPath)).toBe(true);
      expect(fs.statSync(dirPath).isDirectory()).toBe(true);
    }
  });

  test('generate-route.js is executable', () => {
    const scriptPath = path.join(__dirname, '..', 'template', 'scripts', 'generate-route.js');
    expect(fs.existsSync(scriptPath)).toBe(true);
    
    // Check shebang
    const content = fs.readFileSync(scriptPath, 'utf8');
    expect(content.startsWith('#!/usr/bin/env node')).toBe(true);
  });

  test('.create-worker-app metadata file is valid', () => {
    const metadataPath = path.join(__dirname, '..', 'template', '.create-worker-app');
    const metadata = JSON.parse(fs.readFileSync(metadataPath, 'utf8'));
    
    expect(metadata.generator).toBe('create-worker-app');
    expect(metadata.repository).toBeDefined();
    expect(metadata.templates).toBeDefined();
    expect(metadata.commands).toBeDefined();
  });

  test('package.json bin field is correct', () => {
    const packageJsonPath = path.join(__dirname, '..', 'package.json');
    const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
    
    expect(packageJson.bin).toBeDefined();
    expect(packageJson.bin['create-worker-app']).toBe('./index.js');
  });
});