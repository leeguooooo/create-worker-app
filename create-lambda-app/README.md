# Create Lambda App

🚀 A professional scaffolding tool for AWS Lambda functions in Go, inspired by create-worker-app but designed specifically for serverless architectures.

## Features

- **Multiple Architecture Patterns**:
  - Clean Architecture with use cases and domain separation
  - Simple handler-based structure for quick development
  - Domain-Driven Design (DDD) with aggregates and events

- **Deployment Options**:
  - AWS SAM (Serverless Application Model)
  - AWS CDK (Cloud Development Kit)
  - Serverless Framework
  - Terraform

- **Built-in Features**:
  - API Gateway integration with OpenAPI documentation
  - DynamoDB support with repository patterns
  - SQS message processing
  - SNS event publishing
  - S3 bucket operations
  - Cognito authentication
  - EventBridge integration
  - Secrets Manager

- **Developer Experience**:
  - Interactive CLI with beautiful prompts
  - Handler generator for common patterns (CRUD, Auth, Events)
  - Comprehensive testing setup
  - CI/CD with GitHub Actions
  - Multi-environment configuration
  - Hot reloading for local development

## Installation

```bash
go install github.com/leeguooooo/create-lambda-app@latest
```

After installation, the binary will be in `$GOPATH/bin` (usually `~/go/bin`). You can run it with:

```bash
# If ~/go/bin is in your PATH
create-lambda-app

# Or use the full path
~/go/bin/create-lambda-app
```

Or clone and build:

```bash
git clone https://github.com/leeguooooo/create-lambda-app.git
cd create-lambda-app
go build -o create-lambda-app
./create-lambda-app  # Run directly
```

## Quick Start

```bash
# Interactive mode
create-lambda-app

# With project name
create-lambda-app my-serverless-api

# With options
create-lambda-app my-api --deployment sam --features api,dynamodb,sqs
```

## Usage

### Creating a New Project

1. Run `create-lambda-app` and follow the interactive prompts:
   - Enter project name
   - Add description
   - Choose deployment tool (SAM/CDK/Serverless/Terraform)
   - Select features (API, DynamoDB, SQS, etc.)
   - Choose architecture pattern (Clean/Simple/DDD)
   - Select testing framework

2. Navigate to your project:
   ```bash
   cd my-project
   make deps
   make run-local
   ```

### Project Structure

#### Clean Architecture
```
my-project/
├── cmd/                    # Lambda function entry points
├── internal/
│   ├── domain/            # Business logic and entities
│   ├── usecases/          # Application use cases
│   ├── interfaces/        # Lambda/API handlers
│   └── infrastructure/    # AWS services, database
├── pkg/                   # Shared packages
├── test/                  # Tests and mocks
├── deployments/           # Environment configs
└── docs/                  # Documentation
```

#### Simple Architecture
```
my-project/
├── handlers/              # Lambda handlers
├── models/                # Data models
├── services/              # Business logic
├── utils/                 # Utilities
├── config/                # Configuration
└── test/                  # Tests
```

#### DDD Architecture
```
my-project/
├── domain/
│   ├── aggregate/         # Aggregate roots
│   ├── entity/            # Domain entities
│   ├── valueobject/       # Value objects
│   └── repository/        # Repository interfaces
├── application/
│   ├── command/           # Commands and handlers
│   └── query/             # Queries and handlers
├── infrastructure/        # External services
└── interfaces/            # Lambda/API handlers
```

### Generating New Handlers

After creating a project, use the built-in generator:

```bash
make generate-handler
```

Choose from templates:
- **API Handler**: RESTful endpoint with validation
- **SQS Handler**: Message queue processor
- **EventBridge Handler**: Event-driven handler
- **S3 Handler**: File storage events
- **DynamoDB Stream**: Table change processor
- **Scheduled Handler**: Cron/rate-based tasks

### Development

```bash
# Run locally
make run-local

# Run tests
make test
make test-coverage

# Lint code
make lint

# Build functions
make build
```

### Deployment

```bash
# Deploy to development
make deploy-dev

# Deploy to staging
make deploy-staging

# Deploy to production
make deploy-prod
```

## Configuration

### Environment Variables

Create `.env.local` from `.env.example`:

