package templates

// Generator and utility templates

const HandlerGenerator = `package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

type HandlerConfig struct {
	Name           string
	Type           string
	Architecture   string
	HasDynamoDB    bool
	HasSQS         bool
	HasS3          bool
	HasEventBridge bool
}

func main() {
	fmt.Println(bold("🚀 Lambda Handler Generator"))
	fmt.Println()

	config := &HandlerConfig{}

	// Get handler name
	if err := survey.AskOne(&survey.Input{
		Message: "Handler name (e.g., user-service):",
		Help:    "The name will be used for the function and file names",
	}, &config.Name, survey.WithValidator(survey.Required)); err != nil {
		fmt.Println(red("Error:"), err)
		os.Exit(1)
	}

	// Normalize name
	config.Name = strings.ToLower(strings.ReplaceAll(config.Name, " ", "-"))

	// Get handler type
	if err := survey.AskOne(&survey.Select{
		Message: "Handler type:",
		Options: []string{
			"API (API Gateway triggered)",
			"SQS (Queue message processor)",
			"EventBridge (Event handler)",
			"S3 (Object storage events)",
			"DynamoDB Stream (Table stream processor)",
			"Scheduled (Cron/Rate based)",
			"Generic (Custom trigger)",
		},
		Default: "API",
	}, &config.Type); err != nil {
		fmt.Println(red("Error:"), err)
		os.Exit(1)
	}

	// Detect architecture from metadata
	metadata, _ := os.ReadFile(".create-lambda-app")
	config.Architecture = "clean" // default
	if strings.Contains(string(metadata), "simple") {
		config.Architecture = "simple"
	} else if strings.Contains(string(metadata), "ddd") {
		config.Architecture = "ddd"
	}

	// Generate handler based on type and architecture
	if err := generateHandler(config); err != nil {
		fmt.Println(red("Error generating handler:"), err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(green("✨ Handler generated successfully!"))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("1. Review the generated code in %s\n", getHandlerPath(config))
	fmt.Println("2. Update the deployment configuration to include the new function")
	fmt.Println("3. Run 'make build' to build the new handler")
	fmt.Println("4. Deploy with 'make deploy-dev'")
}

func generateHandler(config *HandlerConfig) error {
	var handlerTemplate string
	path := getHandlerPath(config)

	// Create directory
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Select template based on type and architecture
	switch config.Type {
	case "API (API Gateway triggered)":
		handlerTemplate = getAPIHandlerTemplate(config.Architecture)
	case "SQS (Queue message processor)":
		handlerTemplate = getSQSHandlerTemplate(config.Architecture)
	case "EventBridge (Event handler)":
		handlerTemplate = getEventBridgeHandlerTemplate(config.Architecture)
	case "S3 (Object storage events)":
		handlerTemplate = getS3HandlerTemplate(config.Architecture)
	case "DynamoDB Stream (Table stream processor)":
		handlerTemplate = getDynamoDBStreamHandlerTemplate(config.Architecture)
	case "Scheduled (Cron/Rate based)":
		handlerTemplate = getScheduledHandlerTemplate(config.Architecture)
	default:
		handlerTemplate = getGenericHandlerTemplate(config.Architecture)
	}

	// Parse and execute template
	tmpl, err := template.New("handler").Parse(handlerTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Create test file
	if err := generateTestFile(config); err != nil {
		return fmt.Errorf("failed to generate test file: %w", err)
	}

	// Update deployment configuration
	fmt.Println(yellow("Remember to update your deployment configuration to include the new function!"))

	return nil
}

func getHandlerPath(config *HandlerConfig) string {
	switch config.Architecture {
	case "simple":
		return fmt.Sprintf("handlers/%s.go", config.Name)
	case "ddd":
		return fmt.Sprintf("interfaces/lambda/%s/handler.go", config.Name)
	default: // clean
		return fmt.Sprintf("cmd/%s/main.go", config.Name)
	}
}

func generateTestFile(config *HandlerConfig) error {
	var testPath string
	
	switch config.Architecture {
	case "simple":
		testPath = fmt.Sprintf("handlers/%s_test.go", config.Name)
	case "ddd":
		testPath = fmt.Sprintf("interfaces/lambda/%s/handler_test.go", config.Name)
	default: // clean
		testPath = fmt.Sprintf("cmd/%s/main_test.go", config.Name)
	}

	// Create test template
	testTemplate := ` + "`" + `package {{if eq .Architecture "simple"}}handlers{{else if eq .Architecture "ddd"}}{{.Name}}{{else}}main{{end}}

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        events.APIGatewayProxyRequest
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful request",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/test",
			},
			expectedStatus: 200,
		},
		{
			name: "invalid request",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				Path:       "/test",
				Body:       "invalid json",
			},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create handler
			handler := NewHandler(nil) // Pass mock dependencies
			
			// Execute
			response, err := handler.HandleRequest(context.Background(), tt.request)
			
			// Assert
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, response.StatusCode)
			
			if tt.expectedBody != "" {
				assert.Contains(t, response.Body, tt.expectedBody)
			}
		})
	}
}
` + "`" + `

	// Parse and execute template
	tmpl, err := template.New("test").Parse(testTemplate)
	if err != nil {
		return err
	}

	// Create test file
	file, err := os.Create(testPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, config)
}

// Template functions for different handler types
func getAPIHandlerTemplate(architecture string) string {
	if architecture == "simple" {
		return simpleAPIHandlerTemplate
	}
	return cleanAPIHandlerTemplate
}

func getSQSHandlerTemplate(architecture string) string {
	if architecture == "simple" {
		return simpleSQSHandlerTemplate
	}
	return cleanSQSHandlerTemplate
}

func getEventBridgeHandlerTemplate(architecture string) string {
	return eventBridgeHandlerTemplate
}

func getS3HandlerTemplate(architecture string) string {
	return s3HandlerTemplate
}

func getDynamoDBStreamHandlerTemplate(architecture string) string {
	return dynamoDBStreamHandlerTemplate
}

func getScheduledHandlerTemplate(architecture string) string {
	return scheduledHandlerTemplate
}

func getGenericHandlerTemplate(architecture string) string {
	return genericHandlerTemplate
}

// Handler templates
const cleanAPIHandlerTemplate = ` + "`" + `package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	// Add your dependencies here
}

func NewHandler() *Handler {
	return &Handler{
		// Initialize dependencies
	}
}

func (h *Handler) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Ctx(ctx).Info().
		Str("method", request.HTTPMethod).
		Str("path", request.Path).
		Msg("Processing {{.Name}} request")

	// Add your handler logic here
	response := map[string]interface{}{
		"message": "Hello from {{.Name}}",
		"path":    request.Path,
		"method":  request.HTTPMethod,
	}

	body, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func main() {
	handler := NewHandler()
	lambda.Start(handler.HandleRequest)
}
` + "`" + `

const simpleAPIHandlerTemplate = ` + "`" + `package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

func {{.Name}}Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Ctx(ctx).Info().
		Str("method", request.HTTPMethod).
		Str("path", request.Path).
		Msg("Processing {{.Name}} request")

	// Add your handler logic here
	response := map[string]interface{}{
		"message": "Hello from {{.Name}}",
		"path":    request.Path,
		"method":  request.HTTPMethod,
	}

	body, _ := json.Marshal(response)

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func init() {
	lambda.Start({{.Name}}Handler)
}
` + "`" + `

const cleanSQSHandlerTemplate = ` + "`" + `package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	// Add your dependencies here
}

type Message struct {
	// Define your message structure
	ID      string ` + "`json:\"id\"`" + `
	Type    string ` + "`json:\"type\"`" + `
	Payload json.RawMessage ` + "`json:\"payload\"`" + `
}

func NewHandler() *Handler {
	return &Handler{
		// Initialize dependencies
	}
}

func (h *Handler) HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Ctx(ctx).Info().
		Int("message_count", len(sqsEvent.Records)).
		Msg("Processing SQS messages")

	for _, record := range sqsEvent.Records {
		if err := h.processMessage(ctx, record); err != nil {
			log.Ctx(ctx).Error().
				Err(err).
				Str("message_id", record.MessageId).
				Msg("Failed to process message")
			// Return error to retry the message
			return err
		}
	}

	return nil
}

func (h *Handler) processMessage(ctx context.Context, record events.SQSMessage) error {
	var msg Message
	if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	log.Ctx(ctx).Info().
		Str("message_id", msg.ID).
		Str("type", msg.Type).
		Msg("Processing message")

	// Add your message processing logic here
	switch msg.Type {
	case "user.created":
		// Handle user created event
	case "order.placed":
		// Handle order placed event
	default:
		log.Ctx(ctx).Warn().
			Str("type", msg.Type).
			Msg("Unknown message type")
	}

	return nil
}

func main() {
	handler := NewHandler()
	lambda.Start(handler.HandleRequest)
}
` + "`" + `

const simpleSQSHandlerTemplate = ` + "`" + `package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type Message struct {
	// Define your message structure
	ID      string ` + "`json:\"id\"`" + `
	Type    string ` + "`json:\"type\"`" + `
	Payload json.RawMessage ` + "`json:\"payload\"`" + `
}

func {{.Name}}Handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Ctx(ctx).Info().
		Int("message_count", len(sqsEvent.Records)).
		Msg("Processing SQS messages")

	for _, record := range sqsEvent.Records {
		var msg Message
		if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
			log.Ctx(ctx).Error().
				Err(err).
				Str("message_id", record.MessageId).
				Msg("Failed to unmarshal message")
			return err
		}

		// Process message
		log.Ctx(ctx).Info().
			Str("message_id", msg.ID).
			Str("type", msg.Type).
			Msg("Processing message")

		// Add your message processing logic here
	}

	return nil
}

func init() {
	lambda.Start({{.Name}}Handler)
}
` + "`" + `

const eventBridgeHandlerTemplate = ` + "`" + `package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	// Add your dependencies here
}

func NewHandler() *Handler {
	return &Handler{
		// Initialize dependencies
	}
}

func (h *Handler) HandleRequest(ctx context.Context, event events.CloudWatchEvent) error {
	log.Ctx(ctx).Info().
		Str("source", event.Source).
		Str("detail_type", event.DetailType).
		Str("id", event.ID).
		Msg("Processing EventBridge event")

	// Parse the detail
	var detail map[string]interface{}
	if err := json.Unmarshal(event.Detail, &detail); err != nil {
		return fmt.Errorf("failed to unmarshal event detail: %w", err)
	}

	// Handle different event types
	switch event.DetailType {
	case "UserCreated":
		return h.handleUserCreated(ctx, detail)
	case "OrderPlaced":
		return h.handleOrderPlaced(ctx, detail)
	default:
		log.Ctx(ctx).Warn().
			Str("detail_type", event.DetailType).
			Msg("Unknown event type")
	}

	return nil
}

func (h *Handler) handleUserCreated(ctx context.Context, detail map[string]interface{}) error {
	// Add your user created logic here
	log.Ctx(ctx).Info().
		Interface("detail", detail).
		Msg("Handling user created event")
	return nil
}

func (h *Handler) handleOrderPlaced(ctx context.Context, detail map[string]interface{}) error {
	// Add your order placed logic here
	log.Ctx(ctx).Info().
		Interface("detail", detail).
		Msg("Handling order placed event")
	return nil
}

func main() {
	handler := NewHandler()
	lambda.Start(handler.HandleRequest)
}
` + "`" + `

const s3HandlerTemplate = ` + "`" + `package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	s3Client *s3.Client
}

func NewHandler() (*Handler, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Handler{
		s3Client: s3.NewFromConfig(cfg),
	}, nil
}

func (h *Handler) HandleRequest(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		log.Ctx(ctx).Info().
			Str("bucket", record.S3.Bucket.Name).
			Str("key", record.S3.Object.Key).
			Str("event", record.EventName).
			Int64("size", record.S3.Object.Size).
			Msg("Processing S3 event")

		if err := h.processObject(ctx, record); err != nil {
			log.Ctx(ctx).Error().
				Err(err).
				Str("bucket", record.S3.Bucket.Name).
				Str("key", record.S3.Object.Key).
				Msg("Failed to process object")
			return err
		}
	}

	return nil
}

func (h *Handler) processObject(ctx context.Context, record events.S3EventRecord) error {
	// Handle different event types
	switch record.EventName {
	case "s3:ObjectCreated:Put", "s3:ObjectCreated:Post":
		return h.handleObjectCreated(ctx, record)
	case "s3:ObjectRemoved:Delete":
		return h.handleObjectDeleted(ctx, record)
	default:
		log.Ctx(ctx).Info().
			Str("event", record.EventName).
			Msg("Unhandled event type")
	}

	return nil
}

func (h *Handler) handleObjectCreated(ctx context.Context, record events.S3EventRecord) error {
	// Add your object created logic here
	// Example: Download object, process it, store results
	
	log.Ctx(ctx).Info().
		Str("bucket", record.S3.Bucket.Name).
		Str("key", record.S3.Object.Key).
		Msg("Processing new object")

	return nil
}

func (h *Handler) handleObjectDeleted(ctx context.Context, record events.S3EventRecord) error {
	// Add your object deleted logic here
	// Example: Clean up related data
	
	log.Ctx(ctx).Info().
		Str("bucket", record.S3.Bucket.Name).
		Str("key", record.S3.Object.Key).
		Msg("Processing deleted object")

	return nil
}

func main() {
	handler, err := NewHandler()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create handler")
	}

	lambda.Start(handler.HandleRequest)
}
` + "`" + `

const dynamoDBStreamHandlerTemplate = ` + "`" + `package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	// Add your dependencies here
}

func NewHandler() *Handler {
	return &Handler{
		// Initialize dependencies
	}
}

func (h *Handler) HandleRequest(ctx context.Context, event events.DynamoDBEvent) error {
	log.Ctx(ctx).Info().
		Int("record_count", len(event.Records)).
		Msg("Processing DynamoDB stream")

	for _, record := range event.Records {
		log.Ctx(ctx).Info().
			Str("event_name", record.EventName).
			Str("event_id", record.EventID).
			Msg("Processing stream record")

		switch record.EventName {
		case "INSERT":
			if err := h.handleInsert(ctx, record); err != nil {
				return err
			}
		case "MODIFY":
			if err := h.handleModify(ctx, record); err != nil {
				return err
			}
		case "REMOVE":
			if err := h.handleRemove(ctx, record); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Handler) handleInsert(ctx context.Context, record events.DynamoDBEventRecord) error {
	// Access new image
	newImage := record.Change.NewImage
	
	log.Ctx(ctx).Info().
		Interface("new_image", newImage).
		Msg("Handling INSERT event")

	// Add your insert logic here
	// Example: Send notification, update search index, etc.

	return nil
}

func (h *Handler) handleModify(ctx context.Context, record events.DynamoDBEventRecord) error {
	// Access both old and new images
	oldImage := record.Change.OldImage
	newImage := record.Change.NewImage
	
	log.Ctx(ctx).Info().
		Interface("old_image", oldImage).
		Interface("new_image", newImage).
		Msg("Handling MODIFY event")

	// Add your modify logic here
	// Example: Compare changes, send updates, etc.

	return nil
}

func (h *Handler) handleRemove(ctx context.Context, record events.DynamoDBEventRecord) error {
	// Access old image
	oldImage := record.Change.OldImage
	
	log.Ctx(ctx).Info().
		Interface("old_image", oldImage).
		Msg("Handling REMOVE event")

	// Add your remove logic here
	// Example: Clean up related data, send notifications, etc.

	return nil
}

func main() {
	handler := NewHandler()
	lambda.Start(handler.HandleRequest)
}
` + "`" + `

const scheduledHandlerTemplate = ` + "`" + `package main

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	// Add your dependencies here
}

func NewHandler() *Handler {
	return &Handler{
		// Initialize dependencies
	}
}

func (h *Handler) HandleRequest(ctx context.Context, event events.CloudWatchEvent) error {
	log.Ctx(ctx).Info().
		Str("id", event.ID).
		Time("time", event.Time).
		Msg("Processing scheduled event")

	startTime := time.Now()

	// Add your scheduled job logic here
	if err := h.performScheduledTask(ctx); err != nil {
		log.Ctx(ctx).Error().
			Err(err).
			Msg("Failed to perform scheduled task")
		return err
	}

	duration := time.Since(startTime)
	log.Ctx(ctx).Info().
		Dur("duration", duration).
		Msg("Scheduled task completed")

	return nil
}

func (h *Handler) performScheduledTask(ctx context.Context) error {
	// Add your scheduled task logic here
	// Examples:
	// - Generate reports
	// - Clean up old data
	// - Send digest emails
	// - Sync data between systems
	
	log.Ctx(ctx).Info().Msg("Performing scheduled task")

	// Simulate some work
	time.Sleep(100 * time.Millisecond)

	return nil
}

func main() {
	handler := NewHandler()
	lambda.Start(handler.HandleRequest)
}
` + "`" + `

const genericHandlerTemplate = ` + "`" + `package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	// Add your dependencies here
}

// Define your input and output types
type Input struct {
	// Add your input fields
	Message string ` + "`json:\"message\"`" + `
}

type Output struct {
	// Add your output fields
	Success bool   ` + "`json:\"success\"`" + `
	Message string ` + "`json:\"message\"`" + `
	Result  interface{} ` + "`json:\"result,omitempty\"`" + `
}

func NewHandler() *Handler {
	return &Handler{
		// Initialize dependencies
	}
}

func (h *Handler) HandleRequest(ctx context.Context, input Input) (Output, error) {
	log.Ctx(ctx).Info().
		Str("message", input.Message).
		Msg("Processing request")

	// Add your handler logic here
	
	return Output{
		Success: true,
		Message: "Request processed successfully",
		Result:  map[string]string{
			"input": input.Message,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}, nil
}

func main() {
	handler := NewHandler()
	lambda.Start(handler.HandleRequest)
}
` + "`" + `
`

