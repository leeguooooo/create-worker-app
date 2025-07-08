package templates

// Documentation templates

const ArchitectureDoc = `# Architecture

## Overview

This project follows the {{.Architecture}} architecture pattern for building scalable and maintainable serverless applications.

{{- if eq .Architecture "clean" }}
## Clean Architecture

The application is organized into concentric layers, with dependencies pointing inward:

### Layers

1. **Domain Layer** (innermost)
   - Entities: Core business objects
   - Repositories: Data access interfaces
   - Services: Domain logic

2. **Use Cases Layer**
   - Application-specific business rules
   - Orchestrates data flow between layers
   - Contains all use case implementations

3. **Interface Layer**
   - Lambda handlers
   - API controllers
   - Presentation logic

4. **Infrastructure Layer** (outermost)
   - Database implementations
   - AWS service integrations
   - External service clients

### Dependency Rule

Dependencies only point inward. Inner layers know nothing about outer layers.

` + "```" + `
┌─────────────────────────────────────────┐
│          Infrastructure Layer           │
│  (AWS Services, Database, External APIs) │
├─────────────────────────────────────────┤
│           Interface Layer               │
│    (Lambda Handlers, API Routes)        │
├─────────────────────────────────────────┤
│           Use Cases Layer               │
│     (Application Business Rules)        │
├─────────────────────────────────────────┤
│            Domain Layer                 │
│   (Entities, Business Rules, Ports)    │
└─────────────────────────────────────────┘
` + "```" + `
{{- else if eq .Architecture "simple" }}
## Simple Architecture

The application uses a straightforward, function-based architecture optimized for simplicity and quick development.

### Structure

- **handlers/**: Lambda function handlers
- **models/**: Data models and structures
- **services/**: Business logic and service integrations
- **utils/**: Shared utilities and helpers
- **config/**: Configuration management

### Flow

` + "```" + `
Request → Handler → Service → Model → Response
            ↓         ↓         ↓
          Utils    Config    External Services
` + "```" + `
{{- else if eq .Architecture "ddd" }}
## Domain-Driven Design (DDD)

The application follows DDD principles with clear bounded contexts and aggregate roots.

### Key Concepts

1. **Aggregates**: Cluster of domain objects treated as a single unit
2. **Entities**: Objects with unique identity
3. **Value Objects**: Immutable objects without identity
4. **Domain Events**: Capture important business occurrences
5. **Repositories**: Persistence abstractions

### Layer Structure

` + "```" + `
┌─────────────────────────────────────────┐
│        Interfaces Layer                 │
│    (Lambda Handlers, API Routes)        │
├─────────────────────────────────────────┤
│        Application Layer                │
│  (Commands, Queries, Event Handlers)    │
├─────────────────────────────────────────┤
│          Domain Layer                   │
│ (Aggregates, Entities, Value Objects)  │
├─────────────────────────────────────────┤
│       Infrastructure Layer              │
│  (Persistence, Messaging, External)     │
└─────────────────────────────────────────┘
` + "```" + `
{{- end }}

## Lambda Functions

### Function Types

{{- if .HasFeature "api" }}
#### API Functions
- Handle HTTP requests via API Gateway
- RESTful endpoints
- Request/response transformation
- Authentication and authorization
{{- end }}

{{- if .HasFeature "sqs" }}
#### Message Processors
- Process SQS queue messages
- Batch processing capabilities
- Dead letter queue handling
- Retry mechanisms
{{- end }}

{{- if .HasFeature "eventbridge" }}
#### Event Handlers
- React to EventBridge events
- Event pattern matching
- Asynchronous processing
- Event replay support
{{- end }}

### Function Configuration

Each Lambda function is configured with:
- Memory: 512MB (default, adjustable)
- Timeout: 30 seconds (API), 180 seconds (async)
- Environment variables
- IAM role with least privileges
- X-Ray tracing enabled
- CloudWatch Logs integration

## Data Flow

### Synchronous Flow (API)
` + "```" + `
Client → API Gateway → Lambda → Business Logic → Database → Response
` + "```" + `

### Asynchronous Flow (Events)
` + "```" + `
Event Source → Lambda → Business Logic → Database/Queue → Next Process
` + "```" + `

## Security

### Authentication & Authorization
{{- if .HasFeature "cognito" }}
- AWS Cognito for user management
- JWT token validation
- Role-based access control
{{- else }}
- API key authentication (configure as needed)
- Custom authorizer support
{{- end }}

### Data Protection
- Encryption at rest (DynamoDB, S3)
- Encryption in transit (TLS)
- Secrets Manager for sensitive data
- IAM roles with minimal permissions

## Scalability

### Auto-scaling
- Lambda functions scale automatically
- DynamoDB on-demand billing
- API Gateway handles load distribution

### Performance Optimization
- Connection pooling for databases
- Caching strategies (if applicable)
- Efficient serialization
- Minimal cold starts

## Monitoring & Observability

### CloudWatch Metrics
- Function invocations
- Error rates
- Duration metrics
- Custom business metrics

### X-Ray Tracing
- End-to-end request tracing
- Performance bottleneck identification
- Service map visualization

### Logging
- Structured JSON logging
- Correlation IDs
- Log aggregation in CloudWatch

## Error Handling

### Retry Strategies
- Exponential backoff for transient errors
- Dead letter queues for failed messages
- Circuit breaker pattern (where applicable)

### Error Types
- Business errors (4xx)
- System errors (5xx)
- Validation errors
- External service errors

## Testing Strategy

### Unit Tests
- Test individual functions/methods
- Mock external dependencies
- High code coverage target (>80%)

### Integration Tests
- Test component interactions
- Use test containers for databases
- Verify AWS service integrations

### End-to-End Tests
- Test complete workflows
- Use staging environment
- Automated test suites

## Development Workflow

1. **Local Development**
   - Use ` + "`make run-local`" + ` for local testing
   - Docker containers for dependencies
   - Hot reloading where possible

2. **Testing**
   - Write tests first (TDD encouraged)
   - Run ` + "`make test`" + ` before committing
   - Integration tests for critical paths

3. **Code Review**
   - Pull request workflow
   - Automated CI checks
   - Architecture compliance

4. **Deployment**
   - Automated via GitHub Actions
   - Environment promotion (dev → staging → prod)
   - Rollback capabilities

## Best Practices

1. **Code Organization**
   - Single responsibility principle
   - Clear module boundaries
   - Consistent naming conventions

2. **Error Handling**
   - Always handle errors explicitly
   - Use custom error types
   - Log errors with context

3. **Performance**
   - Minimize Lambda package size
   - Reuse connections
   - Optimize for cold starts

4. **Security**
   - Never hardcode secrets
   - Use least privilege IAM
   - Validate all inputs

5. **Monitoring**
   - Add custom metrics
   - Set up alerts
   - Monitor costs

## Further Reading

- [AWS Lambda Best Practices](https://docs.aws.amazon.com/lambda/latest/dg/best-practices.html)
- [Serverless Architecture Patterns](https://serverlessland.com/patterns)
{{- if eq .Architecture "clean" }}
- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
{{- else if eq .Architecture "ddd" }}
- [Domain-Driven Design by Eric Evans](https://www.domainlanguage.com/ddd/)
{{- end }}
`

