package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/thoraf20/payment-processor/api"
	"github.com/thoraf20/payment-processor/config"
	"github.com/thoraf20/payment-processor/engine"
	"github.com/thoraf20/payment-processor/logger"
	"github.com/thoraf20/payment-processor/processors"
	"github.com/thoraf20/payment-processor/repository"
	"go.uber.org/zap"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		zap.L().Info("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		zap.L().Fatal("Failed to load config", zap.Error(err))
	}
	
	// Initialize logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		zap.L().Fatal("Failed to initialize logger", zap.Error(err))
	}
	defer log.Sync()
	zap.ReplaceGlobals(log) // Set as global logger
	
	// Set up database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Verify database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database", zap.Error(err))
	}

	// Initialize repositories
	paymentRepo := repository.NewPostgresPaymentRepository(db, log)

	// Verify the repository implements all methods
var _ repository.PaymentRepository = (*repository.PostgresPaymentRepository)(nil)

	// Initialize payment processors
	stripeProcessor := processors.NewStripeProcessor(cfg.StripeAPIKey)
	// flutterwaveProcessor := processors.NewFlutterwaveProcessor(cfg.FlutterWaveAPIKey, paymentRepo)
	// paystackProcessor := processors.NewPaystackProcessor(cfg.PayStackAPIKey, paymentRepo)

	// Create processor router (if you have multiple processors)
	// processorRouter := engine.NewPaymentEngine(
	// 	stripeProcessor,
	// 	// flutterwaveProcessor,
	// 	// paystackProcessor,
	// )

	// Initialize payment engine
	paymentEngine := engine.NewPaymentEngine(stripeProcessor, paymentRepo)

	// Initialize HTTP server with all dependencies
	server := api.NewServer(log, paymentEngine)

	// Create HTTP server with timeouts
	httpServer := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      server,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Info("Starting server", zap.String("port", cfg.HTTPPort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", zap.Error(err))
		}
	}()
	
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Info("Shutting down server...")
	
	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("Server shutdown failed", zap.Error(err))
	}

	// Close database connection
	if err := db.Close(); err != nil {
		log.Error("Database connection close failed", zap.Error(err))
	}

	log.Info("Server exited properly")
}