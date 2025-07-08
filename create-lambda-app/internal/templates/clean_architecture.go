package templates

// Clean Architecture specific templates

const CleanBaseEntity = `package entities

import (
	"time"

	"github.com/google/uuid"
)

// BaseEntity contains common fields for all entities
type BaseEntity struct {
	ID        string    ` + "`json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
	Version   int       ` + "`json:\"version\"`" + `
}

// NewBaseEntity creates a new base entity with generated ID
func NewBaseEntity() BaseEntity {
	now := time.Now().UTC()
	return BaseEntity{
		ID:        uuid.New().String(),
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}
}

// UpdateVersion increments the version and updates the timestamp
func (e *BaseEntity) UpdateVersion() {
	e.Version++
	e.UpdatedAt = time.Now().UTC()
}

// Example User entity
type User struct {
	BaseEntity
	Email     string    ` + "`json:\"email\"`" + `
	Name      string    ` + "`json:\"name\"`" + `
	Status    string    ` + "`json:\"status\"`" + `
	DeletedAt *time.Time ` + "`json:\"deleted_at,omitempty\"`" + `
}

// NewUser creates a new user entity
func NewUser(email, name string) *User {
	return &User{
		BaseEntity: NewBaseEntity(),
		Email:      email,
		Name:       name,
		Status:     "active",
	}
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == "active" && u.DeletedAt == nil
}

// SoftDelete marks the user as deleted
func (u *User) SoftDelete() {
	now := time.Now().UTC()
	u.DeletedAt = &now
	u.Status = "deleted"
	u.UpdateVersion()
}
`

const CleanRepositoryInterface = `package repositories

import (
	"context"
	
	"{{.Module}}/internal/domain/entities"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create saves a new user
	Create(ctx context.Context, user *entities.User) error
	
	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id string) (*entities.User, error)
	
	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	
	// Update saves changes to an existing user
	Update(ctx context.Context, user *entities.User) error
	
	// Delete removes a user
	Delete(ctx context.Context, id string) error
	
	// List retrieves users with pagination
	List(ctx context.Context, offset, limit int) ([]*entities.User, error)
	
	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
}

// TransactionManager handles database transactions
type TransactionManager interface {
	// WithTransaction executes a function within a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// RepositoryError represents a repository-level error
type RepositoryError struct {
	Code    string
	Message string
	Err     error
}

func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Common error codes
const (
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeDuplicate    = "DUPLICATE"
	ErrCodeInvalidInput = "INVALID_INPUT"
	ErrCodeInternal     = "INTERNAL"
)

// Common errors
var (
	ErrUserNotFound = &RepositoryError{
		Code:    ErrCodeNotFound,
		Message: "user not found",
	}
	
	ErrUserAlreadyExists = &RepositoryError{
		Code:    ErrCodeDuplicate,
		Message: "user already exists",
	}
)
`

const CleanUseCaseInterface = `package usecases

import (
	"context"
	
	"{{.Module}}/internal/domain/entities"
)

// CreateUserInput represents the input for creating a user
type CreateUserInput struct {
	Email string ` + "`json:\"email\" validate:\"required,email\"`" + `
	Name  string ` + "`json:\"name\" validate:\"required,min=2,max=100\"`" + `
}

// CreateUserOutput represents the output of creating a user
type CreateUserOutput struct {
	User *entities.User ` + "`json:\"user\"`" + `
}

// UserUseCase defines user-related use cases
type UserUseCase interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error)
	
	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, userID string) (*entities.User, error)
	
	// UpdateUser updates user information
	UpdateUser(ctx context.Context, userID string, input UpdateUserInput) (*entities.User, error)
	
	// DeleteUser deletes a user
	DeleteUser(ctx context.Context, userID string) error
	
	// ListUsers lists users with pagination
	ListUsers(ctx context.Context, page, pageSize int) (*ListUsersOutput, error)
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	Name   *string ` + "`json:\"name,omitempty\" validate:\"omitempty,min=2,max=100\"`" + `
	Status *string ` + "`json:\"status,omitempty\" validate:\"omitempty,oneof=active inactive\"`" + `
}

// ListUsersOutput represents paginated user list output
type ListUsersOutput struct {
	Users      []*entities.User ` + "`json:\"users\"`" + `
	TotalCount int64           ` + "`json:\"total_count\"`" + `
	Page       int             ` + "`json:\"page\"`" + `
	PageSize   int             ` + "`json:\"page_size\"`" + `
}

// UseCaseError represents a use case level error
type UseCaseError struct {
	Type    string
	Message string
	Details map[string]interface{}
}

func (e *UseCaseError) Error() string {
	return e.Message
}

// Error types
const (
	ErrTypeValidation   = "VALIDATION_ERROR"
	ErrTypeNotFound     = "NOT_FOUND"
	ErrTypeConflict     = "CONFLICT"
	ErrTypeUnauthorized = "UNAUTHORIZED"
	ErrTypeInternal     = "INTERNAL_ERROR"
)
`

