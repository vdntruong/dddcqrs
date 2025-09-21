package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"github.com/vdntruong/dddcqrs/order-reporting-service/internal/handlers"
	"github.com/vdntruong/dddcqrs/order-reporting-service/internal/readmodels"
	"github.com/vdntruong/dddcqrs/shared/infrastructure/eventbus"
)

func main() {
    // Initialize database
    db := initDatabase()
    defer db.Close()
    
    // Initialize Redis
    redisClient := initRedis()
    defer redisClient.Close()
    
    // Initialize Kafka event bus
    kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
    eventBus := eventbus.NewKafkaEventBus(kafkaBrokers)
    defer eventBus.Close()
    
    // Initialize read models
    orderReadModel := readmodels.NewOrderReadModel(db, redisClient)
    customerReadModel := readmodels.NewCustomerReadModel(db, redisClient)
    
    // Initialize projection handlers
    orderProjectionHandler := &handlers.OrderProjectionHandler{
        OrderReadModel:    orderReadModel,
        CustomerReadModel: customerReadModel,
    }
    
    // Initialize query handlers
    getOrderHandler := &handlers.GetOrderHandler{
        ReadModel: orderReadModel,
    }
    
    listOrdersHandler := &handlers.ListOrdersHandler{
        ReadModel: orderReadModel,
    }
    
    getOrderAnalyticsHandler := &handlers.GetOrderAnalyticsHandler{
        ReadModel: orderReadModel,
    }
    
    // Initialize HTTP router
    router := mux.NewRouter()
    
    // API routes
    api := router.PathPrefix("/api/v1").Subrouter()
    api.HandleFunc("/orders/{id}", getOrderHandler.HandleHTTP).Methods("GET")
    api.HandleFunc("/orders", listOrdersHandler.HandleHTTP).Methods("GET")
    api.HandleFunc("/analytics/orders", getOrderAnalyticsHandler.HandleHTTP).Methods("GET")
    
    // Health check
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    }).Methods("GET")
    
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
    port := getEnv("PORT", "8081")
    server := &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }
    
    // Graceful shutdown
    go func() {
        log.Printf("Order Reporting Service starting on port %s", port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()
    
    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    // Graceful shutdown with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := server.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }
    
    log.Println("Server exited")
}

func initDatabase() *sql.DB {
    databaseURL := getEnv("DATABASE_URL", "postgres://postgres:password@localhost:5432/ecommerce_orders?sslmode=disable")
    
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    
    if err := db.Ping(); err != nil {
        log.Fatalf("Failed to ping database: %v", err)
    }
    
    // Set connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    log.Println("Connected to database")
    return db
}

func initRedis() *redis.Client {
    redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
    
    opt, err := redis.ParseURL(redisURL)
    if err != nil {
        log.Fatalf("Failed to parse Redis URL: %v", err)
    }
    
    client := redis.NewClient(opt)
    
    if err := client.Ping(context.Background()).Err(); err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    
    log.Println("Connected to Redis")
    return client
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
