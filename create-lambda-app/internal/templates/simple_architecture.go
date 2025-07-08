package templates

// Simple Architecture specific templates

const SimpleHandler = `package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"{{.Module}}/config"
	"{{.Module}}/models"
	"{{.Module}}/services"
	"{{.Module}}/utils"
	"github.com/rs/zerolog/log"
)

// Handler handles the main Lambda function
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load configuration")
		return utils.ErrorResponse(http.StatusInternalServerError, "Internal server error"), nil
	}

	// Initialize service
	svc := services.NewService(cfg)

	// Route based on path and method
	switch {
	case request.Path == "/users" && request.HTTPMethod == http.MethodPost:
		return createUser(ctx, svc, request)
	case request.Path == "/users" && request.HTTPMethod == http.MethodGet:
		return listUsers(ctx, svc, request)
	case request.Path == "/users/{id}" && request.HTTPMethod == http.MethodGet:
		return getUser(ctx, svc, request)
	default:
		return utils.ErrorResponse(http.StatusNotFound, "Route not found"), nil
	}
}

func createUser(ctx context.Context, svc *services.Service, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var input models.CreateUserInput
	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		return utils.ErrorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Validate input
	if err := input.Validate(); err != nil {
		return utils.ValidationErrorResponse(err), nil
	}

	// Create user
	user, err := svc.CreateUser(ctx, input)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to create user")
		return utils.HandleServiceError(err), nil
	}

	return utils.SuccessResponse(http.StatusCreated, user), nil
}

func getUser(ctx context.Context, svc *services.Service, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := request.PathParameters["id"]
	if userID == "" {
		return utils.ErrorResponse(http.StatusBadRequest, "User ID is required"), nil
	}

	user, err := svc.GetUser(ctx, userID)
	if err != nil {
		return utils.HandleServiceError(err), nil
	}

	return utils.SuccessResponse(http.StatusOK, user), nil
}

func listUsers(ctx context.Context, svc *services.Service, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse query parameters
	params := utils.ParseListParams(request.QueryStringParameters)

	result, err := svc.ListUsers(ctx, params)
	if err != nil {
		return utils.HandleServiceError(err), nil
	}

	return utils.SuccessResponse(http.StatusOK, result), nil
}

func main() {
	lambda.Start(Handler)
}
`

const SimpleModels = `package models

import (
	"errors"
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    ` + "`json:\"id\" dynamodbav:\"id\"`" + `
	Email     string    ` + "`json:\"email\" dynamodbav:\"email\"`" + `
	Name      string    ` + "`json:\"name\" dynamodbav:\"name\"`" + `
	Status    string    ` + "`json:\"status\" dynamodbav:\"status\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\" dynamodbav:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\" dynamodbav:\"updated_at\"`" + `
	DeletedAt *time.Time ` + "`json:\"deleted_at,omitempty\" dynamodbav:\"deleted_at,omitempty\"`" + `
}

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Email string ` + "`json:\"email\"`" + `
	Name  string ` + "`json:\"name\"`" + `
}

// Validate validates the create user input
func (i *CreateUserInput) Validate() error {
	if i.Email == "" {
		return errors.New("email is required")
	}
	if i.Name == "" {
		return errors.New("name is required")
	}
	// Add more validation as needed
	return nil
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Name   *string ` + "`json:\"name,omitempty\"`" + `
	Status *string ` + "`json:\"status,omitempty\"`" + `
}

// ListUsersResult represents the result of listing users
type ListUsersResult struct {
	Users      []*User ` + "`json:\"users\"`" + `
	TotalCount int64   ` + "`json:\"total_count\"`" + `
	Page       int     ` + "`json:\"page\"`" + `
	PageSize   int     ` + "`json:\"page_size\"`" + `
}

// ListParams represents pagination parameters
type ListParams struct {
	Page     int
	PageSize int
	Status   string
}

// ServiceError represents a service-level error
type ServiceError struct {
	Code    string
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// Common errors
var (
	ErrUserNotFound      = &ServiceError{Code: "USER_NOT_FOUND", Message: "user not found"}
	ErrUserAlreadyExists = &ServiceError{Code: "USER_EXISTS", Message: "user already exists"}
	ErrInvalidInput      = &ServiceError{Code: "INVALID_INPUT", Message: "invalid input"}
)
`