const LocalSetupScript = `#!/bin/bash

set -e

echo "🚀 Setting up local development environment for {{.Name}}"
echo

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
echo "Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "Please install Go 1.21 or higher: https://golang.org/dl/"
    exit 1
else
    echo -e "${GREEN}✓ Go $(go version | awk '{print $3}')${NC}"
fi

# Check AWS CLI
if ! command -v aws &> /dev/null; then
    echo -e "${YELLOW}⚠️  AWS CLI is not installed${NC}"
    echo "Install from: https://aws.amazon.com/cli/"
else
    echo -e "${GREEN}✓ AWS CLI $(aws --version | awk '{print $1}')${NC}"
fi

{{- if eq .DeploymentTool "sam" }}
# Check SAM CLI
if ! command -v sam &> /dev/null; then
    echo -e "${RED}❌ SAM CLI is not installed${NC}"
    echo "Install from: https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html"
    exit 1
else
    echo -e "${GREEN}✓ SAM CLI $(sam --version | awk '{print $4}')${NC}"
fi
{{- else if eq .DeploymentTool "serverless" }}
# Check Serverless Framework
if ! command -v serverless &> /dev/null; then
    echo -e "${RED}❌ Serverless Framework is not installed${NC}"
    echo "Install with: npm install -g serverless"
    exit 1
else
    echo -e "${GREEN}✓ Serverless Framework$(NC}"
fi
{{- else if eq .DeploymentTool "terraform" }}
# Check Terraform
if ! command -v terraform &> /dev/null; then
    echo -e "${RED}❌ Terraform is not installed${NC}"
    echo "Install from: https://www.terraform.io/downloads"
    exit 1
else
    echo -e "${GREEN}✓ Terraform $(terraform version -json | jq -r .terraform_version)${NC}"
fi
{{- end }}

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠️  Docker is not installed${NC}"
    echo "Docker is optional but recommended for local testing"
else
    echo -e "${GREEN}✓ Docker $(docker --version | awk '{print $3}' | sed 's/,//')${NC}"
fi

echo

# Install Go dependencies
echo "Installing Go dependencies..."
go mod download
echo -e "${GREEN}✓ Dependencies installed${NC}"

# Install development tools
echo
echo "Installing development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
{{- if eq .TestingFramework "ginkgo" }}
go install github.com/onsi/ginkgo/v2/ginkgo@latest
{{- end }}
{{- if eq .TestingFramework "testify" }}
go install github.com/vektra/mockery/v2@latest
{{- end }}
echo -e "${GREEN}✓ Development tools installed${NC}"

# Setup environment
echo
echo "Setting up environment..."
if [ ! -f .env.local ]; then
    cp .env.example .env.local
    echo -e "${GREEN}✓ Created .env.local from .env.example${NC}"
    echo -e "${YELLOW}⚠️  Please update .env.local with your configuration${NC}"
else
    echo -e "${GREEN}✓ .env.local already exists${NC}"
fi

{{- if .HasFeature "dynamodb" }}
# Start local DynamoDB
echo
echo "Starting local DynamoDB..."
if command -v docker &> /dev/null; then
    docker-compose up -d dynamodb-local
    echo -e "${GREEN}✓ Local DynamoDB started on port 8000${NC}"
    
    # Create tables
    echo "Creating DynamoDB tables..."
    # Add table creation commands here
else
    echo -e "${YELLOW}⚠️  Docker not available, skipping local DynamoDB setup${NC}"
fi
{{- end }}

{{- if .HasFeature "sqs" }}
# Start LocalStack for SQS
echo
echo "Starting LocalStack..."
if command -v docker &> /dev/null; then
    docker-compose up -d localstack
    echo -e "${GREEN}✓ LocalStack started on port 4566${NC}"
    
    # Create queues
    echo "Creating SQS queues..."
    aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name {{.Name}}-messages --region us-east-1
    aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name {{.Name}}-messages-dlq --region us-east-1
    echo -e "${GREEN}✓ SQS queues created${NC}"
else
    echo -e "${YELLOW}⚠️  Docker not available, skipping LocalStack setup${NC}"
fi
{{- end }}

# Run initial build
echo
echo "Running initial build..."
make build
echo -e "${GREEN}✓ Build successful${NC}"

# Run tests
echo
echo "Running tests..."
make test
echo -e "${GREEN}✓ Tests passed${NC}"

echo
echo -e "${GREEN}✨ Local setup complete!${NC}"
echo
echo "Next steps:"
echo "1. Update .env.local with your configuration"
echo "2. Run 'make run-local' to start the local development server"
echo "3. Run 'make generate-handler' to create new Lambda handlers"
echo "4. Run 'make help' to see all available commands"
echo
echo "Happy coding! 🎉"
`