const CleanLambdaHandler = `package lambda

import (
	"context"
	"encoding/json"
	"net/http"
	
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
	
	"{{.Module}}/internal/infrastructure/config"
	"{{.Module}}/internal/usecases"
	"{{.Module}}/pkg/errors"
	"{{.Module}}/pkg/middleware"
)

// Handler represents a Lambda handler with dependencies
type Handler struct {
	userUseCase usecases.UserUseCase
	config      *config.Config
}

// NewHandler creates a new Lambda handler
func NewHandler(userUseCase usecases.UserUseCase, cfg *config.Config) *Handler {
	return &Handler{
		userUseCase: userUseCase,
		config:      cfg,
	}
}

// HandleRequest processes the Lambda request
func (h *Handler) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Add request ID to context
	ctx = middleware.WithRequestID(ctx, request.RequestContext.RequestID)
	
	// Log request
	log.Ctx(ctx).Info().
		Str("method", request.HTTPMethod).
		Str("path", request.Path).
		Str("request_id", request.RequestContext.RequestID).
		Msg("Processing request")
	
	// Route based on path and method
	switch {
	case request.Path == "/users" && request.HTTPMethod == http.MethodPost:
		return h.createUser(ctx, request)
	case request.Path == "/users" && request.HTTPMethod == http.MethodGet:
		return h.listUsers(ctx, request)
	case request.Path == "/users/{id}" && request.HTTPMethod == http.MethodGet:
		return h.getUser(ctx, request)
	case request.Path == "/users/{id}" && request.HTTPMethod == http.MethodPut:
		return h.updateUser(ctx, request)
	case request.Path == "/users/{id}" && request.HTTPMethod == http.MethodDelete:
		return h.deleteUser(ctx, request)
	default:
		return errorResponse(http.StatusNotFound, "Route not found")
	}
}

// createUser handles user creation
func (h *Handler) createUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var input usecases.CreateUserInput
	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body")
	}
	
	output, err := h.userUseCase.CreateUser(ctx, input)
	if err != nil {
		return handleUseCaseError(err)
	}
	
	return successResponse(http.StatusCreated, output)
}

// getUser handles retrieving a user
func (h *Handler) getUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := request.PathParameters["id"]
	if userID == "" {
		return errorResponse(http.StatusBadRequest, "User ID is required")
	}
	
	user, err := h.userUseCase.GetUser(ctx, userID)
	if err != nil {
		return handleUseCaseError(err)
	}
	
	return successResponse(http.StatusOK, user)
}

// listUsers handles listing users with pagination
func (h *Handler) listUsers(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	page := 1
	pageSize := 20
	
	// Parse query parameters
	if pageParam := request.QueryStringParameters["page"]; pageParam != "" {
		// Parse page parameter
	}
	if pageSizeParam := request.QueryStringParameters["page_size"]; pageSizeParam != "" {
		// Parse page size parameter
	}
	
	output, err := h.userUseCase.ListUsers(ctx, page, pageSize)
	if err != nil {
		return handleUseCaseError(err)
	}
	
	return successResponse(http.StatusOK, output)
}

// updateUser handles updating a user
func (h *Handler) updateUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := request.PathParameters["id"]
	if userID == "" {
		return errorResponse(http.StatusBadRequest, "User ID is required")
	}
	
	var input usecases.UpdateUserInput
	if err := json.Unmarshal([]byte(request.Body), &input); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body")
	}
	
	user, err := h.userUseCase.UpdateUser(ctx, userID, input)
	if err != nil {
		return handleUseCaseError(err)
	}
	
	return successResponse(http.StatusOK, user)
}

// deleteUser handles deleting a user
func (h *Handler) deleteUser(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	userID := request.PathParameters["id"]
	if userID == "" {
		return errorResponse(http.StatusBadRequest, "User ID is required")
	}
	
	if err := h.userUseCase.DeleteUser(ctx, userID); err != nil {
		return handleUseCaseError(err)
	}
	
	return successResponse(http.StatusNoContent, nil)
}

// Helper functions

func successResponse(statusCode int, data interface{}) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"success": true,
		"data":    data,
	})
	
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
			"X-Request-ID": middleware.GetRequestID(context.Background()),
		},
		Body: string(body),
	}, nil
}

func errorResponse(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
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
			"X-Request-ID": middleware.GetRequestID(context.Background()),
		},
		Body: string(body),
	}, nil
}

func handleUseCaseError(err error) (events.APIGatewayProxyResponse, error) {
	if ucErr, ok := err.(*usecases.UseCaseError); ok {
		switch ucErr.Type {
		case usecases.ErrTypeValidation:
			return errorResponse(http.StatusBadRequest, ucErr.Message)
		case usecases.ErrTypeNotFound:
			return errorResponse(http.StatusNotFound, ucErr.Message)
		case usecases.ErrTypeConflict:
			return errorResponse(http.StatusConflict, ucErr.Message)
		case usecases.ErrTypeUnauthorized:
			return errorResponse(http.StatusUnauthorized, ucErr.Message)
		default:
			return errorResponse(http.StatusInternalServerError, "Internal server error")
		}
	}
	
	log.Error().Err(err).Msg("Unexpected error in handler")
	return errorResponse(http.StatusInternalServerError, "Internal server error")
}

// Start initializes and starts the Lambda function
func Start() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
	
	// Initialize dependencies
	// TODO: Initialize repositories, use cases, etc.
	
	// Create handler
	handler := NewHandler(nil, cfg) // Pass real dependencies
	
	// Start Lambda
	lambda.Start(handler.HandleRequest)
}
`

