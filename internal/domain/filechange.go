package domain

import "time"

type FileChangeData struct {
	Language  string    `json:"language" sql:"TEXT NOT NULL"`
	Timestamp time.Time `json:"timestamp" sql:"DATETIME NOT NULL"`
}

// FileChangeAnonymousStats represents anonymized statistics for file changes per language
type FileChangeAnonymousStats struct {
	Timestamp     time.Time `json:"timestamp" sql:"DATETIME NOT NULL"`
	Language      string    `json:"language" sql:"TEXT NOT NULL"`
	ChangesInSpan int64     `json:"changes_in_span" sql:"INTEGER NOT NULL"`
}

// TableName returns the custom table name for SQLite storage
func (FileChangeData) TableName() string {
	return "file_changes"
}

// TableName returns the custom table name for anonymous storage
func (FileChangeAnonymousStats) TableName() string {
	return "file_changes_anonymous"
}

// GetTimestamp implements the Anonymizable interface
func (f FileChangeData) GetTimestamp() time.Time {
	return f.Timestamp
}

// Anonymize implements the Anonymizable interface
func (f FileChangeData) Anonymize(records []any, intervalStart time.Time) ([]FileChangeAnonymousStats, error) {
	// Map to count changes per language
	languageCounts := make(map[string]int64)

	// Count changes for each language
	for _, r := range records {
		if change, ok := r.(FileChangeData); ok {
			languageCounts[change.Language]++
		}
	}

	// Convert to slice of anonymous stats
	var stats []FileChangeAnonymousStats
	for lang, count := range languageCounts {
		stats = append(stats, FileChangeAnonymousStats{
			Timestamp:     intervalStart,
			Language:      lang,
			ChangesInSpan: count,
		})
	}

	return stats, nil
}