const DeploymentDoc = `# Deployment Guide

## Overview

This guide covers deploying the {{.Name}} application using {{.DeploymentTool}}.

## Prerequisites

- AWS Account with appropriate permissions
- AWS CLI configured with credentials
- {{.DeploymentTool}} installed
{{- if eq .DeploymentTool "cdk" }}
- Node.js 18+ and npm
{{- else if eq .DeploymentTool "serverless" }}
- Node.js 18+ and npm
{{- else if eq .DeploymentTool "terraform" }}
- Terraform 1.5+
{{- end }}

## Environments

The application supports multiple environments:
- **Development** (dev): For active development and testing
- **Staging** (staging): Pre-production environment
- **Production** (prod): Live production environment

## Build Process

Before deployment, build all Lambda functions:

` + "```bash" + `
make build
` + "```" + `

This creates optimized binaries in the ` + "`build/`" + ` directory.

{{- if eq .DeploymentTool "sam" }}
## Deployment with AWS SAM

### Configuration

SAM configuration is stored in ` + "`samconfig.toml`" + ` with environment-specific settings.

### Deploy to Development

` + "```bash" + `
make deploy-dev
` + "```" + `

Or manually:
` + "```bash" + `
sam deploy --config-env dev --parameter-overrides file://deployments/dev.yaml
` + "```" + `

### Deploy to Staging

` + "```bash" + `
make deploy-staging
` + "```" + `

### Deploy to Production

` + "```bash" + `
make deploy-prod
` + "```" + `

### View Stack Outputs

` + "```bash" + `
aws cloudformation describe-stacks \
  --stack-name {{.Name}}-<env> \
  --query 'Stacks[0].Outputs'
` + "```" + `
{{- else if eq .DeploymentTool "cdk" }}
## Deployment with AWS CDK

### Setup

1. Install CDK dependencies:
` + "```bash" + `
cd cdk
npm install
` + "```" + `

2. Bootstrap CDK (first time only):
` + "```bash" + `
cdk bootstrap
` + "```" + `

### Deploy to Development

` + "```bash" + `
make deploy-dev
` + "```" + `

Or manually:
` + "```bash" + `
cd cdk
npm run deploy:dev
` + "```" + `

### Deploy to Staging

` + "```bash" + `
make deploy-staging
` + "```" + `

### Deploy to Production

` + "```bash" + `
make deploy-prod
` + "```" + `

### View Stack Outputs

` + "```bash" + `
cdk list
cdk deploy --outputs-file outputs.json
` + "```" + `
{{- else if eq .DeploymentTool "serverless" }}
## Deployment with Serverless Framework

### Configuration

Serverless configuration is in ` + "`serverless.yml`" + ` with stage-specific variables.

### Deploy to Development

` + "```bash" + `
make deploy-dev
` + "```" + `

Or manually:
` + "```bash" + `
serverless deploy --stage dev
` + "```" + `

### Deploy to Staging

` + "```bash" + `
make deploy-staging
` + "```" + `

### Deploy to Production

` + "```bash" + `
make deploy-prod
` + "```" + `

### View Deployment Info

` + "```bash" + `
serverless info --stage <env>
` + "```" + `
{{- else if eq .DeploymentTool "terraform" }}
## Deployment with Terraform

### Setup

1. Initialize Terraform:
` + "```bash" + `
cd terraform
terraform init
` + "```" + `

2. Create workspace for each environment:
` + "```bash" + `
terraform workspace new dev
terraform workspace new staging
terraform workspace new prod
` + "```" + `

### Deploy to Development

` + "```bash" + `
make deploy-dev
` + "```" + `

Or manually:
` + "```bash" + `
cd terraform
terraform workspace select dev
terraform apply -var-file=environments/dev.tfvars
` + "```" + `

### Deploy to Staging

` + "```bash" + `
make deploy-staging
` + "```" + `

### Deploy to Production

` + "```bash" + `
make deploy-prod
` + "```" + `

### View Outputs

` + "```bash" + `
terraform output
` + "```" + `
{{- end }}

## CI/CD Pipeline

### GitHub Actions

The project includes GitHub Actions workflows for automated deployment:

1. **CI Pipeline** (` + "`.github/workflows/ci.yml`" + `)
   - Runs on every push and PR
   - Executes tests and linting
   - Builds Lambda functions

2. **Deploy Pipeline** (` + "`.github/workflows/deploy.yml`" + `)
   - Deploys to dev on push to ` + "`develop`" + ` branch
   - Deploys to staging on push to ` + "`main`" + ` branch
   - Manual deployment to production

### Setting up GitHub Secrets

Add these secrets to your GitHub repository:

- ` + "`AWS_ACCESS_KEY_ID`" + `: AWS access key for dev/staging
- ` + "`AWS_SECRET_ACCESS_KEY`" + `: AWS secret key for dev/staging
- ` + "`PROD_AWS_ACCESS_KEY_ID`" + `: AWS access key for production
- ` + "`PROD_AWS_SECRET_ACCESS_KEY`" + `: AWS secret key for production

## Environment Variables

### Common Variables

All environments use these variables:
- ` + "`APP_NAME`" + `: Application name
- ` + "`APP_ENV`" + `: Environment (dev/staging/prod)
- ` + "`LOG_LEVEL`" + `: Logging level

### Environment-Specific Variables

Configure in deployment parameter files:
- ` + "`deployments/dev.yaml`" + `
- ` + "`deployments/staging.yaml`" + `
- ` + "`deployments/prod.yaml`" + `

## Post-Deployment

### Verification

1. Check Lambda functions:
` + "```bash" + `
aws lambda list-functions --query "Functions[?starts_with(FunctionName, '{{.Name}}')]"
` + "```" + `

2. Test API endpoints:
` + "```bash" + `
curl https://<api-gateway-url>/health
` + "```" + `

3. Monitor logs:
` + "```bash" + `
aws logs tail /aws/lambda/{{.Name}}-<function-name> --follow
` + "```" + `

### Monitoring

Set up CloudWatch dashboards and alarms:

1. Function errors
2. API Gateway 4xx/5xx errors
3. DynamoDB throttles
4. SQS queue depth

## Rollback

### Quick Rollback

{{- if eq .DeploymentTool "sam" }}
` + "```bash" + `
aws cloudformation cancel-update-stack --stack-name {{.Name}}-<env>
` + "```" + `
{{- else if eq .DeploymentTool "cdk" }}
` + "```bash" + `
cdk deploy --rollback
` + "```" + `
{{- else if eq .DeploymentTool "serverless" }}
` + "```bash" + `
serverless rollback --stage <env>
` + "```" + `
{{- else if eq .DeploymentTool "terraform" }}
` + "```bash" + `
terraform apply -var-file=environments/<env>.tfvars -refresh=true
` + "```" + `
{{- end }}

### Manual Rollback

1. Identify the previous working version
2. Check out the git tag/commit
3. Run the deployment process

## Troubleshooting

### Common Issues

1. **Deployment Fails**
   - Check AWS credentials
   - Verify IAM permissions
   - Review CloudFormation events

2. **Lambda Timeout**
   - Increase timeout in configuration
   - Check for infinite loops
   - Review CloudWatch logs

3. **Permission Denied**
   - Check IAM role policies
   - Verify resource permissions
   - Review execution role

### Debug Commands

View CloudFormation stack events:
` + "```bash" + `
aws cloudformation describe-stack-events \
  --stack-name {{.Name}}-<env> \
  --query 'StackEvents[0:10]'
` + "```" + `

View Lambda function configuration:
` + "```bash" + `
aws lambda get-function-configuration \
  --function-name {{.Name}}-<env>-<function>
` + "```" + `

## Security Considerations

1. **IAM Roles**
   - Use least privilege principle
   - Separate roles per function
   - Regular permission audits

2. **Secrets Management**
   - Use AWS Secrets Manager
   - Rotate secrets regularly
   - Never commit secrets

3. **Network Security**
   - Use VPC endpoints when needed
   - Configure security groups
   - Enable AWS WAF for APIs

## Cost Optimization

1. **Monitor Usage**
   - Set up billing alerts
   - Use AWS Cost Explorer
   - Tag all resources

2. **Optimize Functions**
   - Right-size memory allocation
   - Minimize package size
   - Use provisioned concurrency wisely

3. **Clean Up**
   - Remove unused resources
   - Delete old log groups
   - Archive old data

## Support

For deployment issues:
1. Check CloudWatch Logs
2. Review GitHub Actions logs
3. Consult AWS documentation
4. Open an issue in the repository
`

