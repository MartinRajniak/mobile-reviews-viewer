package main

import (
    "backend/internal/poller"
    "log"
	"os"
	"os/signal"
	"syscall"
    "time"
)

func main() {
	logger := log.New(os.Stdout, "[POLLER] ", log.LstdFlags)

    logger.Println("Starting App Store Review Poller...")

    // Start reviewPoller
    // TODO: change to minutes for production
    pollInterval := 5 * time.Second
    reviewPoller := poller.NewPoller(logger, pollInterval)
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
