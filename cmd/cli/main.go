package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/nilszeilon/devstats/internal/collector"
	"github.com/nilszeilon/devstats/internal/domain"
	"github.com/nilszeilon/devstats/internal/storage"
)

func main() {
	log.Println("Starting devstats...")
	// Get the current working directory (where the program was started from)
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Create absolute paths for all files
	dbPath := filepath.Join(baseDir, "devstats.db")
	log.Printf("Using database at: %s", dbPath)

	// init sqlite storage
	store, err := storage.NewSQLiteStore[domain.KeypressData](dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	// Create keypress collector
	keypressCollector := collector.NewKeypressCollector(store)

	// Start collecting
	if err := keypressCollector.Start(); err != nil {
		log.Fatalf("Failed to start keypress collector: %v", err)
	}

	// init sqlite storage
	fileChangeStore, err := storage.NewSQLiteStore[domain.FileChangeData](dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileChangeStore.Close()

	// Create the collector with paths to watch
	paths := []string{
		"/Users/nilszeilon",
		// Add more paths as needed
	}

	fileCollector, err := collector.NewFileChangeCollector(fileChangeStore, paths)
	if err != nil {
		log.Fatal(err)
	}

	// Start collecting
	err = fileCollector.Start()
	if err != nil {
		log.Fatal(err)
	}

	// Don't forget to stop it when done
	defer fileCollector.Stop()

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
