# CQRS and Domain-Driven Design in Golang

This repository demonstrates the implementation of **CQRS (Command Query Responsibility Segregation)** and **Domain-Driven Design (DDD)** patterns using Go. These architectural patterns help build scalable, maintainable, and testable applications by separating concerns and focusing on business logic.

## Table of Contents

- [Overview](#overview)
- [Domain-Driven Design (DDD)](#domain-driven-design-ddd)
- [CQRS (Command Query Responsibility Segregation)](#cqrs-command-query-responsibility-segregation)
- [Change Data Capture (CDC)](#change-data-capture-cdc)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Benefits](#benefits)
- [Contributing](#contributing)

## Overview

### What is Domain-Driven Design (DDD)?

Domain-Driven Design is a software development approach that focuses on creating a rich domain model that reflects the business logic and rules. It emphasizes:

- **Ubiquitous Language**: A common vocabulary between developers and domain experts
- **Bounded Contexts**: Clear boundaries between different parts of the system
- **Domain Models**: Rich objects that encapsulate business logic
- **Strategic Design**: High-level architectural decisions

### What is CQRS?

Command Query Responsibility Segregation (CQRS) separates read and write operations into **completely separate services**:

- **Command Service**: Handles write operations and state changes
- **Query Service**: Handles read operations and data retrieval
- **Event Bus**: Enables communication between services through domain events
- **Eventual Consistency**: Query service is updated asynchronously via events
- **Outbox Pattern**: Ensures reliable event publishing from command service

## Domain-Driven Design (DDD)

### Core DDD Concepts in Golang

#### 1. Entities

Entities are objects that have a unique identity and lifecycle. They are mutable and can change over time.

```go
type User struct {
    ID       UserID
    Email    string
    Name     string
    createdAt time.Time
}

type UserID string

func (u *User) ChangeEmail(newEmail string) error {
    if !isValidEmail(newEmail) {
        return errors.New("invalid email format")
    }
    u.Email = newEmail
    return nil
}
```

#### 2. Value Objects

Value objects are immutable objects defined by their attributes rather than identity.

```go
type Money struct {
    amount   int64
    currency string
}

func NewMoney(amount int64, currency string) Money {
    return Money{amount: amount, currency: currency}
}

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, errors.New("cannot add different currencies")
    }
    return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}
```

#### 3. Aggregates

Aggregates are clusters of related entities and value objects that are treated as a single unit. They ensure consistency and enforce business rules.

```go
type Order struct {
    ID       OrderID
    Customer Customer
    Items    []OrderItem
    Status   OrderStatus
    Total    Money
}

type OrderID string

func (o *Order) AddItem(product Product, quantity int) error {
    if o.Status != OrderStatusDraft {
        return errors.New("cannot modify completed order")
    }
    
    item := OrderItem{
        Product:  product,
        Quantity: quantity,
        Price:    product.Price,
    }
    o.Items = append(o.Items, item)
    o.recalculateTotal()
    return nil
}

func (o *Order) recalculateTotal() {
    total := int64(0)
    for _, item := range o.Items {
        total += item.Price.Amount * int64(item.Quantity)
    }
    o.Total = Money{amount: total, currency: "USD"}
}
```

#### 4. Domain Services

Domain services contain business logic that doesn't naturally fit into entities or value objects.

```go
type PricingService struct {
    discountRepository DiscountRepository
}

func (ps *PricingService) CalculateDiscountedPrice(order *Order, customer *Customer) (Money, error) {
    basePrice := order.Total
    
    // Apply customer-specific discounts
    discounts, err := ps.discountRepository.GetActiveDiscounts(customer.ID)
    if err != nil {
        return Money{}, err
    }
    
    discountedAmount := basePrice.Amount
    for _, discount := range discounts {
        if discount.IsApplicable(order) {
            discountedAmount = discount.Apply(discountedAmount)
        }
    }
    
    return Money{amount: discountedAmount, currency: basePrice.Currency}, nil
}
```

#### 5. Repositories

Repositories abstract data access and provide a domain-focused interface.

```go
type UserRepository interface {
    Save(user *User) error
    FindByID(id UserID) (*User, error)
    FindByEmail(email string) (*User, error)
    Delete(id UserID) error
}

type userRepository struct {
    db *sql.DB
}

func (r *userRepository) Save(user *User) error {
    query := `INSERT INTO users (id, email, name, created_at) VALUES (?, ?, ?, ?)`
    _, err := r.db.Exec(query, user.ID, user.Email, user.Name, user.CreatedAt)
    return err
}

func (r *userRepository) FindByID(id UserID) (*User, error) {
    query := `SELECT id, email, name, created_at FROM users WHERE id = ?`
    row := r.db.QueryRow(query, id)
    
    var user User
    err := row.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

## CQRS (Command Query Responsibility Segregation)

### Architecture Overview

True CQRS separates Command and Query operations into **completely separate services**:

```
┌─────────────────┐    Events    ┌─────────────────┐
│  Command Service │ ──────────► │   Query Service │
│                 │              │                 │
│ - Write Models  │              │ - Read Models   │
│ - Domain Logic  │              │ - Optimized     │
│ - Event Store   │              │   Queries       │
│ - Outbox        │              │ - Projections   │
└─────────────────┘              └─────────────────┘
```

### Command Service (Write Operations)

The Command Service handles all write operations and publishes domain events.

```go
// Command Service - handles writes
type CommandService struct {
    userRepo    UserRepository
    eventStore  EventStore
    outbox      OutboxRepository
}

type CreateUserCommand struct {
    Email string `json:"email"`
    Name  string `json:"name"`
}

type CreateUserHandler struct {
    service *CommandService
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) error {
    // Validate command
    if err := h.validateCommand(cmd); err != nil {
        return err
    }
    
    // Create user aggregate
    user := &User{
        ID:    UserID(uuid.New().String()),
        Email: cmd.Email,
        Name:  cmd.Name,
        CreatedAt: time.Now(),
    }
    
    // Save aggregate to event store
    events := []DomainEvent{
        UserCreatedEvent{
            UserID:    user.ID,
            Email:     user.Email,
            Name:      user.Name,
            CreatedAt: user.CreatedAt,
        },
    }
    
    if err := h.service.eventStore.SaveEvents(ctx, string(user.ID), events, 0); err != nil {
        return err
    }
    
    // Save to outbox for reliable event publishing
    return h.service.outbox.SaveEvent(ctx, events[0])
}

func (h *CreateUserHandler) validateCommand(cmd CreateUserCommand) error {
    if cmd.Email == "" {
        return errors.New("email is required")
    }
    if cmd.Name == "" {
        return errors.New("name is required")
    }
    return nil
}
```

### Outbox Pattern Implementation

The Outbox Pattern ensures reliable event publishing from the Command Service.

```go
type OutboxRepository interface {
    SaveEvent(ctx context.Context, event DomainEvent) error
    GetUnprocessedEvents(ctx context.Context) ([]OutboxEvent, error)
    MarkAsProcessed(ctx context.Context, eventID string) error
}

type OutboxEvent struct {
    ID        string    `json:"id"`
    EventType string    `json:"event_type"`
    EventData []byte    `json:"event_data"`
    CreatedAt time.Time `json:"created_at"`
    Processed bool      `json:"processed"`
}

type outboxRepository struct {
    db *sql.DB
}

func (r *outboxRepository) SaveEvent(ctx context.Context, event DomainEvent) error {
    eventData, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    query := `INSERT INTO outbox_events (id, event_type, event_data, created_at) VALUES (?, ?, ?, ?)`
    _, err = r.db.ExecContext(ctx, query, uuid.New().String(), event.Type(), eventData, time.Now())
    return err
}

func (r *outboxRepository) GetUnprocessedEvents(ctx context.Context) ([]OutboxEvent, error) {
    query := `SELECT id, event_type, event_data, created_at FROM outbox_events WHERE processed = false ORDER BY created_at`
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var events []OutboxEvent
    for rows.Next() {
        var event OutboxEvent
        err := rows.Scan(&event.ID, &event.EventType, &event.EventData, &event.CreatedAt)
        if err != nil {
            return nil, err
        }
        events = append(events, event)
    }
    
    return events, nil
}

// Event Publisher - runs as background process
type EventPublisher struct {
    outboxRepo OutboxRepository
    eventBus   EventBus
}

func (ep *EventPublisher) ProcessEvents(ctx context.Context) error {
    events, err := ep.outboxRepo.GetUnprocessedEvents(ctx)
    if err != nil {
        return err
    }
    
    for _, event := range events {
        if err := ep.eventBus.Publish(ctx, event); err != nil {
            return err
        }
        
        if err := ep.outboxRepo.MarkAsProcessed(ctx, event.ID); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Query Service (Read Operations)

The Query Service handles all read operations and maintains read models through event projections.

```go
// Query Service - handles reads
type QueryService struct {
    readModel UserReadModel
}

type GetUserQuery struct {
    UserID string `json:"user_id"`
}

type GetUserHandler struct {
    service *QueryService
}

func (h *GetUserHandler) Handle(ctx context.Context, query GetUserQuery) (*UserDTO, error) {
    user, err := h.service.readModel.GetUser(ctx, query.UserID)
    if err != nil {
        return nil, err
    }
    
    return &UserDTO{
        ID:        user.ID,
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: user.CreatedAt,
    }, nil
}

type UserDTO struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}
```

### Event Projections

The Query Service updates its read models through event projections.

```go
type UserReadModel interface {
    GetUser(ctx context.Context, userID string) (*UserReadModel, error)
    CreateUser(ctx context.Context, user *UserReadModel) error
    UpdateUserEmail(ctx context.Context, userID, email string) error
}

type UserReadModel struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

type UserProjectionHandler struct {
    readModel UserReadModel
}

func (h *UserProjectionHandler) Handle(ctx context.Context, event DomainEvent) error {
    switch e := event.(type) {
    case UserCreatedEvent:
        return h.readModel.CreateUser(ctx, &UserReadModel{
            ID:        string(e.UserID),
            Email:     e.Email,
            Name:      e.Name,
            CreatedAt: e.CreatedAt,
        })
    case UserEmailChangedEvent:
        return h.readModel.UpdateUserEmail(ctx, string(e.UserID), e.NewEmail)
    }
    return nil
}

// Event Consumer - runs in Query Service
type EventConsumer struct {
    projectionHandler *UserProjectionHandler
    eventBus         EventBus
}

func (ec *EventConsumer) Start(ctx context.Context) error {
    return ec.eventBus.Subscribe(ctx, "user-events", ec.handleEvent)
}

func (ec *EventConsumer) handleEvent(ctx context.Context, event DomainEvent) error {
    return ec.projectionHandler.Handle(ctx, event)
}
```

### Event Sourcing Integration

Event Sourcing can be combined with CQRS to store events instead of current state.

```go
type DomainEvent interface {
    Type() string
    AggregateID() string
    OccurredAt() time.Time
}

type UserCreatedEvent struct {
    UserID    UserID
    Email     string
    Name      string
    CreatedAt time.Time
}

func (e UserCreatedEvent) Type() string {
    return "UserCreated"
}

func (e UserCreatedEvent) AggregateID() string {
    return string(e.UserID)
}

func (e UserCreatedEvent) OccurredAt() time.Time {
    return e.CreatedAt
}

type EventStore interface {
    SaveEvents(aggregateID string, events []DomainEvent, expectedVersion int) error
    GetEvents(aggregateID string) ([]DomainEvent, error)
}
```

### Projection Handlers

Projections update read models based on domain events.

```go
type ProjectionHandler interface {
    Handle(event DomainEvent) error
}

type UserProjectionHandler struct {
    readModel UserReadModel
}

func (h *UserProjectionHandler) Handle(event DomainEvent) error {
    switch e := event.(type) {
    case UserCreatedEvent:
        return h.readModel.CreateUser(UserReadModel{
            ID:    string(e.UserID),
            Email: e.Email,
            Name:  e.Name,
            CreatedAt: e.CreatedAt,
        })
    case UserEmailChangedEvent:
        return h.readModel.UpdateUserEmail(string(e.UserID), e.NewEmail)
    }
    return nil
}
```

## Change Data Capture (CDC)

CDC is an alternative to the traditional Outbox pattern for achieving eventual consistency between the Command and Query sides. Instead of writing domain events explicitly to an outbox table, CDC taps into the database's transaction log (e.g., PostgreSQL WAL) to emit change events (insert/update/delete) to a streaming platform like Kafka. Your Query service consumes these change events to update read models (projections).

### When to use CDC vs Outbox

- **Use CDC when:**
  - You want minimal application code changes in the Command service.
  - Your database supports logical decoding and reliable log-based capture.
  - You need to capture any table changes, including those from legacy systems or multiple writers.
- **Use Outbox when:**
  - You prefer explicit domain events and tighter control of the event schema.
  - You need exactly-once semantics with application-level idempotency and ordering.
  - You want to avoid tight coupling to specific database features.

### Architecture (CDC-based CQRS)

```mermaid
flowchart LR
  A[Command Service<br/>PostgreSQL (WAL)] -- WAL --> B[Kafka Connect + Debezium<br/>(CDC)]
  B -- topics:user.public.users --> C[Kafka]
  C -- consume --> D[Query Service<br/>Projections/Read Models]
```

### Docker Compose additions (Kafka Connect with Debezium)

Below is an example of adding Kafka Connect with the Debezium PostgreSQL connector to the existing `docker-compose.yml`. Note: enabling logical decoding on Postgres may require custom configuration (e.g., `wal_level=logical`).

```yaml
# Add to docker-compose.yml
  kafka-connect:
    image: confluentinc/cp-kafka-connect:7.4.0
    depends_on:
      kafka:
        condition: service_healthy
      zookeeper:
        condition: service_started
    ports:
      - "8083:8083"
    environment:
      CONNECT_BOOTSTRAP_SERVERS: "kafka:9092"
      CONNECT_REST_ADVERTISED_HOST_NAME: kafka-connect
      CONNECT_REST_PORT: 8083
      CONNECT_GROUP_ID: "cdc-connect-group"
      CONNECT_CONFIG_STORAGE_TOPIC: "_connect-configs"
      CONNECT_OFFSET_STORAGE_TOPIC: "_connect-offsets"
      CONNECT_STATUS_STORAGE_TOPIC: "_connect-status"
      CONNECT_KEY_CONVERTER: "org.apache.kafka.connect.json.JsonConverter"
      CONNECT_VALUE_CONVERTER: "org.apache.kafka.connect.json.JsonConverter"
      CONNECT_INTERNAL_KEY_CONVERTER: "org.apache.kafka.connect.json.JsonConverter"
      CONNECT_INTERNAL_VALUE_CONVERTER: "org.apache.kafka.connect.json.JsonConverter"
      CONNECT_PLUGIN_PATH: "/usr/share/java,/etc/kafka-connect/jars"
      # Optional: reduce noisy SMT errors during dev
      CONNECT_LOG4J_ROOT_LOGLEVEL: "INFO"
    volumes:
      - ./connect-jars:/etc/kafka-connect/jars
```

To enable logical decoding on PostgreSQL (development), you can use a custom image or command flags. Example:

```yaml
  postgres:
    image: debezium/postgres:16
    environment:
      POSTGRES_DB: ecommerce_orders
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
      # Debezium image already sets wal_level=logical, max_wal_senders, etc.
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

### Sample Debezium PostgreSQL connector

Create the connector by POSTing this JSON to `http://localhost:8083/connectors`:

```json
{
  "name": "ecommerce-orders-postgres-cdc",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.hostname": "postgres",
    "database.port": "5432",
    "database.user": "postgres",
    "database.password": "password",
    "database.dbname": "ecommerce_orders",
    "topic.prefix": "ecommerce",
    "slot.name": "ecommerce_slot",
    "publication.autocreate.mode": "filtered",
    "plugin.name": "pgoutput",
    "tombstones.on.delete": "false",
    "include.schema.changes": "false",
    "decimal.handling.mode": "double",
    "time.precision.mode": "adaptive_time_microseconds",
    "snapshot.mode": "initial",
    "schema.include.list": "public",
    "table.include.list": "public.orders,public.order_items,public.customers"
  }
}
```

This will produce topics like `ecommerce.public.orders`, with messages (create/update/delete) carrying `before`/`after` rows and metadata.

### Consuming CDC events in the Query Service

- **Projection design:** Map table-oriented CDC events to your read models. For example, when consuming from `ecommerce.public.orders`, translate `after` payloads into `OrderReadModel` upserts, and map deletes to tombstoning.
- **Idempotency:** Debezium provides keys and source metadata (`lsn`, `ts_ms`). Use these to implement idempotent upserts in the read store. Store the last processed LSN per topic/partition to avoid duplicates on replays.
- **Ordering and retries:** Process per-key ordering; use partitioning by aggregate ID if possible. Implement retry with backoff and dead-letter topics for poison messages.
- **Enrichment:** If you still prefer domain semantics, you can enrich CDC events within the consumer to emit domain-level projections.

### CDC vs Outbox: trade-offs

- **CDC pros:** Minimal app changes, great for legacy systems, can capture all writes. Operates at DB boundary with strong ordering guarantees per table.
- **CDC cons:** Events are table-centric (not domain-centric), schema drift needs management, dependency on DB-specific features, snapshot behavior must be planned.
- **Outbox pros:** Domain-event-centric, explicit schemas, well-aligned with DDD aggregates, easier evolution of event contracts.
- **Outbox cons:** Requires app code to write outbox and a background publisher, dual-writes to outbox and main tables (usually within one TX).

---

## Project Structure with Go Workspace

Using Go workspaces allows us to manage multiple related modules while keeping services separate and independently deployable.

### Example: E-commerce Order Management System

```
ecommerce-orders/
├── go.work                    # Go workspace file
├── docker-compose.yml
├── README.md
├── shared/                    # Shared domain module
│   ├── go.mod
│   ├── go.sum
│   ├── domain/
│   │   ├── entities/
│   │   │   ├── order.go
│   │   │   ├── order_item.go
│   │   │   └── customer.go
│   │   ├── valueobjects/
│   │   │   ├── money.go
│   │   │   ├── address.go
│   │   │   └── order_status.go
│   │   ├── aggregates/
│   │   │   └── order_aggregate.go
│   │   ├── services/
│   │   │   ├── pricing_service.go
│   │   │   └── inventory_service.go
│   │   └── events/
│   │       ├── order_created.go
│   │       ├── order_updated.go
│   │       └── order_cancelled.go
│   └── infrastructure/
│       ├── eventbus/
│       └── uuid/
├── order-management-service/   # Command service module
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── handlers/
│   │   │   ├── create_order_handler.go
│   │   │   ├── update_order_handler.go
│   │   │   └── cancel_order_handler.go
│   │   ├── repositories/
│   │   │   ├── order_repository.go
│   │   │   └── outbox_repository.go
│   │   └── interfaces/
│   │       └── http/
│   │           └── handlers.go
│   └── Dockerfile
└── order-reporting-service/    # Query service module
    ├── go.mod
    ├── go.sum
    ├── cmd/
    │   └── main.go
    ├── internal/
    │   ├── handlers/
    │   │   ├── get_order_handler.go
    │   │   ├── list_orders_handler.go
    │   │   └── order_analytics_handler.go
    │   ├── projections/
    │   │   ├── order_projection.go
    │   │   └── customer_projection.go
    │   ├── readmodels/
    │   │   ├── order_read_model.go
    │   │   └── analytics_read_model.go
    │   └── interfaces/
    │       └── http/
    │           └── handlers.go
    └── Dockerfile
```

### Go Workspace Configuration

```go
// go.work
go 1.25

use (
    ./shared
    ./order-management-service
    ./order-reporting-service
)
```

### Module Dependencies

Each service module depends on the shared domain module:

```go
// order-management-service/go.mod
module github.com/vdntruong/dddcqrs/order-management-service

go 1.25

require (
    github.com/vdntruong/dddcqrs/shared v0.0.0
    github.com/gorilla/mux v1.8.0
    github.com/lib/pq v1.10.9
    github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
    github.com/redis/go-redis/v9 v9.3.0
)

replace github.com/vdntruong/dddcqrs/shared => ../shared
```

```go
// order-reporting-service/go.mod
module github.com/vdntruong/dddcqrs/order-reporting-service

go 1.25

require (
    github.com/vdntruong/dddcqrs/shared v0.0.0
    github.com/gorilla/mux v1.8.0
    github.com/lib/pq v1.10.9
    github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
    github.com/redis/go-redis/v9 v9.3.0
)

replace github.com/vdntruong/dddcqrs/shared => ../shared
```

```go
// shared/go.mod
module github.com/vdntruong/dddcqrs/shared

go 1.25

require (
    github.com/google/uuid v1.6.0
    github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
    github.com/redis/go-redis/v9 v9.3.0
)
```

## Getting Started

### Prerequisites

- Go 1.25 (latest version)
- Docker (for PostgreSQL, Kafka, Redis)
- Make (optional)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/vdntruong/dddcqrs.git
cd dddcqrs
```

2. Initialize the Go workspace:
```bash
go work init
go work use ./shared ./order-management-service ./order-reporting-service
```

3. Install dependencies for all modules:
```bash
go mod download
```

4. Start the infrastructure:
```bash
docker-compose up -d
```

5. Run the services:

**Order Management Service (Command Side):**
```bash
cd order-management-service
go run cmd/main.go
```

**Order Reporting Service (Query Side):**
```bash
cd order-reporting-service
go run cmd/main.go
```

### Development Workflow

**Working with the workspace:**
```bash
# Run tests for all modules
go test ./...

# Build all services
go build ./order-management-service/cmd
go build ./order-reporting-service/cmd

# Run specific service tests
go test ./order-management-service/...
go test ./order-reporting-service/...
go test ./shared/...

# Add new dependency to a specific module
cd order-management-service
go get github.com/new-dependency

# Update workspace after adding dependencies
go work sync
```

### Running Tests

```bash
# Test all modules in workspace
go test ./...

# Test specific module
go test ./shared/...
go test ./order-management-service/...
go test ./order-reporting-service/...
```

### Code Quality and Linting

The project uses `golangci-lint` for code quality checks and `pre-commit` hooks for automated code formatting.

```bash
# Run linting on all modules
make lint

# Run linting on specific modules
make lint-shared
make lint-command
make lint-query

# Auto-fix linting issues
make lint-fix

# Install pre-commit hooks
pre-commit install

# Run pre-commit hooks manually
pre-commit run --all-files
```

### Development Setup

For a complete development environment setup:

```bash
# Run the setup script
./scripts/setup-dev.sh

# Or manually setup
make dev-setup
pre-commit install
```

### CI/CD Pipeline

The project includes GitHub Actions workflows for continuous integration:

- **Code Quality**: Runs `golangci-lint` on all modules
- **Testing**: Executes unit tests with PostgreSQL and Redis services
- **Building**: Builds all services and Docker images
- **Docker**: Tests Docker image builds

The CI pipeline runs on:
- Push to `main` and `develop` branches
- Pull requests to `main` and `develop` branches

### Pre-commit Hooks

Pre-commit hooks ensure code quality before commits:

- **Trailing whitespace removal**
- **End-of-file fixing**
- **YAML validation**
- **Large file detection**
- **Merge conflict detection**
- **Go formatting** (`go fmt`)
- **Go vetting** (`go vet`)
- **Go imports** (`goimports`)
- **Go mod tidy**
- **golangci-lint** execution

## Examples

### Why E-commerce Order Management?

This example is perfect for learning CQRS because it demonstrates:

**Clear Command/Query Separation:**
- **Commands**: Create Order, Update Order, Cancel Order, Add Items
- **Queries**: Get Order Details, List Orders, Order Analytics, Customer History

**Real Business Complexity:**
- **Domain Events**: OrderCreated, OrderUpdated, OrderCancelled, PaymentProcessed
- **Business Rules**: Inventory checks, pricing calculations, shipping rules
- **Read Models**: Order summaries, analytics dashboards, customer profiles

**Scalability Needs:**
- **High Read Volume**: Order lookups, analytics, reporting
- **Moderate Write Volume**: Order creation and updates
- **Different Performance Requirements**: Fast reads vs complex business logic

### Complete CQRS Implementation with Separate Services

Here's a complete example showing how to implement CQRS with separate Command and Query services:

#### Order Management Service Main

```go
// order-management-service/cmd/main.go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/gorilla/mux"
    "github.com/vdntruong/dddcqrs/shared/infrastructure/eventbus"
    "github.com/vdntruong/dddcqrs/order-management-service/internal/handlers"
    "github.com/vdntruong/dddcqrs/order-management-service/internal/repositories"
)

func main() {
    // Initialize dependencies
    db := initDatabase()
    eventBus := eventbus.NewKafkaEventBus()
    
    // Initialize repositories
    orderRepo := repositories.NewOrderRepository(db)
    outboxRepo := repositories.NewOutboxRepository(db)
    eventStore := repositories.NewEventStore(db)
    
    // Initialize command service
    commandService := &handlers.CommandService{
        OrderRepo:  orderRepo,
        EventStore: eventStore,
        Outbox:     outboxRepo,
    }
    
    // Initialize command handlers
    createOrderHandler := &handlers.CreateOrderHandler{
        Service: commandService,
    }
    
    // Initialize HTTP router
    router := mux.NewRouter()
    router.HandleFunc("/orders", createOrderHandler.HandleHTTP).Methods("POST")
    
    // Start event publisher (background process)
    eventPublisher := &handlers.EventPublisher{
        OutboxRepo: outboxRepo,
        EventBus:   eventBus,
    }
    
    go func() {
        if err := eventPublisher.ProcessEvents(context.Background()); err != nil {
            log.Printf("Event publisher error: %v", err)
        }
    }()
    
    // Start HTTP server
    log.Println("Order Management Service starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

#### Order Reporting Service Main

```go
// order-reporting-service/cmd/main.go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/gorilla/mux"
    "github.com/vdntruong/dddcqrs/shared/infrastructure/eventbus"
    "github.com/vdntruong/dddcqrs/order-reporting-service/internal/handlers"
    "github.com/vdntruong/dddcqrs/order-reporting-service/internal/readmodels"
)

func main() {
    // Initialize dependencies
    db := initDatabase()
    redisClient := initRedis()
    eventBus := eventbus.NewKafkaEventBus()
    
    // Initialize read models
    orderReadModel := readmodels.NewOrderReadModel(db, redisClient)
    
    // Initialize projection handlers
    orderProjectionHandler := &handlers.OrderProjectionHandler{
        ReadModel: orderReadModel,
    }
    
    // Initialize query handlers
    getOrderHandler := &handlers.GetOrderHandler{
        ReadModel: orderReadModel,
    }
    
    // Initialize HTTP router
    router := mux.NewRouter()
    router.HandleFunc("/orders/{id}", getOrderHandler.HandleHTTP).Methods("GET")
    
    // Start event consumer (background process)
    eventConsumer := &handlers.EventConsumer{
        ProjectionHandler: orderProjectionHandler,
        EventBus:         eventBus,
    }
    
    go func() {
        if err := eventConsumer.Start(context.Background()); err != nil {
            log.Printf("Event consumer error: %v", err)
        }
    }()
    
    // Start HTTP server
    log.Println("Order Reporting Service starting on :8081")
    log.Fatal(http.ListenAndServe(":8081", router))
}
```

#### Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  # PostgreSQL for Command Side (Write Operations)
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: ecommerce_orders
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Kafka for Event Streaming
  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    volumes:
      - zookeeper_data:/var/lib/zookeeper/data

  kafka:
    image: confluentinc/cp-kafka:7.4.0
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: true
    volumes:
      - kafka_data:/var/lib/kafka/data
    healthcheck:
      test: ["CMD", "kafka-broker-api-versions", "--bootstrap-server", "localhost:9092"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis for Caching (Query Side)
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Order Management Service (Command Side)
  order-management-service:
    build:
      context: .
      dockerfile: order-management-service/Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      kafka:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://postgres:password@postgres:5432/ecommerce_orders?sslmode=disable
      KAFKA_BROKERS: kafka:9092
      KAFKA_TOPIC_ORDERS: orders
      LOG_LEVEL: info
    restart: unless-stopped

  # Order Reporting Service (Query Side)
  order-reporting-service:
    build:
      context: .
      dockerfile: order-reporting-service/Dockerfile
    ports:
      - "8081:8081"
    depends_on:
      postgres:
        condition: service_healthy
      kafka:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DATABASE_URL: postgres://postgres:password@postgres:5432/ecommerce_orders?sslmode=disable
      KAFKA_BROKERS: kafka:9092
      KAFKA_TOPIC_ORDERS: orders
      REDIS_URL: redis://redis:6379
      LOG_LEVEL: info
    restart: unless-stopped

volumes:
  postgres_data:
  kafka_data:
  zookeeper_data:
  redis_data:
```

#### Dockerfile Examples

**Order Management Service Dockerfile:**
```dockerfile
# order-management-service/Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work go.work.sum ./
COPY shared/ ./shared/
COPY order-management-service/ ./order-management-service/

# Build the order management service
WORKDIR /app/order-management-service
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/order-management-service/main .

CMD ["./main"]
```

**Order Reporting Service Dockerfile:**
```dockerfile
# order-reporting-service/Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy workspace files
COPY go.work go.work.sum ./
COPY shared/ ./shared/
COPY order-reporting-service/ ./order-reporting-service/

# Build the order reporting service
WORKDIR /app/order-reporting-service
RUN go build -o main cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/order-reporting-service/main .

CMD ["./main"]
```

#### Usage Example

```bash
# Start services
docker-compose up -d

# Create an order (Order Management Service)
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": "customer-123",
    "items": [
      {"product_id": "product-456", "quantity": 2, "price": 29.99}
    ],
    "shipping_address": {
      "street": "123 Main St",
      "city": "New York",
      "state": "NY",
      "zip": "10001"
    }
  }'

# Get order details (Order Reporting Service) - Note: eventual consistency!
curl http://localhost:8081/orders/{order-id}

# Get order analytics (Order Reporting Service)
curl http://localhost:8081/analytics/orders?period=monthly
```

### Technology Stack

**Command Side (Order Management Service):**
- **Database**: PostgreSQL 16 (ACID transactions, strong consistency)
- **Event Streaming**: Apache Kafka (reliable event delivery)
- **Language**: Go 1.25 (latest version)
- **Framework**: Gorilla Mux (HTTP routing)

**Query Side (Order Reporting Service):**
- **Database**: PostgreSQL 16 (read-optimized queries)
- **Caching**: Redis 7 (fast data access)
- **Event Streaming**: Apache Kafka (event consumption)
- **Language**: Go 1.25 (latest version)
- **Framework**: Gorilla Mux (HTTP routing)

**Infrastructure:**
- **Containerization**: Docker & Docker Compose
- **Orchestration**: Health checks and service dependencies
- **Monitoring**: Structured logging and health endpoints

### Eventual Consistency Flow

1. **Order Management Service** receives create order request
2. **Order Management Service** saves order aggregate to PostgreSQL
3. **Order Management Service** saves event to outbox table (same transaction)
4. **Event Publisher** (background process) reads from outbox
5. **Event Publisher** publishes event to Kafka topic
6. **Order Reporting Service** consumes event from Kafka
7. **Order Reporting Service** updates read model via projection
8. **Order Reporting Service** caches data in Redis for fast access
9. **Order Reporting Service** can now serve the order data

### The Outbox Pattern

The Outbox Pattern is crucial for reliable event publishing in CQRS:

```go
// Transaction ensures both aggregate and event are saved atomically
func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUserCommand) error {
    tx, err := h.service.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Save aggregate
    user := &User{...}
    if err := h.service.userRepo.SaveWithTx(ctx, tx, user); err != nil {
        return err
    }
    
    // Save event to outbox in same transaction
    event := UserCreatedEvent{...}
    if err := h.service.outbox.SaveEventWithTx(ctx, tx, event); err != nil {
        return err
    }
    
    return tx.Commit() // Both operations succeed or both fail
}
```

**Why Outbox Pattern?**
- **Atomicity**: Aggregate and event are saved in the same transaction
- **Reliability**: Events are never lost, even if Kafka is down
- **Consistency**: Ensures eventual consistency between services
- **Recovery**: Failed events can be retried automatically

### Kafka Event Bus Implementation

```go
// shared/infrastructure/eventbus/kafka_eventbus.go
package eventbus

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/confluentinc/confluent-kafka-go/v2/kafka"
    "github.com/vdntruong/dddcqrs/shared/domain/events"
)

type KafkaEventBus struct {
    producer *kafka.Producer
    consumer *kafka.Consumer
}

func NewKafkaEventBus() *KafkaEventBus {
    // Producer configuration
    producer, err := kafka.NewProducer(&kafka.ConfigMap{
        "bootstrap.servers": "localhost:9092",
        "client.id":        "order-management-service",
        "acks":            "all",
    })
    if err != nil {
        panic(err)
    }
    
    // Consumer configuration
    consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
        "bootstrap.servers": "localhost:9092",
        "group.id":         "order-reporting-service",
        "auto.offset.reset": "earliest",
    })
    if err != nil {
        panic(err)
    }
    
    return &KafkaEventBus{
        producer: producer,
        consumer: consumer,
    }
}

func (k *KafkaEventBus) Publish(ctx context.Context, event events.DomainEvent) error {
    eventData, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    topic := "orders"
    message := &kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Value: eventData,
        Headers: []kafka.Header{
            {Key: "event-type", Value: []byte(event.Type())},
            {Key: "aggregate-id", Value: []byte(event.AggregateID())},
        },
    }
    
    return k.producer.Produce(message, nil)
}

func (k *KafkaEventBus) Subscribe(ctx context.Context, topic string, handler func(events.DomainEvent) error) error {
    err := k.consumer.Subscribe(topic, nil)
    if err != nil {
        return err
    }
    
    go func() {
        for {
            msg, err := k.consumer.ReadMessage(-1)
            if err != nil {
                continue
            }
            
            var event events.DomainEvent
            if err := json.Unmarshal(msg.Value, &event); err != nil {
                continue
            }
            
            if err := handler(event); err != nil {
                // Handle error, possibly retry or dead letter queue
                continue
            }
            
            k.consumer.CommitMessage(msg)
        }
    }()
    
    return nil
}
```

### Redis Caching Implementation

```go
// order-reporting-service/internal/readmodels/order_read_model.go
package readmodels

import (
    "context"
    "encoding/json"
    "time"
    
    "github.com/redis/go-redis/v9"
    "github.com/vdntruong/dddcqrs/shared/domain/events"
)

type OrderReadModel struct {
    db    *sql.DB
    redis *redis.Client
}

func NewOrderReadModel(db *sql.DB, redis *redis.Client) *OrderReadModel {
    return &OrderReadModel{
        db:    db,
        redis: redis,
    }
}

func (rm *OrderReadModel) GetOrder(ctx context.Context, orderID string) (*OrderDTO, error) {
    // Try cache first
    cached, err := rm.redis.Get(ctx, "order:"+orderID).Result()
    if err == nil {
        var order OrderDTO
        if err := json.Unmarshal([]byte(cached), &order); err == nil {
            return &order, nil
        }
    }
    
    // Fallback to database
    query := `SELECT id, customer_id, status, total_amount, created_at FROM orders WHERE id = $1`
    row := rm.db.QueryRowContext(ctx, query, orderID)
    
    var order OrderDTO
    err = row.Scan(&order.ID, &order.CustomerID, &order.Status, &order.TotalAmount, &order.CreatedAt)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    orderData, _ := json.Marshal(order)
    rm.redis.Set(ctx, "order:"+orderID, orderData, 1*time.Hour)
    
    return &order, nil
}

func (rm *OrderReadModel) CreateOrder(ctx context.Context, order *OrderDTO) error {
    // Save to database
    query := `INSERT INTO orders (id, customer_id, status, total_amount, created_at) VALUES ($1, $2, $3, $4, $5)`
    _, err := rm.db.ExecContext(ctx, query, order.ID, order.CustomerID, order.Status, order.TotalAmount, order.CreatedAt)
    if err != nil {
        return err
    }
    
    // Cache the result
    orderData, _ := json.Marshal(order)
    rm.redis.Set(ctx, "order:"+order.ID, orderData, 1*time.Hour)
    
    return nil
}
```

This architecture provides:
- **Independent scaling** of read and write operations
- **Eventual consistency** between services
- **Reliable event delivery** via outbox pattern
- **Technology flexibility** (different databases for read/write)
- **Fault tolerance** (services can fail independently)

## Best Practices

### 1. Start Simple
Begin with a simple CQRS implementation before adding complexity. Don't over-engineer from the start.

### 2. Domain Focus
Keep domain logic in the domain layer, not in infrastructure. The domain should be independent of external concerns.

### 3. Event-Driven Architecture
Use domain events to maintain consistency across aggregates and enable loose coupling.

### 4. Validation
Validate commands before processing. Implement proper input validation at the application boundary.

### 5. Error Handling
Implement proper error handling and rollback mechanisms. Use Go's error handling patterns effectively.

### 6. Testing
Write comprehensive tests for domain logic and handlers. Use interfaces to enable easy mocking.

### 7. Bounded Contexts
Define clear boundaries between different parts of your system. Each bounded context should have its own domain model.

### 8. Ubiquitous Language
Use the same terminology in code as used by domain experts. This improves communication and reduces misunderstandings.

## Benefits

### CQRS Benefits

1. **Independent Services**: Command and Query services can be deployed, scaled, and maintained independently
2. **Technology Flexibility**: Different databases and technologies for read vs write operations
3. **Performance Optimization**: Read models optimized for queries, write models optimized for business logic
4. **Scalability**: Scale read and write workloads independently based on demand
5. **Fault Isolation**: Failure in one service doesn't affect the other
6. **Team Autonomy**: Different teams can work on command and query services independently
7. **Event-Driven Architecture**: Loose coupling through domain events
8. **Eventual Consistency**: Acceptable for most business scenarios, enables better performance

### DDD Benefits

1. **Business Alignment**: Code reflects business logic and rules
2. **Maintainability**: Clear domain boundaries and business logic encapsulation
3. **Testability**: Easy to unit test individual components
4. **Communication**: Ubiquitous language improves team communication
5. **Evolution**: System can evolve with changing business requirements

### Combined Benefits

1. **Clear Architecture**: Well-defined layers and responsibilities
2. **Independent Scaling**: Different parts of the system can scale independently
3. **Technology Flexibility**: Can use different technologies for different concerns
4. **Team Productivity**: Clear boundaries enable parallel development
5. **System Resilience**: Failures in one part don't affect others

### Go Workspace Benefits for CQRS

1. **Module Isolation**: Each service is a separate Go module with its own dependencies
2. **Shared Code Management**: Domain models and infrastructure shared via workspace
3. **Independent Deployment**: Each service can be built and deployed independently
4. **Version Management**: Different services can use different versions of shared dependencies
5. **Development Efficiency**: Single workspace for development, separate modules for deployment
6. **CI/CD Optimization**: Build only changed modules, cache dependencies per module
7. **Team Collaboration**: Different teams can work on different modules without conflicts
8. **Dependency Management**: Clear separation of service-specific vs shared dependencies

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Resources

- [Domain-Driven Design Reference](https://www.domainlanguage.com/ddd/reference/)
- [CQRS Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/cqrs)
- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [Event Sourcing Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/event-sourcing)
- [Real-Time database change tracking in Go (CDC)](https://packagemain.tech/p/real-time-database-change-tracking)

## Acknowledgments

- Thanks to Eric Evans for Domain-Driven Design
- Thanks to Greg Young for CQRS
- Thanks to the Go community for excellent tooling and practices
