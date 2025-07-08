package templates

// Base templates that are common across all architectures

const GoMod = `module {{.Module}}

go 1.21

require (
	github.com/aws/aws-lambda-go v1.41.0
	github.com/aws/aws-sdk-go-v2 v1.24.0
	github.com/aws/aws-sdk-go-v2/config v1.26.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.26.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.47.0
	github.com/aws/aws-sdk-go-v2/service/sqs v1.28.0
	github.com/aws/smithy-go v1.19.0
	github.com/caarlos0/env/v10 v10.0.0
	github.com/google/uuid v1.5.0
	github.com/joho/godotenv v1.5.1
	github.com/rs/zerolog v1.31.0
	{{- if eq .TestingFramework "testify" }}
	github.com/stretchr/testify v1.8.4
	github.com/vektra/mockery/v2 v2.38.0
	{{- else if eq .TestingFramework "ginkgo" }}
	github.com/onsi/ginkgo/v2 v2.13.2
	github.com/onsi/gomega v1.30.0
	{{- end }}
	{{- if .HasFeature "api" }}
	github.com/gin-gonic/gin v1.9.1
	github.com/swaggo/swag v1.16.2
	github.com/swaggo/gin-swagger v1.6.0
	{{- end }}
	{{- if .HasFeature "cognito" }}
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.31.0
	{{- end }}
	{{- if .HasFeature "secrets" }}
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.25.0
	{{- end }}
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/trace v1.21.0
)
`