const APIDoc = `# API Documentation

{{- if .HasFeature "api" }}
## Overview

The {{.Name}} API provides RESTful endpoints for managing application resources.

### Base URL

- Development: ` + "`https://dev-api.{{.Name}}.com`" + `
- Staging: ` + "`https://staging-api.{{.Name}}.com`" + `
- Production: ` + "`https://api.{{.Name}}.com`" + `

### Authentication

{{- if .HasFeature "cognito" }}
The API uses JWT tokens issued by AWS Cognito. Include the token in the Authorization header:

` + "```" + `
Authorization: Bearer <token>
` + "```" + `
{{- else }}
Configure authentication based on your requirements. Options include:
- API Keys
- JWT tokens
- AWS IAM authentication
{{- end }}

### Common Headers

- ` + "`Content-Type: application/json`" + `
- ` + "`X-Request-ID`" + `: Unique request identifier for tracing

### Response Format

All responses follow this structure:

` + "```json" + `
{
  "success": true,
  "data": {
    // Response data
  },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
` + "```" + `

Error responses:

` + "```json" + `
{
  "success": false,
  "error": {
    "type": "VALIDATION_ERROR",
    "message": "Invalid input",
    "details": {
      // Error details
    }
  },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
` + "```" + `

## Endpoints

### Health Check

` + "```" + `
GET /health
` + "```" + `

Check API health status.

**Response:**
` + "```json" + `
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "timestamp": "2024-01-15T09:30:00Z"
  }
}
` + "```" + `

### Users

#### Create User

` + "```" + `
POST /users
` + "```" + `

Create a new user.

**Request Body:**
` + "```json" + `
{
  "email": "user@example.com",
  "name": "John Doe"
}
` + "```" + `

**Response:**
` + "```json" + `
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "active",
    "created_at": "2024-01-15T09:30:00Z",
    "updated_at": "2024-01-15T09:30:00Z"
  }
}
` + "```" + `

#### Get User

` + "```" + `
GET /users/{id}
` + "```" + `

Retrieve a user by ID.

**Path Parameters:**
- ` + "`id`" + `: User ID (UUID)

**Response:**
` + "```json" + `
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "status": "active",
    "created_at": "2024-01-15T09:30:00Z",
    "updated_at": "2024-01-15T09:30:00Z"
  }
}
` + "```" + `

#### List Users

` + "```" + `
GET /users
` + "```" + `

List all users with pagination.

**Query Parameters:**
- ` + "`page`" + `: Page number (default: 1)
- ` + "`page_size`" + `: Items per page (default: 20, max: 100)
- ` + "`status`" + `: Filter by status (active/inactive)

**Response:**
` + "```json" + `
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "name": "John Doe",
        "status": "active",
        "created_at": "2024-01-15T09:30:00Z",
        "updated_at": "2024-01-15T09:30:00Z"
      }
    ],
    "total_count": 100,
    "page": 1,
    "page_size": 20
  }
}
` + "```" + `

#### Update User

` + "```" + `
PUT /users/{id}
` + "```" + `

Update user information.

**Path Parameters:**
- ` + "`id`" + `: User ID (UUID)

**Request Body:**
` + "```json" + `
{
  "name": "Jane Doe",
  "status": "inactive"
}
` + "```" + `

**Response:**
` + "```json" + `
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "Jane Doe",
    "status": "inactive",
    "created_at": "2024-01-15T09:30:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
` + "```" + `

#### Delete User

` + "```" + `
DELETE /users/{id}
` + "```" + `

Delete a user (soft delete).

**Path Parameters:**
- ` + "`id`" + `: User ID (UUID)

**Response:**
` + "```" + `
204 No Content
` + "```" + `

## Error Codes

### HTTP Status Codes

- ` + "`200 OK`" + `: Successful request
- ` + "`201 Created`" + `: Resource created successfully
- ` + "`204 No Content`" + `: Successful request with no content
- ` + "`400 Bad Request`" + `: Invalid request data
- ` + "`401 Unauthorized`" + `: Authentication required
- ` + "`403 Forbidden`" + `: Insufficient permissions
- ` + "`404 Not Found`" + `: Resource not found
- ` + "`409 Conflict`" + `: Resource already exists
- ` + "`422 Unprocessable Entity`" + `: Validation error
- ` + "`429 Too Many Requests`" + `: Rate limit exceeded
- ` + "`500 Internal Server Error`" + `: Server error
- ` + "`503 Service Unavailable`" + `: Service temporarily unavailable

### Error Types

- ` + "`VALIDATION_ERROR`" + `: Input validation failed
- ` + "`NOT_FOUND`" + `: Resource not found
- ` + "`CONFLICT`" + `: Resource conflict (e.g., duplicate)
- ` + "`UNAUTHORIZED`" + `: Authentication failed
- ` + "`FORBIDDEN`" + `: Insufficient permissions
- ` + "`INTERNAL_ERROR`" + `: Internal server error
- ` + "`EXTERNAL_ERROR`" + `: External service error
- ` + "`TIMEOUT`" + `: Request timeout

## Rate Limiting

API rate limits:
- Development: 100 requests per minute
- Staging: 500 requests per minute
- Production: 1000 requests per minute

Rate limit headers:
- ` + "`X-RateLimit-Limit`" + `: Request limit
- ` + "`X-RateLimit-Remaining`" + `: Remaining requests
- ` + "`X-RateLimit-Reset`" + `: Reset timestamp

## Pagination

Paginated endpoints support these parameters:
- ` + "`page`" + `: Page number (starts at 1)
- ` + "`page_size`" + `: Items per page

Response includes:
- ` + "`total_count`" + `: Total number of items
- ` + "`page`" + `: Current page
- ` + "`page_size`" + `: Items per page

## Versioning

The API uses URL versioning. Current version: v1

Future versions will be available at:
- ` + "`/v2/users`" + `
- ` + "`/v3/users`" + `

## OpenAPI Specification

{{- if .HasFeature "api" }}
OpenAPI/Swagger documentation is available at:
- Development: ` + "`https://dev-api.{{.Name}}.com/swagger`" + `
- Staging: ` + "`https://staging-api.{{.Name}}.com/swagger`" + `
- Production: ` + "`https://api.{{.Name}}.com/swagger`" + `

Download OpenAPI spec:
` + "```bash" + `
curl https://api.{{.Name}}.com/openapi.yaml
` + "```" + `
{{- end }}

## SDK Examples

### JavaScript/TypeScript

` + "```javascript" + `
const api = new {{.Name}}API({
  baseURL: 'https://api.{{.Name}}.com',
  apiKey: 'your-api-key'
});

// Create user
const user = await api.users.create({
  email: 'user@example.com',
  name: 'John Doe'
});

// Get user
const user = await api.users.get('user-id');

// List users
const { users, totalCount } = await api.users.list({
  page: 1,
  pageSize: 20
});
` + "```" + `

### Go

` + "```go" + `
client := New{{.Name}}Client("https://api.{{.Name}}.com", "your-api-key")

// Create user
user, err := client.CreateUser(ctx, CreateUserInput{
    Email: "user@example.com",
    Name:  "John Doe",
})

// Get user
user, err := client.GetUser(ctx, "user-id")

// List users
result, err := client.ListUsers(ctx, ListUsersOptions{
    Page:     1,
    PageSize: 20,
})
` + "```" + `

### Python

` + "```python" + `
client = {{.Name}}Client(
    base_url="https://api.{{.Name}}.com",
    api_key="your-api-key"
)

# Create user
user = client.users.create(
    email="user@example.com",
    name="John Doe"
)

# Get user
user = client.users.get("user-id")

# List users
result = client.users.list(page=1, page_size=20)
` + "```" + `

## Webhooks

Configure webhooks to receive real-time notifications:

1. Register webhook endpoint
2. Verify webhook signature
3. Process webhook events

Webhook payload:
` + "```json" + `
{
  "event": "user.created",
  "timestamp": "2024-01-15T09:30:00Z",
  "data": {
    // Event data
  }
}
` + "```" + `

## Testing

### Test Environment

Use the development API for testing:
- Base URL: ` + "`https://dev-api.{{.Name}}.com`" + `
- Test credentials available in documentation

### Postman Collection

Import the Postman collection:
` + "```" + `
https://api.{{.Name}}.com/postman-collection.json
` + "```" + `

### cURL Examples

` + "```bash" + `
# Create user
curl -X POST https://api.{{.Name}}.com/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"email":"user@example.com","name":"John Doe"}'

# Get user
curl https://api.{{.Name}}.com/users/550e8400-e29b-41d4-a716-446655440000 \
  -H "Authorization: Bearer <token>"

# List users
curl "https://api.{{.Name}}.com/users?page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
` + "```" + `

## Support

For API support:
- Documentation: https://docs.{{.Name}}.com
- Status page: https://status.{{.Name}}.com
- Support email: api-support@{{.Name}}.com
{{- else }}
## API Not Configured

This project was created without API Gateway integration. To add API functionality:

1. Update your project configuration to include the API feature
2. Re-run the generator with API support
3. Or manually add API Gateway configuration to your deployment files

For more information, see the architecture documentation.
{{- end }}
`

