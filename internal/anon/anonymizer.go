package anon

import (
	"fmt"
	"time"

	"github.com/nilszeilon/devstats/internal/storage"
)

// Anonymizable defines the interface that source types must implement
type Anonymizable[T any] interface {
	GetTimestamp() time.Time
	Anonymize([]any, time.Time) ([]T, error)
}

// Config holds the configuration for the anonymizer service
type Config struct {
	IntervalSize time.Duration
}

// Service handles the anonymization process
type Service[S Anonymizable[T], T any] struct {
	sourceStore storage.Store[S]
	targetStore storage.Store[T]
	config      Config
}

// NewService creates a new anonymizer service
func NewService[S Anonymizable[T], T any](
	sourceStore storage.Store[S],
	targetStore storage.Store[T],
	config Config,
) (*Service[S, T], error) {
	if config.IntervalSize == 0 {
		return nil, fmt.Errorf("interval size must be greater than 0")
	}

	return &Service[S, T]{
		sourceStore: sourceStore,
		targetStore: targetStore,
		config:      config,
	}, nil
}

// ProcessInterval processes and anonymizes data for a specific time interval
func (s *Service[S, T]) ProcessInterval(start, end time.Time) error {
	// Fetch records from source store
	records, err := s.sourceStore.FindBetween(start, end)
	if err != nil {
		return fmt.Errorf("failed to fetch records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	// Get a sample record to use for anonymization
	sample, ok := records[0].(S)
	if !ok {
		return fmt.Errorf("failed to cast record to source type")
	}

	// Anonymize the records
	anonymizedRecords, err := sample.Anonymize(records, start)
	if err != nil {
		return fmt.Errorf("failed to anonymize records: %w", err)
	}

	// Save each anonymized record
	for _, record := range anonymizedRecords {
		if err := s.targetStore.Save(record); err != nil {
			return fmt.Errorf("failed to save anonymized data: %w", err)
		}
	}

	return nil
}