const CleanConfig = `package config

import (
	"fmt"
	"time"
	
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
	AWSRegion  string ` + "`env:\"AWS_REGION\" envDefault:\"us-east-1\"`" + `
	AWSProfile string ` + "`env:\"AWS_PROFILE\" envDefault:\"default\"`" + `
	
	{{- if .HasFeature "dynamodb" }}
	// DynamoDB
	DynamoDBTablePrefix string ` + "`env:\"DYNAMODB_TABLE_PREFIX\" envDefault:\"{{.Name}}_\"`" + `
	DynamoDBEndpoint    string ` + "`env:\"DYNAMODB_ENDPOINT\"`" + `
	{{- end }}
	
	{{- if .HasFeature "sqs" }}
	// SQS
	SQSQueueURL    string ` + "`env:\"SQS_QUEUE_URL\"`" + `
	SQSDLQueueURL  string ` + "`env:\"SQS_DLQ_URL\"`" + `
	SQSMaxRetries  int    ` + "`env:\"SQS_MAX_RETRIES\" envDefault:\"3\"`" + `
	{{- end }}
	
	{{- if .HasFeature "s3" }}
	// S3
	S3BucketName string ` + "`env:\"S3_BUCKET_NAME\"`" + `
	{{- end }}
	
	{{- if .HasFeature "api" }}
	// API
	APIBaseURL   string   ` + "`env:\"API_BASE_URL\" envDefault:\"http://localhost:3000\"`" + `
	APIKey       string   ` + "`env:\"API_KEY\"`" + `
	CORSOrigins  []string ` + "`env:\"CORS_ORIGINS\" envSeparator:\",\"`" + `
	{{- end }}
	
	{{- if .HasFeature "cognito" }}
	// Cognito
	CognitoUserPoolID string ` + "`env:\"COGNITO_USER_POOL_ID\"`" + `
	CognitoClientID   string ` + "`env:\"COGNITO_CLIENT_ID\"`" + `
	{{- end }}
	
	{{- if .HasFeature "secrets" }}
	// Secrets Manager
	SecretsPrefix string ` + "`env:\"SECRETS_PREFIX\" envDefault:\"{{.Name}}/\"`" + `
	{{- end }}
	
	// Monitoring
	EnableXRay      bool ` + "`env:\"ENABLE_XRAY\" envDefault:\"true\"`" + `
	EnableProfiling bool ` + "`env:\"ENABLE_PROFILING\" envDefault:\"false\"`" + `
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
	
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	log.Info().
		Str("app_name", cfg.AppName).
		Str("environment", cfg.Environment).
		Msg("Configuration loaded successfully")
	
	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	{{- if .HasFeature "dynamodb" }}
	if c.DynamoDBTablePrefix == "" {
		return fmt.Errorf("DYNAMODB_TABLE_PREFIX is required")
	}
	{{- end }}
	
	{{- if .HasFeature "sqs" }}
	if c.SQSQueueURL == "" {
		return fmt.Errorf("SQS_QUEUE_URL is required")
	}
	{{- end }}
	
	{{- if .HasFeature "s3" }}
	if c.S3BucketName == "" {
		return fmt.Errorf("S3_BUCKET_NAME is required")
	}
	{{- end }}
	
	{{- if .HasFeature "cognito" }}
	if c.CognitoUserPoolID == "" || c.CognitoClientID == "" {
		return fmt.Errorf("Cognito configuration is incomplete")
	}
	{{- end }}
	
	return nil
}

// IsDevelopment returns true if running in development environment
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsProduction returns true if running in production environment
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
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

const Logger = `package logger