const SimpleService = `package services

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"{{.Module}}/config"
	"{{.Module}}/models"
	"github.com/rs/zerolog/log"
)

// Service handles business logic
type Service struct {
	config   *config.Config
	dynamoDB *dynamodb.Client
}

// NewService creates a new service instance
func NewService(cfg *config.Config) *Service {
	// Initialize AWS clients
	awsConfig, _ := config.LoadAWSConfig(context.Background())
	
	return &Service{
		config:   cfg,
		dynamoDB: dynamodb.NewFromConfig(awsConfig),
	}
}

// CreateUser creates a new user
func (s *Service) CreateUser(ctx context.Context, input models.CreateUserInput) (*models.User, error) {
	// Check if user already exists
	existing, _ := s.getUserByEmail(ctx, input.Email)
	if existing != nil {
		return nil, models.ErrUserAlreadyExists
	}

	// Create new user
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     input.Email,
		Name:      input.Name,
		Status:    "active",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Save to DynamoDB
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = s.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.config.DynamoDBTableName),
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	log.Ctx(ctx).Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Msg("User created successfully")

	return user, nil
}

// GetUser retrieves a user by ID
func (s *Service) GetUser(ctx context.Context, userID string) (*models.User, error) {
	result, err := s.dynamoDB.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(s.config.DynamoDBTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if result.Item == nil {
		return nil, models.ErrUserNotFound
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// ListUsers lists users with pagination
func (s *Service) ListUsers(ctx context.Context, params models.ListParams) (*models.ListUsersResult, error) {
	// Build scan input
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(s.config.DynamoDBTableName),
		Limit:     aws.Int32(int32(params.PageSize)),
	}

	// Add filter for status if provided
	if params.Status != "" {
		scanInput.FilterExpression = aws.String("#status = :status")
		scanInput.ExpressionAttributeNames = map[string]string{
			"#status": "status",
		}
		scanInput.ExpressionAttributeValues = map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberS{Value: params.Status},
		}
	}

	// Execute scan
	result, err := s.dynamoDB.Scan(ctx, scanInput)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Unmarshal users
	var users []*models.User
	for _, item := range result.Items {
		var user models.User
		if err := attributevalue.UnmarshalMap(item, &user); err != nil {
			log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal user")
			continue
		}
		users = append(users, &user)
	}

	return &models.ListUsersResult{
		Users:      users,
		TotalCount: int64(result.Count),
		Page:       params.Page,
		PageSize:   params.PageSize,
	}, nil
}

// UpdateUser updates user information
func (s *Service) UpdateUser(ctx context.Context, userID string, input models.UpdateUserInput) (*models.User, error) {
	// Get existing user
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.Status != nil {
		user.Status = *input.Status
	}
	user.UpdatedAt = time.Now().UTC()

	// Save to DynamoDB
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = s.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.config.DynamoDBTableName),
		Item:      item,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user (soft delete)
func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	// Get existing user
	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	// Soft delete
	now := time.Now().UTC()
	user.DeletedAt = &now
	user.Status = "deleted"
	user.UpdatedAt = now

	// Save to DynamoDB
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = s.dynamoDB.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.config.DynamoDBTableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// getUserByEmail retrieves a user by email
func (s *Service) getUserByEmail(ctx context.Context, email string) (*models.User, error) {
	result, err := s.dynamoDB.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(s.config.DynamoDBTableName),
		IndexName:              aws.String("email-index"),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var user models.User
	if err := attributevalue.UnmarshalMap(result.Items[0], &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}
`

