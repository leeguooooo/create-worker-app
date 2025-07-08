package templates

// Clean Architecture feature-specific templates

const CleanAPIRouter = `package api

import (
	"github.com/gin-gonic/gin"
	"{{.Module}}/internal/usecases"
	"{{.Module}}/pkg/middleware"
)

// Router handles API routing
type Router struct {
	userUseCase usecases.UserUseCase
}

// NewRouter creates a new router
func NewRouter(userUseCase usecases.UserUseCase) *Router {
	return &Router{
		userUseCase: userUseCase,
	}
}

// Setup sets up all routes
func (r *Router) Setup(engine *gin.Engine) {
	// Apply global middleware
	engine.Use(middleware.Logger())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.RequestID())
	
	// API v1 routes
	v1 := engine.Group("/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", r.createUser)
			users.GET("/:id", r.getUser)
			users.GET("", r.listUsers)
			users.PUT("/:id", r.updateUser)
			users.DELETE("/:id", r.deleteUser)
		}
	}
}
`

const CleanAPIHandlers = `package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"{{.Module}}/internal/usecases"
	"{{.Module}}/pkg/errors"
)

// createUser handles user creation
func (r *Router) createUser(c *gin.Context) {
	var input usecases.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := r.userUseCase.CreateUser(c.Request.Context(), input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, output)
}

// getUser handles getting a user
func (r *Router) getUser(c *gin.Context) {
	userID := c.Param("id")
	
	user, err := r.userUseCase.GetUser(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// listUsers handles listing users
func (r *Router) listUsers(c *gin.Context) {
	page := 1
	pageSize := 20
	
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	output, err := r.userUseCase.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, output)
}

// updateUser handles updating a user
func (r *Router) updateUser(c *gin.Context) {
	userID := c.Param("id")
	
	var input usecases.UpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := r.userUseCase.UpdateUser(c.Request.Context(), userID, input)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// deleteUser handles deleting a user
func (r *Router) deleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	if err := r.userUseCase.DeleteUser(c.Request.Context(), userID); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// handleError converts use case errors to HTTP responses
func handleError(c *gin.Context, err error) {
	if ucErr, ok := err.(*usecases.UseCaseError); ok {
		switch ucErr.Type {
		case usecases.ErrTypeValidation:
			c.JSON(http.StatusBadRequest, gin.H{"error": ucErr.Message})
		case usecases.ErrTypeNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": ucErr.Message})
		case usecases.ErrTypeConflict:
			c.JSON(http.StatusConflict, gin.H{"error": ucErr.Message})
		case usecases.ErrTypeUnauthorized:
			c.JSON(http.StatusUnauthorized, gin.H{"error": ucErr.Message})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
`

const CleanAPIMiddleware = `package api

import (
	"github.com/gin-gonic/gin"
	"{{.Module}}/pkg/middleware"
)

// ApplyMiddleware applies API-specific middleware
func ApplyMiddleware(engine *gin.Engine) {
	// CORS middleware
	engine.Use(middleware.CORS([]string{"*"}))
	
	// Rate limiting
	engine.Use(middleware.RateLimit())
	
	// Request validation
	engine.Use(middleware.ValidateHeaders())
}
`

const CleanAPIResponses = `package api

import (
	"time"

	"github.com/google/uuid"
)

// Response is the standard API response
type Response struct {
	Success bool        ` + "`json:\"success\"`" + `
	Data    interface{} ` + "`json:\"data,omitempty\"`" + `
	Error   *Error      ` + "`json:\"error,omitempty\"`" + `
	Meta    *Meta       ` + "`json:\"meta\"`" + `
}

// Error represents an API error
type Error struct {
	Type    string                 ` + "`json:\"type\"`" + `
	Message string                 ` + "`json:\"message\"`" + `
	Details map[string]interface{} ` + "`json:\"details,omitempty\"`" + `
}

// Meta contains response metadata
type Meta struct {
	RequestID string    ` + "`json:\"request_id\"`" + `
	Timestamp time.Time ` + "`json:\"timestamp\"`" + `
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(data interface{}, requestID string) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			RequestID: requestID,
			Timestamp: time.Now().UTC(),
		},
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(errType, message string, requestID string) Response {
	return Response{
		Success: false,
		Error: &Error{
			Type:    errType,
			Message: message,
		},
		Meta: &Meta{
			RequestID: requestID,
			Timestamp: time.Now().UTC(),
		},
	}
}
`

