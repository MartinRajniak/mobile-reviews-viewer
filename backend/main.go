package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	
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
	// TODO: change to minutes for production
	pollInterval := 5 * time.Second
	reviewPoller := poller.NewPoller(store, logger, config.Apps, pollInterval)
	reviewPoller.Start()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan
	logger.Println("\nShutdown signal received, cleaning up...")

	logger.Println("Stopping poller...")
	reviewPoller.Stop()
    store.SaveState()

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