const SimpleUtils = `package utils

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"{{.Module}}/models"
)

// SuccessResponse creates a successful API response
func SuccessResponse(statusCode int, data interface{}) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"data":    data,
	})

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

// ErrorResponse creates an error API response
func ErrorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"message": message,
		},
	})

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

// ValidationErrorResponse creates a validation error response
func ValidationErrorResponse(err error) events.APIGatewayProxyResponse {
	return ErrorResponse(http.StatusBadRequest, err.Error())
}

// HandleServiceError converts service errors to HTTP responses
func HandleServiceError(err error) events.APIGatewayProxyResponse {
	if svcErr, ok := err.(*models.ServiceError); ok {
		switch svcErr.Code {
		case "USER_NOT_FOUND":
			return ErrorResponse(http.StatusNotFound, svcErr.Message)
		case "USER_EXISTS":
			return ErrorResponse(http.StatusConflict, svcErr.Message)
		case "INVALID_INPUT":
			return ErrorResponse(http.StatusBadRequest, svcErr.Message)
		default:
			return ErrorResponse(http.StatusInternalServerError, "Internal server error")
		}
	}

	return ErrorResponse(http.StatusInternalServerError, "Internal server error")
}

// ParseListParams parses pagination parameters from query string
func ParseListParams(params map[string]string) models.ListParams {
	page := 1
	pageSize := 20

	if p, err := strconv.Atoi(params["page"]); err == nil && p > 0 {
		page = p
	}

	if ps, err := strconv.Atoi(params["page_size"]); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	return models.ListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   params["status"],
	}
}

// GetEnv gets an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}
`

const SimpleConfig = `package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the application
type Config struct {
	// Application
	AppName     string ` + "`env:\"APP_NAME\" envDefault:\"{{.Name}}\"`" + `
	Environment string ` + "`env:\"APP_ENV\" envDefault:\"development\"`" + `
	LogLevel    string ` + "`env:\"LOG_LEVEL\" envDefault:\"info\"`" + `
	
	// AWS
	AWSRegion string ` + "`env:\"AWS_REGION\" envDefault:\"us-east-1\"`" + `
	
	{{- if .HasFeature "dynamodb" }}
	// DynamoDB
	DynamoDBTableName string ` + "`env:\"DYNAMODB_TABLE_NAME\"`" + `
	DynamoDBEndpoint  string ` + "`env:\"DYNAMODB_ENDPOINT\"`" + `
	{{- end }}
	
	{{- if .HasFeature "sqs" }}
	// SQS
	SQSQueueURL string ` + "`env:\"SQS_QUEUE_URL\"`" + `
	{{- end }}
	
	{{- if .HasFeature "s3" }}
	// S3
	S3BucketName string ` + "`env:\"S3_BUCKET_NAME\"`" + `
	{{- end }}
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")
	
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}
	
	// Configure logging
	if err := configureLogging(cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("failed to configure logging: %w", err)
	}
	
	log.Info().
		Str("app_name", cfg.AppName).
		Str("environment", cfg.Environment).
		Msg("Configuration loaded successfully")
	
	return cfg, nil
}

// LoadAWSConfig loads AWS SDK configuration
func LoadAWSConfig(ctx context.Context) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(os.Getenv("AWS_REGION")),
	}

	// Add custom endpoint for local development
	if endpoint := os.Getenv("AWS_ENDPOINT_URL"); endpoint != "" {
		opts = append(opts, config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: endpoint,
				}, nil
			}),
		))
	}

	return config.LoadDefaultConfig(ctx, opts...)
}

// configureLogging configures the global logger
func configureLogging(level string) error {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	
	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	
	zerolog.SetGlobalLevel(logLevel)
	
	// Use console writer for development
	if level == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.Kitchen,
		})
	}
	
	return nil
}
`

