package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nilszeilon/devstats/internal/anon"
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

	// Get user's home directory from environment variable
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get home directory:", err)
	}

	// Create the collector with paths to watch
	paths := []string{
		homeDir,
		// Add more paths as needed
	}

	// Create absolute paths for all files
	dbPath := filepath.Join(baseDir, "devstats.db")
	log.Printf("Using database at: %s", dbPath)

	// Setup anonymizer stores
	anonDBPath := filepath.Join(baseDir, "devstats_anon.db")

	// init sqlite storage
	keypressStore, err := storage.NewSQLiteStore[domain.KeypressData](dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer keypressStore.Close()

	// Create keypress collector
	keypressCollector := collector.NewKeypressCollector(keypressStore)

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

	// Create stores for anonymous data
	keypressAnonStore, err := storage.NewSQLiteStore[domain.KeypressAnonymousStats](anonDBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer keypressAnonStore.Close()

	fileChangeAnonStore, err := storage.NewSQLiteStore[domain.FileChangeAnonymousStats](anonDBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileChangeAnonStore.Close()

	// Create anonymizer services
	keypressAnonymizer, err := anon.NewService[domain.KeypressData, domain.KeypressAnonymousStats](
		keypressStore,
		keypressAnonStore,
		anon.Config{
			IntervalSize: 10 * time.Minute,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fileChangeAnonymizer, err := anon.NewService[domain.FileChangeData, domain.FileChangeAnonymousStats](
		fileChangeStore,
		fileChangeAnonStore,
		anon.Config{
			IntervalSize: 10 * time.Minute,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Start anonymization ticker
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	// Run first anonymization immediately
	now := time.Now()
	start := now.Add(-10 * time.Minute)
	if err := keypressAnonymizer.ProcessInterval(start, now); err != nil {
		log.Printf("Error processing keypress interval: %v", err)
	}
	if err := fileChangeAnonymizer.ProcessInterval(start, now); err != nil {
		log.Printf("Error processing file change interval: %v", err)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either interrupt signal or ticker
	for {
		select {
		case <-sigChan:
			log.Println("Shutting down gracefully...")
			keypressCollector.Stop()
			fileCollector.Stop()
			log.Println("Shutdown complete")
			return
		case t := <-ticker.C:
			start := t.Add(-10 * time.Minute)
			if err := keypressAnonymizer.ProcessInterval(start, t); err != nil {
				log.Printf("Error processing keypress interval: %v", err)
			}
			if err := fileChangeAnonymizer.ProcessInterval(start, t); err != nil {
				log.Printf("Error processing file change interval: %v", err)
			}
		}
	}

	log.Println("Shutting down gracefully...")
	keypressCollector.Stop()
	log.Println("Shutdown complete")
}