const CleanDynamoDBClient = `package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"{{.Module}}/internal/infrastructure/config"
)

// DynamoDBClient wraps the AWS DynamoDB client
type DynamoDBClient struct {
	client *dynamodb.Client
	config *config.Config
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(cfg *config.Config) (*DynamoDBClient, error) {
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := dynamodb.NewFromConfig(awsConfig)

	// Use custom endpoint for local development
	if cfg.DynamoDBEndpoint != "" {
		client = dynamodb.NewFromConfig(awsConfig, func(o *dynamodb.Options) {
			o.EndpointResolver = dynamodb.EndpointResolverFromURL(cfg.DynamoDBEndpoint)
		})
	}

	return &DynamoDBClient{
		client: client,
		config: cfg,
	}, nil
}

// GetClient returns the underlying DynamoDB client
func (c *DynamoDBClient) GetClient() *dynamodb.Client {
	return c.client
}

// GetTableName returns the table name with prefix
func (c *DynamoDBClient) GetTableName(table string) string {
	return c.config.DynamoDBTablePrefix + table
}
`

const CleanDynamoDBRepository = `package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"{{.Module}}/internal/domain/entities"
	"{{.Module}}/internal/domain/repositories"
)

// dynamoDBUserRepository implements UserRepository using DynamoDB
type dynamoDBUserRepository struct {
	client    *DynamoDBClient
	tableName string
}

// NewDynamoDBUserRepository creates a new DynamoDB user repository
func NewDynamoDBUserRepository(client *DynamoDBClient) repositories.UserRepository {
	return &dynamoDBUserRepository{
		client:    client,
		tableName: client.GetTableName("users"),
	}
}

// Create saves a new user
func (r *dynamoDBUserRepository) Create(ctx context.Context, user *entities.User) error {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = r.client.GetClient().PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(id)"),
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *dynamoDBUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	result, err := r.client.GetClient().GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if result.Item == nil {
		return nil, repositories.ErrUserNotFound
	}

	var user entities.User
	if err := attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *dynamoDBUserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	result, err := r.client.GetClient().Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
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
		return nil, repositories.ErrUserNotFound
	}

	var user entities.User
	if err := attributevalue.UnmarshalMap(result.Items[0], &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// Update saves changes to an existing user
func (r *dynamoDBUserRepository) Update(ctx context.Context, user *entities.User) error {
	user.UpdateVersion()

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	_, err = r.client.GetClient().PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(id) AND version = :old_version"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":old_version": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", user.Version-1)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// Delete removes a user
func (r *dynamoDBUserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.GetClient().DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		ConditionExpression: aws.String("attribute_exists(id)"),
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// List retrieves users with pagination
func (r *dynamoDBUserRepository) List(ctx context.Context, offset, limit int) ([]*entities.User, error) {
	// DynamoDB doesn't support offset directly, need to implement with LastEvaluatedKey
	// This is a simplified implementation
	result, err := r.client.GetClient().Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Limit:     aws.Int32(int32(limit)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]*entities.User, 0, len(result.Items))
	for _, item := range result.Items {
		var user entities.User
		if err := attributevalue.UnmarshalMap(item, &user); err != nil {
			continue
		}
		users = append(users, &user)
	}

	return users, nil
}

// Count returns the total number of users
func (r *dynamoDBUserRepository) Count(ctx context.Context) (int64, error) {
	result, err := r.client.GetClient().Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(r.tableName),
		Select:    types.SelectCount,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return int64(result.Count), nil
}
`

const CleanUserRepository = `package repositories

// UserRepository interface is defined in repositories/interfaces.go
// This would contain the concrete implementation if needed
`