// Simple architecture feature templates
const SimpleAPIHandler = `package handlers

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gin-gonic/gin"
	"{{.Module}}/config"
	"{{.Module}}/services"
	"github.com/rs/zerolog/log"
)

// APIHandler handles API Gateway requests using Gin
func APIHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load configuration")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
	}

	// Initialize service
	svc := services.NewService(cfg)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())

	// Setup routes
	setupRoutes(router, svc)

	// Convert API Gateway request to Gin
	return ginLambda.ProxyWithContext(ctx, request, router)
}

func setupRoutes(router *gin.Engine, svc *services.Service) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	// User routes
	users := router.Group("/users")
	{
		users.POST("", createUserHandler(svc))
		users.GET("", listUsersHandler(svc))
		users.GET("/:id", getUserHandler(svc))
		users.PUT("/:id", updateUserHandler(svc))
		users.DELETE("/:id", deleteUserHandler(svc))
	}
}

func createUserHandler(svc *services.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CreateUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		user, err := svc.CreateUser(c.Request.Context(), input)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(201, user)
	}
}

func getUserHandler(svc *services.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		
		user, err := svc.GetUser(c.Request.Context(), userID)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(200, user)
	}
}

func listUsersHandler(svc *services.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		params := parseListParams(c)
		
		result, err := svc.ListUsers(c.Request.Context(), params)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(200, result)
	}
}

func updateUserHandler(svc *services.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		
		var input models.UpdateUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		user, err := svc.UpdateUser(c.Request.Context(), userID, input)
		if err != nil {
			handleError(c, err)
			return
		}

		c.JSON(200, user)
	}
}

func deleteUserHandler(svc *services.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("id")
		
		if err := svc.DeleteUser(c.Request.Context(), userID); err != nil {
			handleError(c, err)
			return
		}

		c.Status(204)
	}
}

func handleError(c *gin.Context, err error) {
	if svcErr, ok := err.(*models.ServiceError); ok {
		switch svcErr.Code {
		case "USER_NOT_FOUND":
			c.JSON(404, gin.H{"error": svcErr.Message})
		case "USER_EXISTS":
			c.JSON(409, gin.H{"error": svcErr.Message})
		case "INVALID_INPUT":
			c.JSON(400, gin.H{"error": svcErr.Message})
		default:
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	log.Error().Err(err).Msg("Unhandled error")
	c.JSON(500, gin.H{"error": "Internal server error"})
}

func parseListParams(c *gin.Context) models.ListParams {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	return models.ListParams{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
	}
}

func main() {
	lambda.Start(APIHandler)
}
`

const SimpleAPIModels = `package models

// API-specific models

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        ` + "`json:\"success\"`" + `
	Data    interface{} ` + "`json:\"data,omitempty\"`" + `
	Error   *APIError   ` + "`json:\"error,omitempty\"`" + `
	Meta    *APIMeta    ` + "`json:\"meta,omitempty\"`" + `
}

// APIError represents an API error
type APIError struct {
	Type    string                 ` + "`json:\"type\"`" + `
	Message string                 ` + "`json:\"message\"`" + `
	Details map[string]interface{} ` + "`json:\"details,omitempty\"`" + `
}

// APIMeta represents API response metadata
type APIMeta struct {
	RequestID string ` + "`json:\"request_id\"`" + `
	Timestamp string ` + "`json:\"timestamp\"`" + `
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	TotalCount int ` + "`json:\"total_count\"`" + `
	Page       int ` + "`json:\"page\"`" + `
	PageSize   int ` + "`json:\"page_size\"`" + `
	TotalPages int ` + "`json:\"total_pages\"`" + `
}
`

const SimpleAPIUtils = `package utils

import (
	"time"

	"github.com/google/uuid"
	"{{.Module}}/models"
)

// BuildAPIResponse builds a standard API response
func BuildAPIResponse(success bool, data interface{}, err *models.APIError) models.APIResponse {
	return models.APIResponse{
		Success: success,
		Data:    data,
		Error:   err,
		Meta: &models.APIMeta{
			RequestID: uuid.New().String(),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// BuildPaginationMeta builds pagination metadata
func BuildPaginationMeta(totalCount int64, page, pageSize int) models.PaginationMeta {
	totalPages := int(totalCount) / pageSize
	if int(totalCount)%pageSize > 0 {
		totalPages++
	}

	return models.PaginationMeta{
		TotalCount: int(totalCount),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
`