const Makefile = `.PHONY: build test clean deploy run-local generate-handler lint fmt

# Variables
BINARY_NAME={{.Name}}
LAMBDA_RUNTIME=provided.al2023
GOOS=linux
GOARCH=amd64
CGO_ENABLED=0

# Colors for output
GREEN=\033[0;32m
RED=\033[0;31m
YELLOW=\033[1;33m
NC=\033[0m # No Color

# Build all Lambda functions
build:
	@echo "$(GREEN)Building Lambda functions...$(NC)"
	@mkdir -p build
	{{- if eq .Architecture "clean" }}
	@for dir in cmd/*; do \
		if [ -d "$$dir" ]; then \
			func=$$(basename $$dir); \
			echo "Building $$func..."; \
			GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -tags lambda.norpc -o build/$$func/bootstrap $$dir/main.go; \
			cd build/$$func && zip -j ../$$func.zip bootstrap && cd ../..; \
		fi \
	done
	{{- else if eq .Architecture "simple" }}
	@for file in handlers/*.go; do \
		if [ -f "$$file" ]; then \
			func=$$(basename $$file .go); \
			echo "Building $$func..."; \
			GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -tags lambda.norpc -o build/$$func/bootstrap $$file; \
			cd build/$$func && zip -j ../$$func.zip bootstrap && cd ../..; \
		fi \
	done
	{{- else }}
	@for dir in cmd/*; do \
		if [ -d "$$dir" ]; then \
			func=$$(basename $$dir); \
			echo "Building $$func..."; \
			GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) go build -tags lambda.norpc -o build/$$func/bootstrap $$dir/main.go; \
			cd build/$$func && zip -j ../$$func.zip bootstrap && cd ../..; \
		fi \
	done
	{{- end }}
	@echo "$(GREEN)Build complete!$(NC)"

# Run tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	{{- if eq .TestingFramework "ginkgo" }}
	@ginkgo -r --cover --race
	{{- else }}
	@go test -v -cover -race ./...
	{{- end }}

# Run tests with coverage report
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf build/
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete!$(NC)"

# Lint code
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@golangci-lint run --fix
	@echo "$(GREEN)Linting complete!$(NC)"

# Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Formatting complete!$(NC)"

# Generate new handler
generate-handler:
	@echo "$(GREEN)Generating new handler...$(NC)"
	@go run scripts/generate-handler.go
	@echo "$(GREEN)Handler generated!$(NC)"

# Run locally with SAM
run-local:
	@echo "$(GREEN)Starting local development server...$(NC)"
	{{- if eq .DeploymentTool "sam" }}
	@sam local start-api --env-vars .env.local
	{{- else if eq .DeploymentTool "serverless" }}
	@serverless offline
	{{- else }}
	@echo "$(RED)Local development not configured for {{.DeploymentTool}}$(NC)"
	{{- end }}

# Deploy to development
deploy-dev:
	@echo "$(GREEN)Deploying to development...$(NC)"
	{{- if eq .DeploymentTool "sam" }}
	@sam deploy --config-env dev --parameter-overrides file://deployments/dev.yaml
	{{- else if eq .DeploymentTool "cdk" }}
	@cd cdk && npm run deploy:dev
	{{- else if eq .DeploymentTool "serverless" }}
	@serverless deploy --stage dev --config serverless.yml --param="file://deployments/dev.yml"
	{{- else if eq .DeploymentTool "terraform" }}
	@cd terraform && terraform apply -var-file="environments/dev.tfvars" -auto-approve
	{{- end }}

# Deploy to staging
deploy-staging:
	@echo "$(GREEN)Deploying to staging...$(NC)"
	{{- if eq .DeploymentTool "sam" }}
	@sam deploy --config-env staging --parameter-overrides file://deployments/staging.yaml
	{{- else if eq .DeploymentTool "cdk" }}
	@cd cdk && npm run deploy:staging
	{{- else if eq .DeploymentTool "serverless" }}
	@serverless deploy --stage staging --config serverless.yml --param="file://deployments/staging.yml"
	{{- else if eq .DeploymentTool "terraform" }}
	@cd terraform && terraform apply -var-file="environments/staging.tfvars" -auto-approve
	{{- end }}

# Deploy to production
deploy-prod:
	@echo "$(RED)Deploying to production...$(NC)"
	@echo "$(YELLOW)Are you sure? [y/N]$(NC)"
	@read -r response; \
	if [ "$$response" = "y" ] || [ "$$response" = "Y" ]; then \
		{{- if eq .DeploymentTool "sam" }}
		sam deploy --config-env prod --parameter-overrides file://deployments/prod.yaml; \
		{{- else if eq .DeploymentTool "cdk" }}
		cd cdk && npm run deploy:prod; \
		{{- else if eq .DeploymentTool "serverless" }}
		serverless deploy --stage prod --config serverless.yml --param="file://deployments/production.yml"; \
		{{- else if eq .DeploymentTool "terraform" }}
		cd terraform && terraform apply -var-file="environments/prod.tfvars"; \
		{{- end }}
		echo "$(GREEN)Production deployment complete!$(NC)"; \
	else \
		echo "$(YELLOW)Production deployment cancelled.$(NC)"; \
	fi

# Install dependencies
deps:
	@echo "$(GREEN)Installing dependencies...$(NC)"
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	{{- if eq .TestingFramework "ginkgo" }}
	@go install github.com/onsi/ginkgo/v2/ginkgo@latest
	{{- end }}
	{{- if eq .TestingFramework "testify" }}
	@go install github.com/vektra/mockery/v2@latest
	{{- end }}
	{{- if .HasFeature "api" }}
	@go install github.com/swaggo/swag/cmd/swag@latest
	{{- end }}
	@echo "$(GREEN)Dependencies installed!$(NC)"

# Generate mocks
generate-mocks:
	@echo "$(GREEN)Generating mocks...$(NC)"
	{{- if eq .TestingFramework "testify" }}
	@mockery --all --output=test/mocks
	{{- else }}
	@echo "$(YELLOW)Mock generation not configured for {{.TestingFramework}}$(NC)"
	{{- end }}

# Run security scan
security:
	@echo "$(GREEN)Running security scan...$(NC)"
	@gosec ./...
	@echo "$(GREEN)Security scan complete!$(NC)"

# Show help
help:
	@echo "$(GREEN){{.Name}} - Available commands:$(NC)"
	@echo "  make build           - Build all Lambda functions"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make lint            - Run linters"
	@echo "  make fmt             - Format code"
	@echo "  make generate-handler - Generate new Lambda handler"
	@echo "  make run-local       - Run locally with SAM/Serverless"
	@echo "  make deploy-dev      - Deploy to development"
	@echo "  make deploy-staging  - Deploy to staging"
	@echo "  make deploy-prod     - Deploy to production"
	@echo "  make deps            - Install dependencies"
	@echo "  make generate-mocks  - Generate test mocks"
	@echo "  make security        - Run security scan"
	@echo "  make help            - Show this help message"

# Default target
all: clean deps lint test build
`

