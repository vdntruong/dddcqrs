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
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/vdntruong/dddcqrs/order-management-service/internal/handlers"
	"github.com/vdntruong/dddcqrs/order-management-service/internal/repositories"
	svcSwagger "github.com/vdntruong/dddcqrs/order-management-service/internal/swagger"
	"github.com/vdntruong/dddcqrs/shared/infrastructure/eventbus"
)

func main() {
	// Initialize database
	db := initDatabase()
	defer db.Close()
	
	// Initialize Kafka event bus
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	eventBus := eventbus.NewKafkaEventBus(kafkaBrokers)
	defer eventBus.Close()
	
	// Initialize repositories
	orderRepo := repositories.NewOrderRepository(db)
	outboxRepo := repositories.NewOutboxRepository(db)
	eventStore := repositories.NewEventStore(db)
	
	// Initialize command service
	commandService := &handlers.CommandService{
		OrderRepo:  orderRepo,
		EventStore: eventStore,
		Outbox:     outboxRepo,
		EventBus:   eventBus,
	}
	
	// Initialize command handlers
	createOrderHandler := &handlers.CreateOrderHandler{
		Service: commandService,
	}
	
	updateOrderHandler := &handlers.UpdateOrderHandler{
		Service: commandService,
	}
	
	confirmOrderHandler := &handlers.ConfirmOrderHandler{
		Service: commandService,
	}
	
	cancelOrderHandler := &handlers.CancelOrderHandler{
		Service: commandService,
	}
	
	// Initialize HTTP router
	router := mux.NewRouter()
	
	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/orders", createOrderHandler.HandleHTTP).Methods("POST")
	api.HandleFunc("/orders/{id}", updateOrderHandler.HandleHTTP).Methods("PUT")
	api.HandleFunc("/orders/{id}/confirm", confirmOrderHandler.HandleHTTP).Methods("POST")
	api.HandleFunc("/orders/{id}/cancel", cancelOrderHandler.HandleHTTP).Methods("POST")
	
	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Swagger docs and UI
	router.HandleFunc("/swagger/doc.json", svcSwagger.ServeDoc).Methods("GET")
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	
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
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}
	
	// Graceful shutdown
	go func() {
		log.Printf("Order Management Service starting on port %s", port)
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

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