const SimpleDynamoDBService = `package services

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"{{.Module}}/config"
	"github.com/rs/zerolog/log"
)

// DynamoDBService handles DynamoDB operations
type DynamoDBService struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBService creates a new DynamoDB service
func NewDynamoDBService(cfg *config.Config) (*DynamoDBService, error) {
	awsConfig, err := config.LoadAWSConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(awsConfig)

	// Custom endpoint for local development
	if cfg.DynamoDBEndpoint != "" {
		client = dynamodb.NewFromConfig(awsConfig, func(o *dynamodb.Options) {
			o.EndpointResolver = dynamodb.EndpointResolverFromURL(cfg.DynamoDBEndpoint)
		})
	}

	return &DynamoDBService{
		client:    client,
		tableName: cfg.DynamoDBTableName,
	}, nil
}

// PutItem saves an item to DynamoDB
func (s *DynamoDBService) PutItem(ctx context.Context, item interface{}) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      av,
	})
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

// GetItem retrieves an item from DynamoDB
func (s *DynamoDBService) GetItem(ctx context.Context, key map[string]types.AttributeValue, result interface{}) error {
	resp, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key:       key,
	})
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	if resp.Item == nil {
		return nil // Item not found
	}

	if err := attributevalue.UnmarshalMap(resp.Item, result); err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return nil
}

// Query performs a query operation
func (s *DynamoDBService) Query(ctx context.Context, input *dynamodb.QueryInput, result interface{}) error {
	input.TableName = &s.tableName
	
	resp, err := s.client.Query(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to query: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(resp.Items, result); err != nil {
		return fmt.Errorf("failed to unmarshal query results: %w", err)
	}

	return nil
}

// Scan performs a scan operation
func (s *DynamoDBService) Scan(ctx context.Context, input *dynamodb.ScanInput, result interface{}) error {
	input.TableName = &s.tableName
	
	resp, err := s.client.Scan(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to scan: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(resp.Items, result); err != nil {
		return fmt.Errorf("failed to unmarshal scan results: %w", err)
	}

	return nil
}

// DeleteItem deletes an item from DynamoDB
func (s *DynamoDBService) DeleteItem(ctx context.Context, key map[string]types.AttributeValue) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &s.tableName,
		Key:       key,
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

// BatchWriteItems writes multiple items
func (s *DynamoDBService) BatchWriteItems(ctx context.Context, items []types.WriteRequest) error {
	_, err := s.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			s.tableName: items,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to batch write items: %w", err)
	}

	return nil
}
`

const SimpleDynamoDBModels = `package models

import (
	"time"
)

// DynamoDBModel is the base model for DynamoDB items
type DynamoDBModel struct {
	PK        string    ` + "`dynamodbav:\"pk\"`" + `
	SK        string    ` + "`dynamodbav:\"sk\"`" + `
	Type      string    ` + "`dynamodbav:\"type\"`" + `
	CreatedAt time.Time ` + "`dynamodbav:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`dynamodbav:\"updated_at\"`" + `
	TTL       *int64    ` + "`dynamodbav:\"ttl,omitempty\"`" + `
}

// UserDynamoDBModel represents a user in DynamoDB
type UserDynamoDBModel struct {
	DynamoDBModel
	ID     string  ` + "`dynamodbav:\"id\"`" + `
	Email  string  ` + "`dynamodbav:\"email\"`" + `
	Name   string  ` + "`dynamodbav:\"name\"`" + `
	Status string  ` + "`dynamodbav:\"status\"`" + `
}

// NewUserDynamoDBModel creates a new user DynamoDB model
func NewUserDynamoDBModel(user *User) *UserDynamoDBModel {
	return &UserDynamoDBModel{
		DynamoDBModel: DynamoDBModel{
			PK:        "USER#" + user.ID,
			SK:        "USER#" + user.ID,
			Type:      "User",
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		ID:     user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Status: user.Status,
	}
}

// ToUser converts DynamoDB model to User
func (m *UserDynamoDBModel) ToUser() *User {
	return &User{
		ID:        m.ID,
		Email:     m.Email,
		Name:      m.Name,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
`

const SimpleSQSHandler = `package handlers

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"{{.Module}}/config"
	"{{.Module}}/models"
	"{{.Module}}/services"
	"github.com/rs/zerolog/log"
)

// SQSHandler processes messages from SQS queue
func SQSHandler(ctx context.Context, sqsEvent events.SQSEvent) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msg("Failed to load configuration")
		return err
	}

	// Initialize service
	svc := services.NewSQSService(cfg)

	// Process each message
	for _, message := range sqsEvent.Records {
		if err := processMessage(ctx, svc, message); err != nil {
			log.Error().
				Err(err).
				Str("message_id", message.MessageId).
				Msg("Failed to process message")
			// Return error to retry the message
			return err
		}
	}

	return nil
}

func processMessage(ctx context.Context, svc *services.SQSService, message events.SQSMessage) error {
	log.Info().
		Str("message_id", message.MessageId).
		Str("body", message.Body).
		Msg("Processing SQS message")

	// Parse message body
	var msg models.QueueMessage
	if err := json.Unmarshal([]byte(message.Body), &msg); err != nil {
		log.Error().Err(err).Msg("Failed to parse message")
		// Don't retry invalid messages
		return nil
	}

	// Process based on message type
	switch msg.Type {
	case "user.created":
		return svc.HandleUserCreated(ctx, msg.Payload)
	case "user.updated":
		return svc.HandleUserUpdated(ctx, msg.Payload)
	case "email.send":
		return svc.HandleSendEmail(ctx, msg.Payload)
	default:
		log.Warn().
			Str("type", msg.Type).
			Msg("Unknown message type")
	}

	return nil
}

func main() {
	lambda.Start(SQSHandler)
}
`

