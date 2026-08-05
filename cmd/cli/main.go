package main

import (
	"fmt"
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

const interval = 10 * time.Minute
const exportInterval = 5 * time.Minute

func main() {
	log.Println("Starting devstats...")
	baseDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := filepath.Join(baseDir, "devstats.db")
	anonDBPath := filepath.Join(baseDir, "devstats_anon.db")
	log.Printf("Using database at: %s", dbPath)

	keypressStore, err := storage.NewSQLiteStore[domain.KeypressData](dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer keypressStore.Close()

	keypressCollector := collector.NewKeypressCollector(keypressStore)
	if err := keypressCollector.Start(); err != nil {
		log.Fatalf("Failed to start keypress collector: %v", err)
	}
	defer keypressCollector.Stop()
	log.Println("Keypress collector started. Press Ctrl+C to stop.")

	keypressAnonStore, err := storage.NewSQLiteStore[domain.KeypressAnonymousStats](anonDBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer keypressAnonStore.Close()

	keypressAnonymizer, err := anon.NewService[domain.KeypressData, domain.KeypressAnonymousStats](
		keypressStore, keypressAnonStore, anon.Config{IntervalSize: interval},
	)
	if err != nil {
		log.Fatal(err)
	}

	process := func(end time.Time) {
		start := end.Add(-interval)
		if err := keypressAnonymizer.ProcessInterval(start, end); err != nil {
			log.Printf("Error processing keypress interval: %v", err)
		}
	}
	process(time.Now())

	// JSONL export: one line per 5-min window (count only, never keys), one file per day.
	exportDir := filepath.Join(homeDir, "notes", "blog", "keystrokes")
	export := func(end time.Time) {
		records, err := keypressStore.FindBetween(end.Add(-exportInterval), end)
		if err != nil {
			log.Printf("Error reading keypresses for export: %v", err)
			return
		}
		if err := os.MkdirAll(exportDir, 0o755); err != nil {
			log.Printf("Error creating export dir: %v", err)
			return
		}
		path := filepath.Join(exportDir, end.UTC().Format("2006-01-02")+".jsonl")
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			log.Printf("Error opening export file: %v", err)
			return
		}
		defer f.Close()
		if _, err := fmt.Fprintf(f, `{"t":%q,"n":%d}`+"\n", end.UTC().Format(time.RFC3339), len(records)); err != nil {
			log.Printf("Error writing export: %v", err)
		}
	}

	ticker := time.NewTicker(interval)
	exportTicker := time.NewTicker(exportInterval)
	defer ticker.Stop()
	defer exportTicker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-sigChan:
			log.Println("Shutting down gracefully...")
			return
		case t := <-ticker.C:
			process(t)
		case t := <-exportTicker.C:
			export(t)
		}
	}
}