const TestUtils = `package testutils

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/require"
)

// CreateAPIGatewayRequest creates a test API Gateway request
func CreateAPIGatewayRequest(method, path string, body interface{}) events.APIGatewayProxyRequest {
	var bodyStr string
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyStr = string(bodyBytes)
	}

	return events.APIGatewayProxyRequest{
		HTTPMethod: method,
		Path:       path,
		Body:       bodyStr,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		RequestContext: events.APIGatewayProxyRequestContext{
			RequestID: "test-request-id",
			Stage:     "test",
		},
	}
}

// CreateSQSEvent creates a test SQS event
func CreateSQSEvent(messages []interface{}) events.SQSEvent {
	var records []events.SQSMessage

	for i, msg := range messages {
		body, _ := json.Marshal(msg)
		records = append(records, events.SQSMessage{
			MessageId: fmt.Sprintf("test-message-%d", i),
			Body:      string(body),
		})
	}

	return events.SQSEvent{
		Records: records,
	}
}

// CreateDynamoDBEvent creates a test DynamoDB stream event
func CreateDynamoDBEvent(eventName string, oldImage, newImage map[string]events.DynamoDBAttributeValue) events.DynamoDBEvent {
	return events.DynamoDBEvent{
		Records: []events.DynamoDBEventRecord{
			{
				EventID:   "test-event-id",
				EventName: eventName,
				Change: events.DynamoDBStreamRecord{
					OldImage: oldImage,
					NewImage: newImage,
				},
			},
		},
	}
}

// CreateS3Event creates a test S3 event
func CreateS3Event(bucket, key, eventName string) events.S3Event {
	return events.S3Event{
		Records: []events.S3EventRecord{
			{
				EventName: eventName,
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: bucket,
					},
					Object: events.S3Object{
						Key: key,
					},
				},
			},
		},
	}
}

// AssertAPIResponse asserts API Gateway response
func AssertAPIResponse(t *testing.T, response events.APIGatewayProxyResponse, expectedStatus int, expectedBody interface{}) {
	require.Equal(t, expectedStatus, response.StatusCode)

	if expectedBody != nil {
		var actualBody interface{}
		err := json.Unmarshal([]byte(response.Body), &actualBody)
		require.NoError(t, err)
		require.Equal(t, expectedBody, actualBody)
	}
}

// TestContext creates a test context with common values
func TestContext() context.Context {
	ctx := context.Background()
	// Add common test context values here
	return ctx
}

// LoadFixture loads a JSON fixture file
func LoadFixture(t *testing.T, path string, v interface{}) {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	err = json.Unmarshal(data, v)
	require.NoError(t, err)
}
`