const SimpleSQSService = `package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"{{.Module}}/config"
	"github.com/rs/zerolog/log"
)

// SQSService handles SQS operations
type SQSService struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSService creates a new SQS service
func NewSQSService(cfg *config.Config) *SQSService {
	awsConfig, _ := config.LoadAWSConfig(context.Background())
	
	return &SQSService{
		client:   sqs.NewFromConfig(awsConfig),
		queueURL: cfg.SQSQueueURL,
	}
}

// SendMessage sends a message to the queue
func (s *SQSService) SendMessage(ctx context.Context, messageType string, payload interface{}) error {
	body, err := json.Marshal(map[string]interface{}{
		"type":    messageType,
		"payload": payload,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.queueURL,
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	log.Info().
		Str("type", messageType).
		Msg("Message sent to SQS")

	return nil
}

// HandleUserCreated handles user created messages
func (s *SQSService) HandleUserCreated(ctx context.Context, payload json.RawMessage) error {
	var user map[string]interface{}
	if err := json.Unmarshal(payload, &user); err != nil {
		return fmt.Errorf("failed to unmarshal user: %w", err)
	}

	log.Info().
		Interface("user", user).
		Msg("Processing user created event")

	// Add your business logic here
	// For example: Send welcome email, create profile, etc.

	return nil
}

// HandleUserUpdated handles user updated messages
func (s *SQSService) HandleUserUpdated(ctx context.Context, payload json.RawMessage) error {
	var user map[string]interface{}
	if err := json.Unmarshal(payload, &user); err != nil {
		return fmt.Errorf("failed to unmarshal user: %w", err)
	}

	log.Info().
		Interface("user", user).
		Msg("Processing user updated event")

	// Add your business logic here

	return nil
}

// HandleSendEmail handles email sending messages
func (s *SQSService) HandleSendEmail(ctx context.Context, payload json.RawMessage) error {
	var email map[string]interface{}
	if err := json.Unmarshal(payload, &email); err != nil {
		return fmt.Errorf("failed to unmarshal email: %w", err)
	}

	log.Info().
		Interface("email", email).
		Msg("Processing send email event")

	// Add your email sending logic here
	// For example: Use SES to send email

	return nil
}
`

const SimpleSQSModels = `package models

import (
	"encoding/json"
	"time"
)

// QueueMessage represents a message in the queue
type QueueMessage struct {
	ID        string          ` + "`json:\"id\"`" + `
	Type      string          ` + "`json:\"type\"`" + `
	Payload   json.RawMessage ` + "`json:\"payload\"`" + `
	Timestamp time.Time       ` + "`json:\"timestamp\"`" + `
	Metadata  map[string]string ` + "`json:\"metadata,omitempty\"`" + `
}

// UserCreatedPayload represents the payload for user created events
type UserCreatedPayload struct {
	UserID string ` + "`json:\"user_id\"`" + `
	Email  string ` + "`json:\"email\"`" + `
	Name   string ` + "`json:\"name\"`" + `
}

// EmailPayload represents the payload for email messages
type EmailPayload struct {
	To      string   ` + "`json:\"to\"`" + `
	From    string   ` + "`json:\"from\"`" + `
	Subject string   ` + "`json:\"subject\"`" + `
	Body    string   ` + "`json:\"body\"`" + `
	CC      []string ` + "`json:\"cc,omitempty\"`" + `
	BCC     []string ` + "`json:\"bcc,omitempty\"`" + `
}
`