const README = `# {{.Name}}

{{.Description}}

## 🚀 Features

- **Architecture**: {{.Architecture}} architecture pattern
- **Deployment**: {{.DeploymentTool}} for infrastructure management
- **Testing**: {{.TestingFramework}} for comprehensive testing
{{- if .HasFeature "api" }}
- **API Gateway**: RESTful API with OpenAPI documentation
{{- end }}
{{- if .HasFeature "dynamodb" }}
- **DynamoDB**: NoSQL database integration
{{- end }}
{{- if .HasFeature "sqs" }}
- **SQS**: Message queue processing
{{- end }}
{{- if .HasFeature "sns" }}
- **SNS**: Event notification system
{{- end }}
{{- if .HasFeature "s3" }}
- **S3**: Object storage integration
{{- end }}
{{- if .HasFeature "cognito" }}
- **Cognito**: User authentication and authorization
{{- end }}
{{- if .HasFeature "eventbridge" }}
- **EventBridge**: Event-driven architecture
{{- end }}

## 📋 Prerequisites

- Go 1.21 or higher
- AWS CLI configured with appropriate credentials
- {{.DeploymentTool}} installed
{{- if eq .DeploymentTool "sam" }}
- SAM CLI (https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)
{{- else if eq .DeploymentTool "cdk" }}
- Node.js 18+ and npm
- AWS CDK CLI: ` + "`npm install -g aws-cdk`" + `
{{- else if eq .DeploymentTool "serverless" }}
- Node.js 18+ and npm
- Serverless Framework: ` + "`npm install -g serverless`" + `
{{- else if eq .DeploymentTool "terraform" }}
- Terraform 1.5+ (https://www.terraform.io/downloads)
{{- end }}

## 🛠️ Installation

1. Clone the repository:
   ` + "```bash" + `
   git clone <repository-url>
   cd {{.Name}}
   ` + "```" + `

2. Install dependencies:
   ` + "```bash" + `
   make deps
   ` + "```" + `

3. Set up environment variables:
   ` + "```bash" + `
   cp .env.example .env.local
   # Edit .env.local with your configuration
   ` + "```" + `

## 🏗️ Project Structure

` + "```" + `
{{.Name}}/
{{- if eq .Architecture "clean" }}
├── cmd/                    # Lambda function entry points
├── internal/               # Private application code
│   ├── domain/            # Business logic and entities
│   ├── usecases/          # Application use cases
│   ├── interfaces/        # Interface adapters (Lambda, API)
│   └── infrastructure/    # External services (AWS, DB)
├── pkg/                   # Public packages
│   ├── logger/            # Structured logging
│   ├── errors/            # Custom error types
│   └── middleware/        # Shared middleware
{{- else if eq .Architecture "simple" }}
├── handlers/              # Lambda function handlers
├── models/                # Data models
├── services/              # Business logic
├── utils/                 # Utility functions
├── config/                # Configuration management
{{- else if eq .Architecture "ddd" }}
├── domain/                # Domain layer (entities, VOs, aggregates)
│   ├── aggregate/         # Aggregate roots
│   ├── entity/            # Domain entities
│   ├── valueobject/       # Value objects
│   ├── repository/        # Repository interfaces
│   └── event/             # Domain events
├── application/           # Application layer
│   ├── command/           # Command handlers
│   ├── query/             # Query handlers
│   └── handler/           # Application services
├── infrastructure/        # Infrastructure layer
│   ├── persistence/       # Data persistence
│   └── messaging/         # Message handling
├── interfaces/            # Interface layer
│   ├── lambda/            # Lambda handlers
│   └── api/               # API handlers
{{- end }}
├── test/                  # Test files and utilities
├── scripts/               # Build and deployment scripts
├── deployments/           # Environment-specific configs
└── docs/                  # Documentation
` + "```" + `

## 🚀 Development

### Running Locally

` + "```bash" + `
make run-local
` + "```" + `

This starts a local development server using {{.DeploymentTool}}.

### Generating New Handlers

` + "```bash" + `
make generate-handler
` + "```" + `

Follow the interactive prompts to create new Lambda handlers with boilerplate code.

### Running Tests

` + "```bash" + `
# Run all tests
make test

# Run tests with coverage
make test-coverage
` + "```" + `

### Code Quality

` + "```bash" + `
# Format code
make fmt

# Run linters
make lint

# Security scan
make security
` + "```" + `

## 📦 Building

Build all Lambda functions:

` + "```bash" + `
make build
` + "```" + `

This creates optimized binaries for the Lambda runtime in the ` + "`build/`" + ` directory.

## 🚢 Deployment

### Development Environment

` + "```bash" + `
make deploy-dev
` + "```" + `

### Staging Environment

` + "```bash" + `
make deploy-staging
` + "```" + `

### Production Environment

` + "```bash" + `
make deploy-prod
` + "```" + `

## 📊 Monitoring

- CloudWatch Logs: All Lambda functions automatically log to CloudWatch
- X-Ray Tracing: Distributed tracing is enabled for all functions
- Custom Metrics: Business metrics are sent to CloudWatch Metrics

## 🔐 Security

- All functions use IAM roles with least-privilege permissions
- Secrets are stored in AWS Secrets Manager
- Environment variables are encrypted at rest
- API endpoints are protected with API Gateway authorizers

## 📖 API Documentation

{{- if .HasFeature "api" }}
API documentation is available at:
- Local: http://localhost:3000/swagger
- Dev: https://dev-api.example.com/swagger
- Prod: https://api.example.com/swagger
{{- end }}

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (` + "`git checkout -b feature/AmazingFeature`" + `)
3. Commit your changes (` + "`git commit -m 'Add some AmazingFeature'`" + `)
4. Push to the branch (` + "`git push origin feature/AmazingFeature`" + `)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.
`

