#!/usr/bin/env node

const fs = require('fs');
const { execSync } = require('child_process');
const path = require('path');

// Read current package.json
const packagePath = path.join(__dirname, '..', 'package.json');
const originalPackage = JSON.parse(fs.readFileSync(packagePath, 'utf8'));

// Create a modified version for GitHub Packages
const githubPackage = {
  ...originalPackage,
  name: '@leeguooooo/create-worker-app',
  publishConfig: {
    registry: 'https://npm.pkg.github.com'
  }
};

console.log('📦 Publishing to GitHub Packages...\n');

try {
  // Backup original package.json
  fs.writeFileSync(`${packagePath}.backup`, JSON.stringify(originalPackage, null, 2));
  
  // Write modified package.json
  fs.writeFileSync(packagePath, JSON.stringify(githubPackage, null, 2));
  
  // Set registry to GitHub Packages
  execSync('npm config set @leeguooooo:registry https://npm.pkg.github.com', { stdio: 'inherit' });
  
  // Publish to GitHub Packages
  console.log('Publishing @leeguooooo/create-worker-app to GitHub Packages...');
  execSync('npm publish', { 
    stdio: 'inherit',
    env: {
      ...process.env,
      npm_config_registry: 'https://npm.pkg.github.com'
    }
  });
  
  console.log('\n✅ Successfully published to GitHub Packages!');
  
} catch (error) {
  console.error('\n❌ Failed to publish:', error.message);
  process.exit(1);
} finally {
  // Restore original package.json
  if (fs.existsSync(`${packagePath}.backup`)) {
    fs.renameSync(`${packagePath}.backup`, packagePath);
  }
  
  // Reset registry
  execSync('npm config delete @leeguooooo:registry', { stdio: 'inherit' });
}