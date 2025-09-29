package main

import (
    "log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
    log.Println("Starting App Store Review Poller...")

    // Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Wait for interrupt signal
	<-sigChan
    log.Println("\nShutdown signal received, cleaning up...")

    log.Println("Shutdown complete")
}
