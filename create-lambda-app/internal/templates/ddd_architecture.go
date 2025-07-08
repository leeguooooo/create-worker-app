package templates

// DDD Architecture specific templates

const DDDAggregateBase = `package aggregate

import (
	"time"

	"github.com/google/uuid"
	"{{.Module}}/domain/event"
)

// AggregateRoot is the base for all aggregate roots
type AggregateRoot struct {
	ID        string    ` + "`json:\"id\"`" + `
	Version   int       ` + "`json:\"version\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
	
	// Uncommitted events
	events []event.DomainEvent
}

// NewAggregateRoot creates a new aggregate root
func NewAggregateRoot(id string) AggregateRoot {
	if id == "" {
		id = uuid.New().String()
	}
	
	now := time.Now().UTC()
	return AggregateRoot{
		ID:        id,
		Version:   0,
		CreatedAt: now,
		UpdatedAt: now,
		events:    make([]event.DomainEvent, 0),
	}
}

// RecordEvent records a domain event
func (a *AggregateRoot) RecordEvent(event event.DomainEvent) {
	a.events = append(a.events, event)
	a.Version++
	a.UpdatedAt = time.Now().UTC()
}

// GetUncommittedEvents returns uncommitted events
func (a *AggregateRoot) GetUncommittedEvents() []event.DomainEvent {
	return a.events
}

// MarkEventsAsCommitted marks events as committed
func (a *AggregateRoot) MarkEventsAsCommitted() {
	a.events = make([]event.DomainEvent, 0)
}