import (
	"context"
	"os"
	
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	*zerolog.Logger
}

// New creates a new logger instance
func New() *Logger {
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "{{.Name}}").
		Logger()
	
	return &Logger{&logger}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	logger := l.With().Logger()
	
	// Add request ID if present
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With().Str("request_id", requestID).Logger()
	}
	
	// Add user ID if present
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger = logger.With().Str("user_id", userID).Logger()
	}
	
	return &Logger{&logger}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := l.With().Interface(key, value).Logger()
	return &Logger{&logger}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	logger := l.With().Fields(fields).Logger()
	return &Logger{&logger}
}

// WithError adds an error to the logger
func (l *Logger) WithError(err error) *Logger {
	logger := l.With().Err(err).Logger()
	return &Logger{&logger}
}

// Global logger instance
var defaultLogger = New()

// WithContext returns the global logger with context
func WithContext(ctx context.Context) *Logger {
	return defaultLogger.WithContext(ctx)
}

// SetGlobalLevel sets the global log level
func SetGlobalLevel(level string) error {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(lvl)
	return nil
}

// Structured logging helpers

// Info logs an info message
func Info(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	if len(fields) > 0 {
		logger = logger.WithFields(fields[0])
	}
	logger.Info().Msg(msg)
}

// Debug logs a debug message
func Debug(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	if len(fields) > 0 {
		logger = logger.WithFields(fields[0])
	}
	logger.Debug().Msg(msg)
}

// Warn logs a warning message
func Warn(ctx context.Context, msg string, fields ...map[string]interface{}) {
	logger := WithContext(ctx)
	if len(fields) > 0 {
		logger = logger.WithFields(fields[0])
	}
	logger.Warn().Msg(msg)
}

// Error logs an error message
func Error(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	logger := WithContext(ctx).WithError(err)
	if len(fields) > 0 {
		logger = logger.WithFields(fields[0])
	}
	logger.Error().Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	logger := WithContext(ctx).WithError(err)
	if len(fields) > 0 {
		logger = logger.WithFields(fields[0])
	}
	logger.Fatal().Msg(msg)
}
`

const CustomErrors = `package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// ErrorTypeValidation indicates a validation error
	ErrorTypeValidation ErrorType = "VALIDATION"
	
	// ErrorTypeNotFound indicates a resource was not found
	ErrorTypeNotFound ErrorType = "NOT_FOUND"
	
	// ErrorTypeConflict indicates a conflict (e.g., duplicate resource)
	ErrorTypeConflict ErrorType = "CONFLICT"
	
	// ErrorTypeUnauthorized indicates an authorization error
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED"
	
	// ErrorTypeForbidden indicates a forbidden access error
	ErrorTypeForbidden ErrorType = "FORBIDDEN"
	
	// ErrorTypeInternal indicates an internal server error
	ErrorTypeInternal ErrorType = "INTERNAL"
	
	// ErrorTypeExternal indicates an external service error
	ErrorTypeExternal ErrorType = "EXTERNAL"
	
	// ErrorTypeTimeout indicates a timeout error
	ErrorTypeTimeout ErrorType = "TIMEOUT"
)