const GitIgnore = `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out
coverage.html

# Go workspace file
go.work

# Dependency directories
vendor/

# Build directories
build/
dist/
.aws-sam/

# Environment files
.env
.env.local
.env.*.local
*.env

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~
.DS_Store

# SAM/CDK/Serverless
{{- if eq .DeploymentTool "sam" }}
.aws-sam/
samconfig.toml
{{- else if eq .DeploymentTool "cdk" }}
cdk/node_modules/
cdk/cdk.out/
cdk/*.js
cdk/*.d.ts
!cdk/jest.config.js
{{- else if eq .DeploymentTool "serverless" }}
.serverless/
{{- else if eq .DeploymentTool "terraform" }}
terraform/.terraform/
terraform/*.tfstate
terraform/*.tfstate.*
terraform/.terraform.lock.hcl
{{- end }}

# Logs
logs/
*.log

# OS files
.DS_Store
Thumbs.db

# Temporary files
*.tmp
*.temp
`

const EnvExample = `# Application Configuration
APP_NAME={{.Name}}
APP_ENV=development
LOG_LEVEL=debug

# AWS Configuration
AWS_REGION=us-east-1
AWS_PROFILE=default

{{- if .HasFeature "dynamodb" }}
# DynamoDB Configuration
DYNAMODB_TABLE_PREFIX={{.Name}}_
DYNAMODB_ENDPOINT=http://localhost:8000
{{- end }}

{{- if .HasFeature "sqs" }}
# SQS Configuration
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789012/my-queue
SQS_DLQ_URL=https://sqs.us-east-1.amazonaws.com/123456789012/my-queue-dlq
{{- end }}

{{- if .HasFeature "s3" }}
# S3 Configuration
S3_BUCKET_NAME={{.Name}}-bucket
{{- end }}

{{- if .HasFeature "api" }}
# API Configuration
API_BASE_URL=http://localhost:3000
API_KEY=your-api-key-here
CORS_ORIGINS=http://localhost:3000,http://localhost:8080
{{- end }}

{{- if .HasFeature "cognito" }}
# Cognito Configuration
COGNITO_USER_POOL_ID=us-east-1_xxxxxxxxx
COGNITO_CLIENT_ID=xxxxxxxxxxxxxxxxxxxxxxxxx
{{- end }}

{{- if .HasFeature "secrets" }}
# Secrets Manager
SECRETS_PREFIX={{.Name}}/
{{- end }}

# Monitoring
ENABLE_XRAY=true
ENABLE_PROFILING=false
`

const DockerCompose = `version: '3.8'

services:
{{- if .HasFeature "dynamodb" }}
  dynamodb-local:
    image: amazon/dynamodb-local:latest
    container_name: {{.Name}}-dynamodb
    ports:
      - "8000:8000"
    command: "-jar DynamoDBLocal.jar -sharedDb -inMemory"
    environment:
      - AWS_ACCESS_KEY_ID=dummy
      - AWS_SECRET_ACCESS_KEY=dummy
      - AWS_REGION=us-east-1
{{- end }}

{{- if .HasFeature "sqs" }}
  localstack:
    image: localstack/localstack:latest
    container_name: {{.Name}}-localstack
    ports:
      - "4566:4566"
    environment:
      - SERVICES=sqs,sns,s3,secretsmanager
      - DEBUG=0
      - DATA_DIR=/tmp/localstack/data
    volumes:
      - "./scripts/localstack:/docker-entrypoint-initaws.d"
      - "localstack-data:/tmp/localstack"
{{- end }}

{{- if .HasFeature "api" }}
  swagger-ui:
    image: swaggerapi/swagger-ui:latest
    container_name: {{.Name}}-swagger
    ports:
      - "8080:8080"
    environment:
      - SWAGGER_JSON=/docs/openapi.yaml
    volumes:
      - "./docs:/docs"
{{- end }}

volumes:
  localstack-data:
`

const Dockerfile = `# Build stage
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy built binaries
COPY --from=builder /app/build ./build

# The specific handler will be specified at runtime
CMD ["./build/bootstrap"]
`