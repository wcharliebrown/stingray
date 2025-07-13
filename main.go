package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"stingray/config"
	"stingray/database"
	"stingray/logging"
)

func main() {
	// Load config
	cfg := config.LoadConfig()

	// Initialize logger
	logger := logging.NewLogger(cfg.LoggingLevel)

	db, err := database.NewDatabase(cfg.GetDSN(), cfg.DebuggingMode)
	if err != nil {
		logger.LogError("Failed to connect to database: %v", err)
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	server := NewServer(db, cfg)

	// Start cleanup goroutines
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Clean up every hour
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := db.CleanupExpiredSessions(); err != nil {
					logger.LogError("Failed to cleanup expired sessions: %v", err)
					log.Printf("Failed to cleanup expired sessions: %v", err)
				}
				if err := db.CleanupExpiredPasswordResetTokens(); err != nil {
					logger.LogError("Failed to cleanup expired password reset tokens: %v", err)
					log.Printf("Failed to cleanup expired password reset tokens: %v", err)
				}
			}
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
			logger.LogError("Server error: %v", err)
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	logger.LogVerbose("Received shutdown signal")
	log.Println("Received shutdown signal")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.LogError("Server shutdown failed: %v", err)
		log.Fatalf("Server shutdown failed: %v", err)
	}
	logger.LogVerbose("Server exited gracefully")
	log.Println("Server exited gracefully")
} 