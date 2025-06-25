# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.1] - 2025-06-25

### Changed
- Documentation updated to English as primary language
- Added Chinese translation as README.zh-CN.md

## [1.1.0] - 2025-06-25

### Added
- 🚀 Interactive route generator with 4 templates (Basic/CRUD/Auth/Webhook)
- 📦 CRUD one-click generation for complete RESTful APIs
- 🧪 Jest testing framework integration with core functionality tests
- 🔄 GitHub Actions CI/CD pipeline for automated testing and npm publishing
- 🏷️ Project source identification (README badges, package.json metadata, .create-worker-app file)
- 📝 Comprehensive documentation with examples and comparisons
- 🌍 English README as primary documentation with Chinese translation

### Changed
- Route generator upgraded from simple CLI to interactive mode with prompts
- README completely rewritten with detailed features and usage examples
- Project structure enhanced with better organization

### Fixed
- TypeScript diagnostics issues in generate-route.js

## [1.0.0] - 2025-06-25

### Added
- Initial release of create-worker-app
- Fast project scaffolding for Cloudflare Workers with Hono.js
- TypeScript support with full type definitions
- Optional OpenAPI/Swagger documentation generation
- Optional database configuration setup
- Built-in route generator for quick API development
- Pre-configured Wrangler for easy deployment
- Production-ready middleware (CORS, logging, error handling)

[1.1.1]: https://github.com/leeguooooo/create-worker-app/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/leeguooooo/create-worker-app/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/leeguooooo/create-worker-app/releases/tag/v1.0.0