const OpenAPISpec = `openapi: 3.0.3
info:
  title: {{.Name}} API
  description: {{.Description}}
  version: 1.0.0
  contact:
    name: API Support
    email: api-support@{{.Name}}.com
servers:
  - url: https://api.{{.Name}}.com
    description: Production
  - url: https://staging-api.{{.Name}}.com
    description: Staging
  - url: https://dev-api.{{.Name}}.com
    description: Development
{{- if .HasFeature "cognito" }}
security:
  - bearerAuth: []
{{- end }}
paths:
  /health:
    get:
      summary: Health check
      operationId: getHealth
      tags:
        - System
      security: []
      responses:
        '200':
          description: API is healthy
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'
  /users:
    get:
      summary: List users
      operationId: listUsers
      tags:
        - Users
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: page_size
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
        - name: status
          in: query
          schema:
            type: string
            enum: [active, inactive]
      responses:
        '200':
          description: User list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserListResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
    post:
      summary: Create user
      operationId: createUser
      tags:
        - Users
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateUserRequest'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '409':
          $ref: '#/components/responses/Conflict'
  /users/{id}:
    get:
      summary: Get user
      operationId: getUser
      tags:
        - Users
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: User details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
    put:
      summary: Update user
      operationId: updateUser
      tags:
        - Users
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateUserRequest'
      responses:
        '200':
          description: User updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/UserResponse'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
    delete:
      summary: Delete user
      operationId: deleteUser
      tags:
        - Users
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '204':
          description: User deleted
        '401':
          $ref: '#/components/responses/Unauthorized'
        '404':
          $ref: '#/components/responses/NotFound'
components:
  {{- if .HasFeature "cognito" }}
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  {{- end }}
  schemas:
    HealthResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            status:
              type: string
            version:
              type: string
            timestamp:
              type: string
              format: date-time
    User:
      type: object
      properties:
        id:
          type: string
          format: uuid
        email:
          type: string
          format: email
        name:
          type: string
        status:
          type: string
          enum: [active, inactive]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
    CreateUserRequest:
      type: object
      required:
        - email
        - name
      properties:
        email:
          type: string
          format: email
        name:
          type: string
          minLength: 2
          maxLength: 100
    UpdateUserRequest:
      type: object
      properties:
        name:
          type: string
          minLength: 2
          maxLength: 100
        status:
          type: string
          enum: [active, inactive]
    UserResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          $ref: '#/components/schemas/User'
        meta:
          $ref: '#/components/schemas/ResponseMeta'
    UserListResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
          properties:
            users:
              type: array
              items:
                $ref: '#/components/schemas/User'
            total_count:
              type: integer
            page:
              type: integer
            page_size:
              type: integer
        meta:
          $ref: '#/components/schemas/ResponseMeta'
    ErrorResponse:
      type: object
      properties:
        success:
          type: boolean
          example: false
        error:
          type: object
          properties:
            type:
              type: string
            message:
              type: string
            details:
              type: object
        meta:
          $ref: '#/components/schemas/ResponseMeta'
    ResponseMeta:
      type: object
      properties:
        request_id:
          type: string
          format: uuid
        timestamp:
          type: string
          format: date-time
  responses:
    BadRequest:
      description: Bad request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    NotFound:
      description: Not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    Conflict:
      description: Conflict
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    InternalError:
      description: Internal server error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
`