```env
APP_NAME=my-project
APP_ENV=development
LOG_LEVEL=debug
AWS_REGION=us-east-1

# Feature-specific configs
DYNAMODB_TABLE_PREFIX=myproject_
SQS_QUEUE_URL=https://sqs.region.amazonaws.com/account/queue
```

### Multi-Environment Setup

Each environment has its own configuration:
- `deployments/dev.yaml`
- `deployments/staging.yaml`
- `deployments/prod.yaml`

## Features

### API Gateway Integration

Automatically sets up:
- RESTful endpoints
- Request/response validation
- CORS configuration
- API key management
- OpenAPI documentation

### DynamoDB Support

Includes:
- Repository pattern implementation
- Global secondary indexes
- Optimistic locking
- Batch operations
- Stream processing

### Message Queue (SQS)

Features:
- Batch message processing
- Dead letter queue setup
- Error handling and retries
- Message attributes

### Authentication (Cognito)

Provides:
- User pool configuration
- JWT token validation
- Custom authorizers
- MFA support

## Architecture Patterns

### Clean Architecture

Best for:
- Large, complex applications
- Teams prioritizing maintainability
- Projects with changing requirements

Features:
- Clear separation of concerns
- Dependency inversion
- Testable business logic
- Framework independence

### Simple Architecture

Best for:
- Small to medium projects
- Rapid prototyping
- Learning serverless
- Straightforward APIs

Features:
- Minimal boilerplate
- Direct AWS SDK usage
- Quick to understand
- Easy to modify

### Domain-Driven Design

Best for:
- Complex business domains
- Event-driven systems
- Teams familiar with DDD
- Microservices architecture

Features:
- Aggregate roots
- Domain events
- CQRS pattern
- Event sourcing ready

## Testing

### Unit Tests

```go
func TestCreateUser(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockUserRepository()
    useCase := NewCreateUserUseCase(mockRepo)
    
    // Act
    user, err := useCase.Execute(ctx, input)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Integration Tests

```go
func TestAPIEndpoint(t *testing.T) {
    // Start local DynamoDB
    container := setupDynamoDBContainer(t)
    defer container.Terminate(ctx)
    
    // Test API
    response := callAPI("/users", "POST", payload)
    assert.Equal(t, 201, response.StatusCode)
}
```

## Deployment Tools

### AWS SAM

- Native AWS tool
- CloudFormation based
- Local testing support
- Built-in best practices

### AWS CDK

- Infrastructure as code
- TypeScript/Python support
- Higher-level constructs
- Multi-stack applications

### Serverless Framework

- Multi-cloud support
- Large plugin ecosystem
- Simple configuration
- Community driven

### Terraform

- Declarative syntax
- State management
- Multi-provider support
- Module reusability

## Best Practices

1. **Security**:
   - Use least privilege IAM roles
   - Enable encryption at rest
   - Implement API authentication
   - Store secrets in Secrets Manager

2. **Performance**:
   - Minimize Lambda package size
   - Use connection pooling
   - Implement caching strategies
   - Monitor cold starts

3. **Monitoring**:
   - Enable X-Ray tracing
   - Set up CloudWatch alarms
   - Track custom metrics
   - Use structured logging

4. **Cost Optimization**:
   - Right-size Lambda memory
   - Use DynamoDB on-demand
   - Implement lifecycle policies
   - Monitor usage with tags

## Comparison with create-worker-app

| Feature | create-lambda-app | create-worker-app |
|---------|------------------|-------------------|
| Runtime | Go (Lambda) | TypeScript (Workers) |
| Platform | AWS | Cloudflare |
| Architecture | Clean/Simple/DDD | Route-based |
| Database | DynamoDB | D1/KV |
| Deployment | SAM/CDK/Serverless | Wrangler |
| Testing | Go test/Testify | Vitest |

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Acknowledgments

Inspired by [create-worker-app](https://github.com/cloudflare/create-worker-app) and the excellent developer experience it provides for Cloudflare Workers.

## Support

- 📖 [Documentation](https://github.com/leeguooooo/create-lambda-app/wiki)
- 🐛 [Issue Tracker](https://github.com/leeguooooo/create-lambda-app/issues)
- 💬 [Discussions](https://github.com/leeguooooo/create-lambda-app/discussions)