const CleanSQSClient = `package aws

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"{{.Module}}/internal/infrastructure/config"
)

// SQSClient wraps the AWS SQS client
type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSClient creates a new SQS client
func NewSQSClient(cfg *config.Config) (*SQSClient, error) {
	awsConfig, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.AWSRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := sqs.NewFromConfig(awsConfig)

	return &SQSClient{
		client:   client,
		queueURL: cfg.SQSQueueURL,
	}, nil
}

// SendMessage sends a message to the queue
func (c *SQSClient) SendMessage(ctx context.Context, messageType string, payload interface{}) error {
	body, err := json.Marshal(map[string]interface{}{
		"type":    messageType,
		"payload": payload,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = c.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(messageType),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// SendBatch sends multiple messages
func (c *SQSClient) SendBatch(ctx context.Context, messages []Message) error {
	entries := make([]types.SendMessageBatchRequestEntry, 0, len(messages))
	
	for i, msg := range messages {
		body, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message %d: %w", i, err)
		}

		entries = append(entries, types.SendMessageBatchRequestEntry{
			Id:          aws.String(fmt.Sprintf("%d", i)),
			MessageBody: aws.String(string(body)),
		})
	}

	_, err := c.client.SendMessageBatch(ctx, &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(c.queueURL),
		Entries:  entries,
	})
	if err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// Message represents a queue message
type Message struct {
	Type    string      ` + "`json:\"type\"`" + `
	Payload interface{} ` + "`json:\"payload\"`" + `
}
`

const CleanSQSHandler = `package lambda

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"{{.Module}}/internal/usecases"
	"github.com/rs/zerolog/log"
)

// sqsHandler handles SQS messages
type sqsHandler struct {
	processMessageUseCase usecases.ProcessMessageUseCase
}

// NewSQSHandler creates a new SQS handler
func NewSQSHandler(processMessageUseCase usecases.ProcessMessageUseCase) *sqsHandler {
	return &sqsHandler{
		processMessageUseCase: processMessageUseCase,
	}
}

// HandleRequest processes SQS events
func (h *sqsHandler) HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	log.Ctx(ctx).Info().
		Int("message_count", len(sqsEvent.Records)).
		Msg("Processing SQS event")

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

func (h *sqsHandler) processMessage(ctx context.Context, record events.SQSMessage) error {
	var message struct {
		Type    string          ` + "`json:\"type\"`" + `
		Payload json.RawMessage ` + "`json:\"payload\"`" + `
	}

	if err := json.Unmarshal([]byte(record.Body), &message); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to unmarshal message")
		// Don't retry invalid messages
		return nil
	}

	input := usecases.ProcessMessageInput{
		MessageID:   record.MessageId,
		MessageType: message.Type,
		Payload:     message.Payload,
	}

	if err := h.processMessageUseCase.Execute(ctx, input); err != nil {
		return err
	}

	return nil
}
`

const CleanProcessMessageUseCase = `package usecases

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
)

// ProcessMessageInput represents the input for processing a message
type ProcessMessageInput struct {
	MessageID   string
	MessageType string
	Payload     json.RawMessage
}

// ProcessMessageUseCase processes queue messages
type ProcessMessageUseCase interface {
	Execute(ctx context.Context, input ProcessMessageInput) error
}

// processMessageUseCase implements ProcessMessageUseCase
type processMessageUseCase struct {
	// Add dependencies
}

// NewProcessMessageUseCase creates a new process message use case
func NewProcessMessageUseCase() ProcessMessageUseCase {
	return &processMessageUseCase{}
}

// Execute processes the message
func (uc *processMessageUseCase) Execute(ctx context.Context, input ProcessMessageInput) error {
	log.Ctx(ctx).Info().
		Str("message_id", input.MessageID).
		Str("type", input.MessageType).
		Msg("Processing message")

	switch input.MessageType {
	case "user.created":
		return uc.handleUserCreated(ctx, input.Payload)
	case "user.updated":
		return uc.handleUserUpdated(ctx, input.Payload)
	default:
		log.Ctx(ctx).Warn().
			Str("type", input.MessageType).
			Msg("Unknown message type")
		return nil
	}
}

func (uc *processMessageUseCase) handleUserCreated(ctx context.Context, payload json.RawMessage) error {
	var data struct {
		UserID string ` + "`json:\"user_id\"`" + `
		Email  string ` + "`json:\"email\"`" + `
	}

	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Process user created event
	log.Ctx(ctx).Info().
		Str("user_id", data.UserID).
		Str("email", data.Email).
		Msg("Processing user created event")

	return nil
}

func (uc *processMessageUseCase) handleUserUpdated(ctx context.Context, payload json.RawMessage) error {
	// Similar implementation
	return nil
}
`