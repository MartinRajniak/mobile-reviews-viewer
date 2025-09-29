package main

import (
	"backend/internal/poller"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := log.New(os.Stdout, "[POLLER] ", log.LstdFlags)

	logger.Println("Starting App Store Review Poller...")

    // Load config
    config, err := loadConfig("config/apps.json")
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

	// Start reviewPoller
	// TODO: change to minutes for production
	pollInterval := 5 * time.Second
	reviewPoller := poller.NewPoller(logger, config.Apps, pollInterval)
	reviewPoller.Start()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan
	log.Println("\nShutdown signal received, cleaning up...")

	log.Println("Stopping poller...")
	reviewPoller.Stop()

	log.Println("Shutdown complete")
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
