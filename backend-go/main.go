package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/handler"
	"backend/internal/poller"
	"backend/internal/storage"
)

func main() {
	// TODO: move to slog to have more control (e.g. log levels)
	logger := log.New(os.Stdout, "[POLLER] ", log.LstdFlags)

	logger.Println("Starting App Store Review Poller...")

	// Initialize storage
	storageFilePath := "data/reviews.json"
	store, err := storage.NewFileStorage(storageFilePath)
	if err != nil {
		logger.Fatalf("Failed to create storage: %v", err)
	}

	// Load existing state from disk
	if err := store.LoadState(); err != nil {
		logger.Printf("Warning: Failed to load existing state: %v", err)
		logger.Println("Starting with empty state...")
	} else {
		logger.Println("Successfully loaded existing review data")
	}

	// Load config
	config, err := loadConfig("config/apps.json")
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Start reviewPoller
	pollInterval := 5 * time.Minute
	reviewPoller := poller.NewPoller(store, logger, config.Apps, pollInterval)
	reviewPoller.Start()

	// Setup HTTP handlers
	h := handler.NewHandler(store)
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/reviews", h.GetRecentReviews)
	mux.HandleFunc("/api/health", h.HealthCheck)
	mux.HandleFunc("/api/average-rating", h.GetAverageRating)

	// Wrap with CORS middleware
	corsHandler := enableCORS(mux)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	go func() {
		log.Println("HTTP server listening on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan
	logger.Println("\nShutdown signal received, cleaning up...")

	logger.Println("Stopping poller...")
	reviewPoller.Stop()

	log.Println("Saving final state...")
	if err := store.SaveState(); err != nil {
		log.Printf("Error saving state: %v", err)
	}

	// Shutdown server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	logger.Println("Shutdown complete")
}

type Config struct {
	Apps []string `json:"apps"`
}

func loadConfig(filepath string) (*Config, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// enableCORS adds CORS headers to allow frontend access
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from frontend
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
