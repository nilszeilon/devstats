package domain

import "time"

type KeypressData struct {
	Key       string    `json:"key" sql:"TEXT NOT NULL"`
	Timestamp time.Time `json:"timestamp" sql:"DATETIME NOT NULL"`
}

// KeypressAnonymousStats represents anonymized statistics for keypresses
type KeypressAnonymousStats struct {
	Timestamp       time.Time `json:"timestamp" sql:"DATETIME NOT NULL"`
	KeypressesCount int64     `json:"keypresses_count" sql:"INTEGER NOT NULL"`
}

// TableName returns the custom table name for SQLite storage
func (KeypressData) TableName() string {
	return "keypresses"
}

// TableName returns the custom table name for anonymous storage
func (KeypressAnonymousStats) TableName() string {
	return "keypresses_anonymous"
}

// GetTimestamp implements the Anonymizable interface
func (k KeypressData) GetTimestamp() time.Time {
	return k.Timestamp
}

// Anonymize implements the Anonymizable interface
func (k KeypressData) Anonymize(records []any, intervalStart time.Time) ([]KeypressAnonymousStats, error) {
	// Create a map to count keypresses per key
	var keyCount int64

	// Count occurrences of each key
	for _, record := range records {
		if _, ok := record.(KeypressData); ok {
			keyCount++
		}
	}

	stats := make([]KeypressAnonymousStats, 0, 1)
	stats = append(stats, KeypressAnonymousStats{
		Timestamp:       intervalStart,
		KeypressesCount: keyCount,
	})

	return stats, nil
}
