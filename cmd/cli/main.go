package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nilszeilon/devstats/internal/collector"
	"github.com/nilszeilon/devstats/internal/storage"
)

func main() {
	log.Println("Starting devstats...")

	// Initialize file storage
	store, err := storage.NewFileStore("devstats.json")
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create keypress collector
	keypressCollector := collector.NewKeypressCollector(store)

	// Start collecting
	if err := keypressCollector.Start(); err != nil {
		log.Fatalf("Failed to start keypress collector: %v", err)
	}

	log.Println("Keypress collector started. Press Ctrl+C to stop.")

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for interrupt signal
	<-sigChan

	log.Println("Shutting down gracefully...")
	keypressCollector.Stop()
	log.Println("Shutdown complete")
}
