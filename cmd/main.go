package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paudarco/orders-db-task/internal/cache"
	"github.com/paudarco/orders-db-task/internal/config"
	"github.com/paudarco/orders-db-task/internal/database"
	"github.com/paudarco/orders-db-task/internal/handlers"
	"github.com/paudarco/orders-db-task/internal/kafka"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize cache
	cache := cache.NewCache()

	// Load initial data from database to cache
	orders, err := db.GetAllOrders()
	if err != nil {
		log.Fatalf("Failed to load orders from database: %v", err)
	}
	cache.LoadFromDB(orders)

	// Initialize Kafka consumer
	consumer := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, db, cache)

	// Start Kafka consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go consumer.Start(ctx)

	// Initialize HTTP handler
	handler := handlers.NewHandler(cache)

	// Set up HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: http.HandlerFunc(handler.GetOrder),
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting server on :%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Cancel Kafka consumer
	cancel()

	// Shutdown HTTP server
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close Kafka consumer
	if err := consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	log.Println("Server exiting")
}