// AppError represents an application error with additional context
type AppError struct {
	Type    ErrorType              ` + "`json:\"type\"`" + `
	Message string                 ` + "`json:\"message\"`" + `
	Details map[string]interface{} ` + "`json:\"details,omitempty\"`" + `
	Cause   error                  ` + "`json:\"-\"`" + `
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error is of a specific type
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Type == t.Type
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	e.Details = details
	return e
}

// WithCause wraps another error
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// Constructor functions for common errors

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
	}
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
	}
}

// NewInternalError creates an internal error
func NewInternalError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
	}
}

// NewExternalError creates an external service error
func NewExternalError(service, message string) *AppError {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: fmt.Sprintf("external service error (%s): %s", service, message),
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *AppError {
	return &AppError{
		Type:    ErrorTypeTimeout,
		Message: fmt.Sprintf("operation timed out: %s", operation),
	}
}

// Helper functions

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeValidation
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeNotFound
}

// IsConflictError checks if an error is a conflict error
func IsConflictError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeConflict
}

// IsUnauthorizedError checks if an error is an unauthorized error
func IsUnauthorizedError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeUnauthorized
}

// IsForbiddenError checks if an error is a forbidden error
func IsForbiddenError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeForbidden
}

// IsInternalError checks if an error is an internal error
func IsInternalError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeInternal
}

// IsExternalError checks if an error is an external error
func IsExternalError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeExternal
}

// IsTimeoutError checks if an error is a timeout error
func IsTimeoutError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Type == ErrorTypeTimeout
}
`

const Middleware = `package middleware

import (
	"context"
	"time"
	
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// contextKey is a custom type for context keys
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	
	// CorrelationIDKey is the context key for correlation ID
	CorrelationIDKey contextKey = "correlation_id"
	
	// UserIDKey is the context key for user ID
	UserIDKey contextKey = "user_id"
)

// HandlerFunc represents a Lambda handler function
type HandlerFunc func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// Middleware represents a middleware function
type Middleware func(HandlerFunc) HandlerFunc

// Chain creates a middleware chain
func Chain(middlewares ...Middleware) Middleware {
	return func(handler HandlerFunc) HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// RequestID middleware adds request ID to context
func RequestID() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			requestID := request.RequestContext.RequestID
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			ctx = WithRequestID(ctx, requestID)
			
			response, err := next(ctx, request)
			
			// Add request ID to response headers
			if response.Headers == nil {
				response.Headers = make(map[string]string)
			}
			response.Headers["X-Request-ID"] = requestID
			
			return response, err
		}
	}
}

// Logging middleware logs request and response details
func Logging() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			start := time.Now()
			
			// Log request
			log.Ctx(ctx).Info().
				Str("method", request.HTTPMethod).
				Str("path", request.Path).
				Interface("headers", request.Headers).
				Interface("query_params", request.QueryStringParameters).
				Msg("Incoming request")
			
			response, err := next(ctx, request)
			
			// Log response
			duration := time.Since(start)
			logger := log.Ctx(ctx).With().
				Int("status_code", response.StatusCode).
				Dur("duration_ms", duration).
				Logger()
			
			if err != nil {
				logger.Error().
					Err(err).
					Msg("Request failed")
			} else {
				logger.Info().
					Msg("Request completed")
			}
			
			return response, err
		}
	}
}

// Recovery middleware recovers from panics
func Recovery() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Ctx(ctx).Error().
						Interface("panic", r).
						Msg("Recovered from panic")
					
					response = events.APIGatewayProxyResponse{
						StatusCode: 500,
						Headers: map[string]string{
							"Content-Type": "application/json",
							"X-Request-ID": GetRequestID(ctx),
						},
						Body: ` + "`" + `{"success":false,"error":{"message":"Internal server error"}}` + "`" + `,
					}
				}
			}()
			
			return next(ctx, request)
		}
	}
}

// CORS middleware adds CORS headers
func CORS(allowedOrigins []string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			response, err := next(ctx, request)
			
			// Initialize headers if nil
			if response.Headers == nil {
				response.Headers = make(map[string]string)
			}
			
			// Set CORS headers
			origin := request.Headers["Origin"]
			if origin == "" {
				origin = request.Headers["origin"]
			}
			
			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
			
			if allowed {
				response.Headers["Access-Control-Allow-Origin"] = origin
				response.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
				response.Headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization, X-Request-ID"
				response.Headers["Access-Control-Max-Age"] = "86400"
			}
			
			// Handle preflight requests
			if request.HTTPMethod == "OPTIONS" {
				response.StatusCode = 204
				response.Body = ""
			}
			
			return response, err
		}
	}
}

// Tracing middleware adds AWS X-Ray tracing
func Tracing() Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			// TODO: Implement X-Ray tracing
			// This would integrate with AWS X-Ray SDK
			return next(ctx, request)
		}
	}
}

// Context helper functions

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID gets request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithCorrelationID adds correlation ID to context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID gets correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID gets user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}
`