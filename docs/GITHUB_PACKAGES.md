# Using create-worker-app from GitHub Packages

This package is available from both npm and GitHub Packages.

## Installation from GitHub Packages

### 1. Setup Authentication

First, you need to authenticate with GitHub Packages. Create a personal access token (PAT) with `read:packages` scope:

1. Go to https://github.com/settings/tokens
2. Click "Generate new token" → "Generate new token (classic)"
3. Select scope: `read:packages`
4. Generate token and copy it

### 2. Configure npm

Create or edit `~/.npmrc`:

```bash
@leeguooooo:registry=https://npm.pkg.github.com
//npm.pkg.github.com/:_authToken=YOUR_GITHUB_TOKEN
```

Replace `YOUR_GITHUB_TOKEN` with your personal access token.

### 3. Install the Package

```bash
# Install from GitHub Packages
npx @leeguooooo/create-worker-app@latest my-app

# Or install globally
npm install -g @leeguooooo/create-worker-app
```

## Benefits of GitHub Packages

1. **Integrated with GitHub** - Package versions are linked to releases
2. **Private packages** - Can publish private packages for your team
3. **Same permissions** - Uses GitHub's permission model
4. **Package insights** - View download statistics and dependencies

## Switching Between Registries

```bash
# Use npm registry (default)
npx create-worker-app@latest my-app

# Use GitHub Packages
npx @leeguooooo/create-worker-app@latest my-app
```

## For Package Maintainers

The package is automatically published to both registries when a new release is created:
- npm: `create-worker-app`
- GitHub Packages: `@leeguooooo/create-worker-app`

Both packages contain identical content and functionality.