// User aggregate example
type User struct {
	AggregateRoot
	Email         string     ` + "`json:\"email\"`" + `
	Name          string     ` + "`json:\"name\"`" + `
	Status        UserStatus ` + "`json:\"status\"`" + `
	PasswordHash  string     ` + "`json:\"-\"`" + `
	EmailVerified bool       ` + "`json:\"email_verified\"`" + `
	DeletedAt     *time.Time ` + "`json:\"deleted_at,omitempty\"`" + `
}

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBlocked  UserStatus = "blocked"
	UserStatusDeleted  UserStatus = "deleted"
)

// NewUser creates a new user aggregate
func NewUser(email, name string) (*User, error) {
	// Validate email
	if email == "" {
		return nil, ErrInvalidEmail
	}
	
	// Validate name
	if name == "" {
		return nil, ErrInvalidName
	}
	
	user := &User{
		AggregateRoot: NewAggregateRoot(""),
		Email:         email,
		Name:          name,
		Status:        UserStatusActive,
		EmailVerified: false,
	}
	
	// Record domain event
	user.RecordEvent(&event.UserCreated{
		UserID: user.ID,
		Email:  email,
		Name:   name,
	})
	
	return user, nil
}

// UpdateProfile updates user profile
func (u *User) UpdateProfile(name string) error {
	if name == "" {
		return ErrInvalidName
	}
	
	oldName := u.Name
	u.Name = name
	
	// Record domain event
	u.RecordEvent(&event.UserProfileUpdated{
		UserID:  u.ID,
		OldName: oldName,
		NewName: name,
	})
	
	return nil
}

// ChangeStatus changes user status
func (u *User) ChangeStatus(status UserStatus) error {
	if u.Status == status {
		return nil // No change
	}
	
	oldStatus := u.Status
	u.Status = status
	
	// Record domain event
	u.RecordEvent(&event.UserStatusChanged{
		UserID:    u.ID,
		OldStatus: string(oldStatus),
		NewStatus: string(status),
	})
	
	return nil
}

// Delete soft deletes the user
func (u *User) Delete() error {
	if u.Status == UserStatusDeleted {
		return ErrUserAlreadyDeleted
	}
	
	now := time.Now().UTC()
	u.DeletedAt = &now
	u.Status = UserStatusDeleted
	
	// Record domain event
	u.RecordEvent(&event.UserDeleted{
		UserID:    u.ID,
		DeletedAt: now,
	})
	
	return nil
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive && u.DeletedAt == nil
}

// Domain errors
var (
	ErrInvalidEmail       = NewDomainError("INVALID_EMAIL", "Invalid email address")
	ErrInvalidName        = NewDomainError("INVALID_NAME", "Invalid name")
	ErrUserAlreadyDeleted = NewDomainError("USER_ALREADY_DELETED", "User is already deleted")
)

// DomainError represents a domain-level error
type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}
`

const DDDEntityBase = `package entity

import (
	"time"

	"github.com/google/uuid"
)

// Entity is the base for all entities
type Entity struct {
	ID        string    ` + "`json:\"id\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// NewEntity creates a new entity
func NewEntity(id string) Entity {
	if id == "" {
		id = uuid.New().String()
	}
	
	now := time.Now().UTC()
	return Entity{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Equals checks if two entities are equal
func (e *Entity) Equals(other *Entity) bool {
	if e == nil || other == nil {
		return false
	}
	return e.ID == other.ID
}

// SetUpdatedAt updates the timestamp
func (e *Entity) SetUpdatedAt() {
	e.UpdatedAt = time.Now().UTC()
}

// Profile entity example
type Profile struct {
	Entity
	UserID      string             ` + "`json:\"user_id\"`" + `
	DisplayName string             ` + "`json:\"display_name\"`" + `
	Bio         string             ` + "`json:\"bio\"`" + `
	AvatarURL   string             ` + "`json:\"avatar_url\"`" + `
	Preferences map[string]string  ` + "`json:\"preferences\"`" + `
}

// NewProfile creates a new profile entity
func NewProfile(userID, displayName string) *Profile {
	return &Profile{
		Entity:      NewEntity(""),
		UserID:      userID,
		DisplayName: displayName,
		Preferences: make(map[string]string),
	}
}

// UpdateBio updates the profile bio
func (p *Profile) UpdateBio(bio string) {
	p.Bio = bio
	p.SetUpdatedAt()
}

// SetPreference sets a user preference
func (p *Profile) SetPreference(key, value string) {
	if p.Preferences == nil {
		p.Preferences = make(map[string]string)
	}
	p.Preferences[key] = value
	p.SetUpdatedAt()
}

// GetPreference gets a user preference
func (p *Profile) GetPreference(key string) (string, bool) {
	if p.Preferences == nil {
		return "", false
	}
	value, exists := p.Preferences[key]
	return value, exists
}
`

const DDDValueObject = `package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

// ValueObject is the interface for all value objects
type ValueObject interface {
	Equals(other ValueObject) bool
	String() string
}

// Email represents an email address value object
type Email struct {
	value string
}

// NewEmail creates a new Email value object
func NewEmail(value string) (*Email, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	
	if !isValidEmail(value) {
		return nil, errors.New("invalid email format")
	}
	
	return &Email{value: value}, nil
}

// Value returns the email value
func (e *Email) Value() string {
	return e.value
}

// String returns the string representation
func (e *Email) String() string {
	return e.value
}

// Equals checks if two emails are equal
func (e *Email) Equals(other ValueObject) bool {
	if otherEmail, ok := other.(*Email); ok {
		return e.value == otherEmail.value
	}
	return false
}

// Domain returns the domain part of the email
func (e *Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	pattern := ` + "`" + `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` + "`" + `
	match, _ := regexp.MatchString(pattern, email)
	return match
}

// Money represents a monetary value
type Money struct {
	amount   int64  // Amount in cents
	currency string // ISO 4217 currency code
}

// NewMoney creates a new Money value object
func NewMoney(amount int64, currency string) (*Money, error) {
	if currency == "" {
		return nil, errors.New("currency is required")
	}
	
	if len(currency) != 3 {
		return nil, errors.New("currency must be a 3-letter ISO code")
	}
	
	return &Money{
		amount:   amount,
		currency: strings.ToUpper(currency),
	}, nil
}

// Amount returns the amount in cents
func (m *Money) Amount() int64 {
	return m.amount
}

// Currency returns the currency code
func (m *Money) Currency() string {
	return m.currency
}

// String returns the string representation
func (m *Money) String() string {
	dollars := float64(m.amount) / 100
	return fmt.Sprintf("%.2f %s", dollars, m.currency)
}

// Equals checks if two money values are equal
func (m *Money) Equals(other ValueObject) bool {
	if otherMoney, ok := other.(*Money); ok {
		return m.amount == otherMoney.amount && m.currency == otherMoney.currency
	}
	return false
}

// Add adds two money values
func (m *Money) Add(other *Money) (*Money, error) {
	if m.currency != other.currency {
		return nil, errors.New("cannot add money with different currencies")
	}
	
	return &Money{
		amount:   m.amount + other.amount,
		currency: m.currency,
	}, nil
}

// Address represents a physical address
type Address struct {
	street     string
	city       string
	state      string
	postalCode string
	country    string
}

// NewAddress creates a new Address value object
func NewAddress(street, city, state, postalCode, country string) (*Address, error) {
	if street == "" || city == "" || country == "" {
		return nil, errors.New("street, city, and country are required")
	}
	
	return &Address{
		street:     street,
		city:       city,
		state:      state,
		postalCode: postalCode,
		country:    country,
	}, nil
}

// String returns the string representation
func (a *Address) String() string {
	parts := []string{a.street}
	
	if a.city != "" {
		parts = append(parts, a.city)
	}
	if a.state != "" {
		parts = append(parts, a.state)
	}
	if a.postalCode != "" {
		parts = append(parts, a.postalCode)
	}
	if a.country != "" {
		parts = append(parts, a.country)
	}
	
	return strings.Join(parts, ", ")
}

// Equals checks if two addresses are equal
func (a *Address) Equals(other ValueObject) bool {
	if otherAddr, ok := other.(*Address); ok {
		return a.street == otherAddr.street &&
			a.city == otherAddr.city &&
			a.state == otherAddr.state &&
			a.postalCode == otherAddr.postalCode &&
			a.country == otherAddr.country
	}
	return false
}
`

const DDDRepository = `package repository

import (
	"context"

	"{{.Module}}/domain/aggregate"
)

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// Save saves a user aggregate
	Save(ctx context.Context, user *aggregate.User) error
	
	// FindByID finds a user by ID
	FindByID(ctx context.Context, id string) (*aggregate.User, error)
	
	// FindByEmail finds a user by email
	FindByEmail(ctx context.Context, email string) (*aggregate.User, error)
	
	// List lists users with pagination
	List(ctx context.Context, offset, limit int) ([]*aggregate.User, error)
	
	// Count returns the total number of users
	Count(ctx context.Context) (int64, error)
	
	// Delete deletes a user
	Delete(ctx context.Context, id string) error
}

// ProfileRepository defines the interface for profile persistence
type ProfileRepository interface {
	// Save saves a profile
	Save(ctx context.Context, profile *entity.Profile) error
	
	// FindByUserID finds a profile by user ID
	FindByUserID(ctx context.Context, userID string) (*entity.Profile, error)
	
	// Delete deletes a profile
	Delete(ctx context.Context, userID string) error
}

// UnitOfWork defines the interface for managing transactions
type UnitOfWork interface {
	// UserRepository returns the user repository
	UserRepository() UserRepository
	
	// ProfileRepository returns the profile repository
	ProfileRepository() ProfileRepository
	
	// Begin begins a transaction
	Begin(ctx context.Context) error
	
	// Commit commits the transaction
	Commit() error
	
	// Rollback rolls back the transaction
	Rollback() error
}

// Specification defines a query specification
type Specification interface {
	// IsSatisfiedBy checks if the specification is satisfied
	IsSatisfiedBy(aggregate interface{}) bool
	
	// And creates an AND specification
	And(spec Specification) Specification
	
	// Or creates an OR specification
	Or(spec Specification) Specification
	
	// Not creates a NOT specification
	Not() Specification
}

// BaseSpecification provides base implementation
type BaseSpecification struct {
	predicate func(interface{}) bool
}

// IsSatisfiedBy checks if the specification is satisfied
func (s *BaseSpecification) IsSatisfiedBy(aggregate interface{}) bool {
	return s.predicate(aggregate)
}

// And creates an AND specification
func (s *BaseSpecification) And(spec Specification) Specification {
	return &BaseSpecification{
		predicate: func(aggregate interface{}) bool {
			return s.IsSatisfiedBy(aggregate) && spec.IsSatisfiedBy(aggregate)
		},
	}
}

// Or creates an OR specification
func (s *BaseSpecification) Or(spec Specification) Specification {
	return &BaseSpecification{
		predicate: func(aggregate interface{}) bool {
			return s.IsSatisfiedBy(aggregate) || spec.IsSatisfiedBy(aggregate)
		},
	}
}

// Not creates a NOT specification
func (s *BaseSpecification) Not() Specification {
	return &BaseSpecification{
		predicate: func(aggregate interface{}) bool {
			return !s.IsSatisfiedBy(aggregate)
		},
	}
}

// ActiveUserSpecification is a specification for active users
func ActiveUserSpecification() Specification {
	return &BaseSpecification{
		predicate: func(aggregate interface{}) bool {
			if user, ok := aggregate.(*aggregate.User); ok {
				return user.IsActive()
			}
			return false
		},
	}
}
`

const DDDEvent = `package event

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the base interface for all domain events
type DomainEvent interface {
	EventID() string
	EventType() string
	OccurredAt() time.Time
	AggregateID() string
}

// BaseDomainEvent provides base implementation for domain events
type BaseDomainEvent struct {
	ID          string    ` + "`json:\"id\"`" + `
	Type        string    ` + "`json:\"type\"`" + `
	AggregateId string    ` + "`json:\"aggregate_id\"`" + `
	Timestamp   time.Time ` + "`json:\"timestamp\"`" + `
}

// NewBaseDomainEvent creates a new base domain event
func NewBaseDomainEvent(eventType, aggregateID string) BaseDomainEvent {
	return BaseDomainEvent{
		ID:          uuid.New().String(),
		Type:        eventType,
		AggregateId: aggregateID,
		Timestamp:   time.Now().UTC(),
	}
}

// EventID returns the event ID
func (e BaseDomainEvent) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e BaseDomainEvent) EventType() string {
	return e.Type
}

// OccurredAt returns when the event occurred
func (e BaseDomainEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// AggregateID returns the aggregate ID
func (e BaseDomainEvent) AggregateID() string {
	return e.AggregateId
}

// User domain events

// UserCreated event
type UserCreated struct {
	BaseDomainEvent
	UserID string ` + "`json:\"user_id\"`" + `
	Email  string ` + "`json:\"email\"`" + `
	Name   string ` + "`json:\"name\"`" + `
}

// NewUserCreated creates a new UserCreated event
func NewUserCreated(userID, email, name string) *UserCreated {
	return &UserCreated{
		BaseDomainEvent: NewBaseDomainEvent("user.created", userID),
		UserID:          userID,
		Email:           email,
		Name:            name,
	}
}

// UserProfileUpdated event
type UserProfileUpdated struct {
	BaseDomainEvent
	UserID  string ` + "`json:\"user_id\"`" + `
	OldName string ` + "`json:\"old_name\"`" + `
	NewName string ` + "`json:\"new_name\"`" + `
}

// UserStatusChanged event
type UserStatusChanged struct {
	BaseDomainEvent
	UserID    string ` + "`json:\"user_id\"`" + `
	OldStatus string ` + "`json:\"old_status\"`" + `
	NewStatus string ` + "`json:\"new_status\"`" + `
}

// UserDeleted event
type UserDeleted struct {
	BaseDomainEvent
	UserID    string    ` + "`json:\"user_id\"`" + `
	DeletedAt time.Time ` + "`json:\"deleted_at\"`" + `
}

// EventStore interface for persisting events
type EventStore interface {
	// Save saves events
	Save(ctx context.Context, events []DomainEvent) error
	
	// GetEvents gets events for an aggregate
	GetEvents(ctx context.Context, aggregateID string) ([]DomainEvent, error)
	
	// GetEventsSince gets events since a specific time
	GetEventsSince(ctx context.Context, since time.Time) ([]DomainEvent, error)
}

// EventPublisher interface for publishing events
type EventPublisher interface {
	// Publish publishes domain events
	Publish(ctx context.Context, events []DomainEvent) error
}

// EventHandler interface for handling events
type EventHandler interface {
	// Handle handles a domain event
	Handle(ctx context.Context, event DomainEvent) error
	
	// CanHandle checks if the handler can handle the event
	CanHandle(event DomainEvent) bool
}

// EventBus interface for event distribution
type EventBus interface {
	// Subscribe subscribes a handler to events
	Subscribe(handler EventHandler)
	
	// Publish publishes events to all handlers
	Publish(ctx context.Context, events []DomainEvent) error
}
`

const DDDCommand = `package command

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Command is the base interface for all commands
type Command interface {
	CommandID() string
	CommandType() string
}

// BaseCommand provides base implementation for commands
type BaseCommand struct {
	ID   string ` + "`json:\"id\"`" + `
	Type string ` + "`json:\"type\"`" + `
}

// NewBaseCommand creates a new base command
func NewBaseCommand(commandType string) BaseCommand {
	return BaseCommand{
		ID:   uuid.New().String(),
		Type: commandType,
	}
}

// CommandID returns the command ID
func (c BaseCommand) CommandID() string {
	return c.ID
}

// CommandType returns the command type
func (c BaseCommand) CommandType() string {
	return c.Type
}

// CommandHandler handles commands
type CommandHandler interface {
	Handle(ctx context.Context, command Command) error
}

// CommandBus dispatches commands to handlers
type CommandBus interface {
	// Register registers a handler for a command type
	Register(commandType string, handler CommandHandler)
	
	// Dispatch dispatches a command to its handler
	Dispatch(ctx context.Context, command Command) error
}

// User commands

// CreateUserCommand creates a new user
type CreateUserCommand struct {
	BaseCommand
	Email string ` + "`json:\"email\"`" + `
	Name  string ` + "`json:\"name\"`" + `
}

// NewCreateUserCommand creates a new CreateUserCommand
func NewCreateUserCommand(email, name string) *CreateUserCommand {
	return &CreateUserCommand{
		BaseCommand: NewBaseCommand("user.create"),
		Email:       email,
		Name:        name,
	}
}

// UpdateUserProfileCommand updates user profile
type UpdateUserProfileCommand struct {
	BaseCommand
	UserID string ` + "`json:\"user_id\"`" + `
	Name   string ` + "`json:\"name\"`" + `
}

// ChangeUserStatusCommand changes user status
type ChangeUserStatusCommand struct {
	BaseCommand
	UserID string ` + "`json:\"user_id\"`" + `
	Status string ` + "`json:\"status\"`" + `
}

// DeleteUserCommand deletes a user
type DeleteUserCommand struct {
	BaseCommand
	UserID string ` + "`json:\"user_id\"`" + `
}

// Command handlers

// CreateUserHandler handles user creation
type CreateUserHandler struct {
	userRepo repository.UserRepository
	eventBus event.EventBus
}

// NewCreateUserHandler creates a new CreateUserHandler
func NewCreateUserHandler(userRepo repository.UserRepository, eventBus event.EventBus) *CreateUserHandler {
	return &CreateUserHandler{
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

// Handle handles the CreateUserCommand
func (h *CreateUserHandler) Handle(ctx context.Context, cmd Command) error {
	createCmd, ok := cmd.(*CreateUserCommand)
	if !ok {
		return errors.New("invalid command type")
	}
	
	// Check if user already exists
	existing, _ := h.userRepo.FindByEmail(ctx, createCmd.Email)
	if existing != nil {
		return errors.New("user already exists")
	}
	
	// Create user aggregate
	user, err := aggregate.NewUser(createCmd.Email, createCmd.Name)
	if err != nil {
		return err
	}
	
	// Save user
	if err := h.userRepo.Save(ctx, user); err != nil {
		return err
	}
	
	// Publish events
	if err := h.eventBus.Publish(ctx, user.GetUncommittedEvents()); err != nil {
		return err
	}
	
	// Mark events as committed
	user.MarkEventsAsCommitted()
	
	return nil
}
`

const DDDQuery = `package query

import (
	"context"
)

// Query is the base interface for all queries
type Query interface {
	QueryType() string
}

// QueryHandler handles queries
type QueryHandler interface {
	Handle(ctx context.Context, query Query) (interface{}, error)
}

// QueryBus dispatches queries to handlers
type QueryBus interface {
	// Register registers a handler for a query type
	Register(queryType string, handler QueryHandler)
	
	// Dispatch dispatches a query to its handler
	Dispatch(ctx context.Context, query Query) (interface{}, error)
}

// User queries

// GetUserByIDQuery gets a user by ID
type GetUserByIDQuery struct {
	UserID string ` + "`json:\"user_id\"`" + `
}

// QueryType returns the query type
func (q *GetUserByIDQuery) QueryType() string {
	return "user.get_by_id"
}

// GetUserByEmailQuery gets a user by email
type GetUserByEmailQuery struct {
	Email string ` + "`json:\"email\"`" + `
}

// QueryType returns the query type
func (q *GetUserByEmailQuery) QueryType() string {
	return "user.get_by_email"
}

// ListUsersQuery lists users with pagination
type ListUsersQuery struct {
	Page     int    ` + "`json:\"page\"`" + `
	PageSize int    ` + "`json:\"page_size\"`" + `
	Status   string ` + "`json:\"status,omitempty\"`" + `
}

// QueryType returns the query type
func (q *ListUsersQuery) QueryType() string {
	return "user.list"
}

// User DTOs

// UserDTO is a data transfer object for users
type UserDTO struct {
	ID            string    ` + "`json:\"id\"`" + `
	Email         string    ` + "`json:\"email\"`" + `
	Name          string    ` + "`json:\"name\"`" + `
	Status        string    ` + "`json:\"status\"`" + `
	EmailVerified bool      ` + "`json:\"email_verified\"`" + `
	CreatedAt     time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt     time.Time ` + "`json:\"updated_at\"`" + `
}

// ListUsersResult is the result of listing users
type ListUsersResult struct {
	Users      []*UserDTO ` + "`json:\"users\"`" + `
	TotalCount int64      ` + "`json:\"total_count\"`" + `
	Page       int        ` + "`json:\"page\"`" + `
	PageSize   int        ` + "`json:\"page_size\"`" + `
}

// Query handlers

// GetUserByIDHandler handles GetUserByIDQuery
type GetUserByIDHandler struct {
	userRepo repository.UserRepository
}

// NewGetUserByIDHandler creates a new GetUserByIDHandler
func NewGetUserByIDHandler(userRepo repository.UserRepository) *GetUserByIDHandler {
	return &GetUserByIDHandler{
		userRepo: userRepo,
	}
}

// Handle handles the query
func (h *GetUserByIDHandler) Handle(ctx context.Context, q Query) (interface{}, error) {
	query, ok := q.(*GetUserByIDQuery)
	if !ok {
		return nil, errors.New("invalid query type")
	}
	
	user, err := h.userRepo.FindByID(ctx, query.UserID)
	if err != nil {
		return nil, err
	}
	
	return &UserDTO{
		ID:            user.ID,
		Email:         user.Email,
		Name:          user.Name,
		Status:        string(user.Status),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

// ListUsersHandler handles ListUsersQuery
type ListUsersHandler struct {
	userRepo repository.UserRepository
}

// Handle handles the query
func (h *ListUsersHandler) Handle(ctx context.Context, q Query) (interface{}, error) {
	query, ok := q.(*ListUsersQuery)
	if !ok {
		return nil, errors.New("invalid query type")
	}
	
	// Calculate offset
	offset := (query.Page - 1) * query.PageSize
	
	// Get users
	users, err := h.userRepo.List(ctx, offset, query.PageSize)
	if err != nil {
		return nil, err
	}
	
	// Get total count
	totalCount, err := h.userRepo.Count(ctx)
	if err != nil {
		return nil, err
	}
	
	// Convert to DTOs
	userDTOs := make([]*UserDTO, len(users))
	for i, user := range users {
		userDTOs[i] = &UserDTO{
			ID:            user.ID,
			Email:         user.Email,
			Name:          user.Name,
			Status:        string(user.Status),
			EmailVerified: user.EmailVerified,
			CreatedAt:     user.CreatedAt,
			UpdatedAt:     user.UpdatedAt,
		}
	}
	
	return &ListUsersResult{
		Users:      userDTOs,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
	}, nil
}
`

const DDDPersistence = `package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"{{.Module}}/domain/aggregate"
	"{{.Module}}/domain/repository"
)

// DynamoDBUserRepository implements UserRepository using DynamoDB
type DynamoDBUserRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBUserRepository creates a new DynamoDB user repository
func NewDynamoDBUserRepository(client *dynamodb.Client, tableName string) *DynamoDBUserRepository {
	return &DynamoDBUserRepository{
		client:    client,
		tableName: tableName,
	}
}

// Save saves a user aggregate
func (r *DynamoDBUserRepository) Save(ctx context.Context, user *aggregate.User) error {
	// Convert to DynamoDB item
	item := map[string]interface{}{
		"pk":             "USER#" + user.ID,
		"sk":             "USER#" + user.ID,
		"type":           "User",
		"id":             user.ID,
		"email":          user.Email,
		"name":           user.Name,
		"status":         string(user.Status),
		"email_verified": user.EmailVerified,
		"password_hash":  user.PasswordHash,
		"created_at":     user.CreatedAt,
		"updated_at":     user.UpdatedAt,
		"version":        user.Version,
	}
	
	if user.DeletedAt != nil {
		item["deleted_at"] = user.DeletedAt
	}
	
	// Marshal to DynamoDB attributes
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}
	
	// Save to DynamoDB with optimistic locking
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      av,
		ConditionExpression: aws.String("attribute_not_exists(pk) OR version = :old_version"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":old_version": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", user.Version-1)},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}
	
	return nil
}

// FindByID finds a user by ID
func (r *DynamoDBUserRepository) FindByID(ctx context.Context, id string) (*aggregate.User, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "USER#" + id},
			"sk": &types.AttributeValueMemberS{Value: "USER#" + id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if result.Item == nil {
		return nil, repository.ErrUserNotFound
	}
	
	return r.unmarshalUser(result.Item)
}

// FindByEmail finds a user by email
func (r *DynamoDBUserRepository) FindByEmail(ctx context.Context, email string) (*aggregate.User, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
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
		return nil, repository.ErrUserNotFound
	}
	
	return r.unmarshalUser(result.Items[0])
}

// List lists users with pagination
func (r *DynamoDBUserRepository) List(ctx context.Context, offset, limit int) ([]*aggregate.User, error) {
	var users []*aggregate.User
	var lastEvaluatedKey map[string]types.AttributeValue
	
	// DynamoDB doesn't support offset directly, so we need to scan
	scanned := 0
	for {
		input := &dynamodb.ScanInput{
			TableName:        aws.String(r.tableName),
			FilterExpression: aws.String("#type = :type"),
			ExpressionAttributeNames: map[string]string{
				"#type": "type",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":type": &types.AttributeValueMemberS{Value: "User"},
			},
			Limit: aws.Int32(int32(limit + offset - scanned)),
		}
		
		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}
		
		result, err := r.client.Scan(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to scan users: %w", err)
		}
		
		for _, item := range result.Items {
			if scanned >= offset && len(users) < limit {
				user, err := r.unmarshalUser(item)
				if err == nil {
					users = append(users, user)
				}
			}
			scanned++
		}
		
		if result.LastEvaluatedKey == nil || len(users) >= limit {
			break
		}
		
		lastEvaluatedKey = result.LastEvaluatedKey
	}
	
	return users, nil
}

// Count returns the total number of users
func (r *DynamoDBUserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	var lastEvaluatedKey map[string]types.AttributeValue
	
	for {
		input := &dynamodb.ScanInput{
			TableName:        aws.String(r.tableName),
			FilterExpression: aws.String("#type = :type"),
			ExpressionAttributeNames: map[string]string{
				"#type": "type",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":type": &types.AttributeValueMemberS{Value: "User"},
			},
			Select: types.SelectCount,
		}
		
		if lastEvaluatedKey != nil {
			input.ExclusiveStartKey = lastEvaluatedKey
		}
		
		result, err := r.client.Scan(ctx, input)
		if err != nil {
			return 0, fmt.Errorf("failed to count users: %w", err)
		}
		
		count += int64(result.Count)
		
		if result.LastEvaluatedKey == nil {
			break
		}
		
		lastEvaluatedKey = result.LastEvaluatedKey
	}
	
	return count, nil
}

// Delete deletes a user
func (r *DynamoDBUserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "USER#" + id},
			"sk": &types.AttributeValueMemberS{Value: "USER#" + id},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	
	return nil
}

// unmarshalUser unmarshals a DynamoDB item to a user aggregate
func (r *DynamoDBUserRepository) unmarshalUser(item map[string]types.AttributeValue) (*aggregate.User, error) {
	var data struct {
		ID            string     ` + "`dynamodbav:\"id\"`" + `
		Email         string     ` + "`dynamodbav:\"email\"`" + `
		Name          string     ` + "`dynamodbav:\"name\"`" + `
		Status        string     ` + "`dynamodbav:\"status\"`" + `
		EmailVerified bool       ` + "`dynamodbav:\"email_verified\"`" + `
		PasswordHash  string     ` + "`dynamodbav:\"password_hash\"`" + `
		CreatedAt     time.Time  ` + "`dynamodbav:\"created_at\"`" + `
		UpdatedAt     time.Time  ` + "`dynamodbav:\"updated_at\"`" + `
		DeletedAt     *time.Time ` + "`dynamodbav:\"deleted_at\"`" + `
		Version       int        ` + "`dynamodbav:\"version\"`" + `
	}
	
	if err := attributevalue.UnmarshalMap(item, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}
	
	// Reconstruct user aggregate
	user := &aggregate.User{
		AggregateRoot: aggregate.AggregateRoot{
			ID:        data.ID,
			Version:   data.Version,
			CreatedAt: data.CreatedAt,
			UpdatedAt: data.UpdatedAt,
		},
		Email:         data.Email,
		Name:          data.Name,
		Status:        aggregate.UserStatus(data.Status),
		EmailVerified: data.EmailVerified,
		PasswordHash:  data.PasswordHash,
		DeletedAt:     data.DeletedAt,
	}
	
	return user, nil
}
`

// DDD Lambda handlers
const DDDAPIRouter = `package api

import (
	"github.com/gin-gonic/gin"
	"{{.Module}}/application/command"
	"{{.Module}}/application/query"
)

// Router holds the API dependencies
type Router struct {
	commandBus command.CommandBus
	queryBus   query.QueryBus
}

// NewRouter creates a new API router
func NewRouter(commandBus command.CommandBus, queryBus query.QueryBus) *Router {
	return &Router{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

// SetupRoutes sets up the API routes
func (r *Router) SetupRoutes(engine *gin.Engine) {
	// Health check
	engine.GET("/health", r.healthCheck)
	
	// User routes
	users := engine.Group("/users")
	{
		users.POST("", r.createUser)
		users.GET("/:id", r.getUser)
		users.GET("", r.listUsers)
		users.PUT("/:id", r.updateUser)
		users.DELETE("/:id", r.deleteUser)
	}
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
		"service": "{{.Name}}",
	})
}

func (r *Router) createUser(c *gin.Context) {
	var req struct {
		Email string ` + "`json:\"email\" binding:\"required,email\"`" + `
		Name  string ` + "`json:\"name\" binding:\"required\"`" + `
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	// Create command
	cmd := command.NewCreateUserCommand(req.Email, req.Name)
	
	// Dispatch command
	if err := r.commandBus.Dispatch(c.Request.Context(), cmd); err != nil {
		handleError(c, err)
		return
	}
	
	c.JSON(201, gin.H{"success": true})
}

func (r *Router) getUser(c *gin.Context) {
	userID := c.Param("id")
	
	// Create query
	q := &query.GetUserByIDQuery{UserID: userID}
	
	// Dispatch query
	result, err := r.queryBus.Dispatch(c.Request.Context(), q)
	if err != nil {
		handleError(c, err)
		return
	}
	
	c.JSON(200, result)
}

func (r *Router) listUsers(c *gin.Context) {
	page := 1
	pageSize := 20
	
	// Parse query parameters
	if p := c.Query("page"); p != "" {
		// Parse page
	}
	if ps := c.Query("page_size"); ps != "" {
		// Parse page size
	}
	
	// Create query
	q := &query.ListUsersQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
	}
	
	// Dispatch query
	result, err := r.queryBus.Dispatch(c.Request.Context(), q)
	if err != nil {
		handleError(c, err)
		return
	}
	
	c.JSON(200, result)
}

func (r *Router) updateUser(c *gin.Context) {
	userID := c.Param("id")
	
	var req struct {
		Name   *string ` + "`json:\"name\"`" + `
		Status *string ` + "`json:\"status\"`" + `
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	
	// Handle profile update
	if req.Name != nil {
		cmd := &command.UpdateUserProfileCommand{
			BaseCommand: command.NewBaseCommand("user.update_profile"),
			UserID:      userID,
			Name:        *req.Name,
		}
		
		if err := r.commandBus.Dispatch(c.Request.Context(), cmd); err != nil {
			handleError(c, err)
			return
		}
	}
	
	// Handle status change
	if req.Status != nil {
		cmd := &command.ChangeUserStatusCommand{
			BaseCommand: command.NewBaseCommand("user.change_status"),
			UserID:      userID,
			Status:      *req.Status,
		}
		
		if err := r.commandBus.Dispatch(c.Request.Context(), cmd); err != nil {
			handleError(c, err)
			return
		}
	}
	
	c.JSON(200, gin.H{"success": true})
}

func (r *Router) deleteUser(c *gin.Context) {
	userID := c.Param("id")
	
	// Create command
	cmd := &command.DeleteUserCommand{
		BaseCommand: command.NewBaseCommand("user.delete"),
		UserID:      userID,
	}
	
	// Dispatch command
	if err := r.commandBus.Dispatch(c.Request.Context(), cmd); err != nil {
		handleError(c, err)
		return
	}
	
	c.Status(204)
}

func handleError(c *gin.Context, err error) {
	// Handle domain errors
	if domainErr, ok := err.(*aggregate.DomainError); ok {
		switch domainErr.Code {
		case "USER_NOT_FOUND":
			c.JSON(404, gin.H{"error": domainErr.Message})
		case "USER_EXISTS":
			c.JSON(409, gin.H{"error": domainErr.Message})
		default:
			c.JSON(400, gin.H{"error": domainErr.Message})
		}
		return
	}
	
	// Default error
	c.JSON(500, gin.H{"error": "Internal server error"})
}
`

const DDDAPIHandlers = `package api

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"{{.Module}}/application/command"
	"{{.Module}}/application/query"
	"{{.Module}}/infrastructure/config"
	"github.com/rs/zerolog/log"
)

var ginLambda *ginadapter.GinLambda

// Handler handles API Gateway requests
type Handler struct {
	router *Router
}

// NewHandler creates a new API handler
func NewHandler(commandBus command.CommandBus, queryBus query.QueryBus) *Handler {
	// Create Gin engine
	engine := gin.New()
	engine.Use(gin.Recovery())
	
	// Create router and setup routes
	router := NewRouter(commandBus, queryBus)
	router.SetupRoutes(engine)
	
	// Create Gin Lambda adapter
	ginLambda = ginadapter.New(engine)
	
	return &Handler{
		router: router,
	}
}

// HandleRequest handles the Lambda request
func (h *Handler) HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Add request ID to logs
	log.Ctx(ctx).Info().
		Str("method", request.HTTPMethod).
		Str("path", request.Path).
		Str("request_id", request.RequestContext.RequestID).
		Msg("Processing API request")
	
	// Process request
	return ginLambda.ProxyWithContext(ctx, request)
}

// Start starts the Lambda handler
func Start() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
	
	// Initialize infrastructure
	infra, err := infrastructure.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize infrastructure")
	}
	
	// Create command bus
	commandBus := infrastructure.NewCommandBus()
	
	// Register command handlers
	commandBus.Register("user.create", command.NewCreateUserHandler(infra.UserRepository(), infra.EventBus()))
	// Register other command handlers...
	
	// Create query bus
	queryBus := infrastructure.NewQueryBus()
	
	// Register query handlers
	queryBus.Register("user.get_by_id", query.NewGetUserByIDHandler(infra.UserRepository()))
	queryBus.Register("user.list", query.NewListUsersHandler(infra.UserRepository()))
	// Register other query handlers...
	
	// Create handler
	handler := NewHandler(commandBus, queryBus)
	
	// Start Lambda
	lambda.Start(handler.HandleRequest)
}
`

const DDDAPIApplicationHandler = `package handler

import (
	"context"

	"{{.Module}}/application/command"
	"{{.Module}}/application/query"
)

// APIHandler handles API requests using CQRS
type APIHandler struct {
	commandBus command.CommandBus
	queryBus   query.QueryBus
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(commandBus command.CommandBus, queryBus query.QueryBus) *APIHandler {
	return &APIHandler{
		commandBus: commandBus,
		queryBus:   queryBus,
	}
}

// CreateUser handles user creation
func (h *APIHandler) CreateUser(ctx context.Context, email, name string) error {
	cmd := command.NewCreateUserCommand(email, name)
	return h.commandBus.Dispatch(ctx, cmd)
}

// GetUser handles getting a user
func (h *APIHandler) GetUser(ctx context.Context, userID string) (*query.UserDTO, error) {
	q := &query.GetUserByIDQuery{UserID: userID}
	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	
	return result.(*query.UserDTO), nil
}

// ListUsers handles listing users
func (h *APIHandler) ListUsers(ctx context.Context, page, pageSize int, status string) (*query.ListUsersResult, error) {
	q := &query.ListUsersQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	}
	
	result, err := h.queryBus.Dispatch(ctx, q)
	if err != nil {
		return nil, err
	}
	
	return result.(*query.ListUsersResult), nil
}

// UpdateUserProfile handles updating user profile
func (h *APIHandler) UpdateUserProfile(ctx context.Context, userID, name string) error {
	cmd := &command.UpdateUserProfileCommand{
		BaseCommand: command.NewBaseCommand("user.update_profile"),
		UserID:      userID,
		Name:        name,
	}
	
	return h.commandBus.Dispatch(ctx, cmd)
}

// ChangeUserStatus handles changing user status
func (h *APIHandler) ChangeUserStatus(ctx context.Context, userID, status string) error {
	cmd := &command.ChangeUserStatusCommand{
		BaseCommand: command.NewBaseCommand("user.change_status"),
		UserID:      userID,
		Status:      status,
	}
	
	return h.commandBus.Dispatch(ctx, cmd)
}

// DeleteUser handles user deletion
func (h *APIHandler) DeleteUser(ctx context.Context, userID string) error {
	cmd := &command.DeleteUserCommand{
		BaseCommand: command.NewBaseCommand("user.delete"),
		UserID:      userID,
	}
	
	return h.commandBus.Dispatch(ctx, cmd)
}
`

// DDD feature templates
const DDDSQSHandler = `package lambda

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"{{.Module}}/application/handler"
	"{{.Module}}/domain/event"
	"{{.Module}}/infrastructure/config"
	"github.com/rs/zerolog/log"
)

// SQSHandler handles SQS messages
type SQSHandler struct {
	messageHandler *handler.MessageHandler
}

// NewSQSHandler creates a new SQS handler
func NewSQSHandler(messageHandler *handler.MessageHandler) *SQSHandler {
	return &SQSHandler{
		messageHandler: messageHandler,
	}
}

// HandleRequest processes SQS events
func (h *SQSHandler) HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
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

func (h *SQSHandler) processMessage(ctx context.Context, record events.SQSMessage) error {
	// Parse message body
	var message struct {
		Type    string          ` + "`json:\"type\"`" + `
		Payload json.RawMessage ` + "`json:\"payload\"`" + `
	}
	
	if err := json.Unmarshal([]byte(record.Body), &message); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed to parse message")
		// Don't retry invalid messages
		return nil
	}
	
	log.Ctx(ctx).Info().
		Str("message_id", record.MessageId).
		Str("type", message.Type).
		Msg("Processing message")
	
	// Process based on message type
	return h.messageHandler.Handle(ctx, message.Type, message.Payload)
}

// Start starts the Lambda handler
func Start() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
	
	// Initialize infrastructure
	infra, err := infrastructure.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize infrastructure")
	}
	
	// Create message handler
	messageHandler := handler.NewMessageHandler(infra.EventBus())
	
	// Register event handlers
	messageHandler.RegisterHandler("user.created", handler.NewUserCreatedHandler(infra.UserRepository()))
	messageHandler.RegisterHandler("user.updated", handler.NewUserUpdatedHandler(infra.UserRepository()))
	// Register other handlers...
	
	// Create SQS handler
	sqsHandler := NewSQSHandler(messageHandler)
	
	// Start Lambda
	lambda.Start(sqsHandler.HandleRequest)
}
`

const DDDSQSClient = `package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"{{.Module}}/domain/event"
	"github.com/rs/zerolog/log"
)

// SQSClient handles SQS operations
type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

// NewSQSClient creates a new SQS client
func NewSQSClient(client *sqs.Client, queueURL string) *SQSClient {
	return &SQSClient{
		client:   client,
		queueURL: queueURL,
	}
}

// PublishEvent publishes a domain event to SQS
func (c *SQSClient) PublishEvent(ctx context.Context, event event.DomainEvent) error {
	// Create message
	message := map[string]interface{}{
		"type":         event.EventType(),
		"aggregate_id": event.AggregateID(),
		"occurred_at":  event.OccurredAt(),
		"payload":      event,
	}
	
	// Marshal message
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	
	// Send to SQS
	_, err = c.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"event_type": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.EventType()),
			},
			"aggregate_id": {
				DataType:    aws.String("String"),
				StringValue: aws.String(event.AggregateID()),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	
	log.Ctx(ctx).Info().
		Str("event_type", event.EventType()).
		Str("aggregate_id", event.AggregateID()).
		Msg("Event published to SQS")
	
	return nil
}

// PublishEvents publishes multiple domain events
func (c *SQSClient) PublishEvents(ctx context.Context, events []event.DomainEvent) error {
	for _, evt := range events {
		if err := c.PublishEvent(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}
`

const DDDMessageHandler = `package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"{{.Module}}/domain/event"
	"github.com/rs/zerolog/log"
)

// MessageHandler handles different message types
type MessageHandler struct {
	handlers map[string]MessageHandlerFunc
	eventBus event.EventBus
}

// MessageHandlerFunc is a function that handles a message
type MessageHandlerFunc func(ctx context.Context, payload json.RawMessage) error

// NewMessageHandler creates a new message handler
func NewMessageHandler(eventBus event.EventBus) *MessageHandler {
	return &MessageHandler{
		handlers: make(map[string]MessageHandlerFunc),
		eventBus: eventBus,
	}
}

// RegisterHandler registers a handler for a message type
func (h *MessageHandler) RegisterHandler(messageType string, handler MessageHandlerFunc) {
	h.handlers[messageType] = handler
}

// Handle handles a message
func (h *MessageHandler) Handle(ctx context.Context, messageType string, payload json.RawMessage) error {
	handler, exists := h.handlers[messageType]
	if !exists {
		log.Ctx(ctx).Warn().
			Str("type", messageType).
			Msg("No handler registered for message type")
		return nil
	}
	
	return handler(ctx, payload)
}

// UserCreatedHandler handles user created events
type UserCreatedHandler struct {
	userRepo repository.UserRepository
}

// NewUserCreatedHandler creates a new user created handler
func NewUserCreatedHandler(userRepo repository.UserRepository) MessageHandlerFunc {
	h := &UserCreatedHandler{userRepo: userRepo}
	return h.Handle
}

// Handle handles the user created event
func (h *UserCreatedHandler) Handle(ctx context.Context, payload json.RawMessage) error {
	var evt event.UserCreated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}
	
	log.Ctx(ctx).Info().
		Str("user_id", evt.UserID).
		Str("email", evt.Email).
		Msg("Processing user created event")
	
	// Add your business logic here
	// For example: Send welcome email, create default profile, etc.
	
	return nil
}

// UserUpdatedHandler handles user updated events
type UserUpdatedHandler struct {
	userRepo repository.UserRepository
}

// NewUserUpdatedHandler creates a new user updated handler
func NewUserUpdatedHandler(userRepo repository.UserRepository) MessageHandlerFunc {
	h := &UserUpdatedHandler{userRepo: userRepo}
	return h.Handle
}

// Handle handles the user updated event
func (h *UserUpdatedHandler) Handle(ctx context.Context, payload json.RawMessage) error {
	var evt event.UserProfileUpdated
	if err := json.Unmarshal(payload, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}
	
	log.Ctx(ctx).Info().
		Str("user_id", evt.UserID).
		Str("old_name", evt.OldName).
		Str("new_name", evt.NewName).
		Msg("Processing user profile updated event")
	
	// Add your business logic here
	// For example: Update search index, notify other services, etc.
	
	return nil
}
`

const DDDDynamoDBRepository = `package repository

import (
	"context"

	"{{.Module}}/domain/aggregate"
	"{{.Module}}/domain/repository"
	"{{.Module}}/infrastructure/persistence"
)

// userRepository implements the domain UserRepository interface
type userRepository struct {
	dynamoRepo *persistence.DynamoDBUserRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(dynamoRepo *persistence.DynamoDBUserRepository) repository.UserRepository {
	return &userRepository{
		dynamoRepo: dynamoRepo,
	}
}

// Save delegates to the DynamoDB repository
func (r *userRepository) Save(ctx context.Context, user *aggregate.User) error {
	return r.dynamoRepo.Save(ctx, user)
}

// FindByID delegates to the DynamoDB repository
func (r *userRepository) FindByID(ctx context.Context, id string) (*aggregate.User, error) {
	return r.dynamoRepo.FindByID(ctx, id)
}

// FindByEmail delegates to the DynamoDB repository
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*aggregate.User, error) {
	return r.dynamoRepo.FindByEmail(ctx, email)
}

// List delegates to the DynamoDB repository
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*aggregate.User, error) {
	return r.dynamoRepo.List(ctx, offset, limit)
}

// Count delegates to the DynamoDB repository
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.dynamoRepo.Count(ctx)
}

// Delete delegates to the DynamoDB repository
func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.dynamoRepo.Delete(ctx, id)
}
`

const DDDUserRepository = `package repository

// This file is generated by DDDRepository template
// It provides the concrete implementation of UserRepository for the infrastructure layer
`