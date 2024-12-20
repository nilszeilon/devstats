package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sync"
	"time"
)

// Store defines the interface for data storage
type Store[T any] interface {
	Save(data T) error
	Get() ([]T, error)
	FindBetween(start, end interface{}) ([]any, error)
}

// FileStore implements Store interface using file storage
type FileStore[T any] struct {
	filepath string
	mu       sync.RWMutex
	data     []T
}

func NewFileStore[T any](filepath string) (*FileStore[T], error) {
	fs := &FileStore[T]{
		filepath: filepath,
		data:     make([]T, 0),
	}

	// Load existing data if file exists
	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		data, err := os.ReadFile(filepath)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, &fs.data); err != nil {
			return nil, err
		}
	}

	return fs, nil
}

func (fs *FileStore[T]) Save(data T) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data = append(fs.data, data)
	return fs.persist()
}

func (fs *FileStore[T]) Get() ([]T, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	return fs.data, nil
}

// FindBetween returns records between start and end timestamps
func (fs *FileStore[T]) FindBetween(start, end interface{}) ([]any, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Convert start and end to time.Time
	startTime, ok := start.(time.Time)
	if !ok {
		return nil, fmt.Errorf("start time must be time.Time, got %T", start)
	}

	endTime, ok := end.(time.Time)
	if !ok {
		return nil, fmt.Errorf("end time must be time.Time, got %T", end)
	}

	var results []any

	for _, item := range fs.data {
		// Use reflection to get the Timestamp field
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		timestampField := v.FieldByName("Timestamp")
		if !timestampField.IsValid() {
			return nil, fmt.Errorf("struct must have Timestamp field")
		}

		timestamp, ok := timestampField.Interface().(time.Time)
		if !ok {
			return nil, fmt.Errorf("Timestamp field must be time.Time")
		}

		// Check if timestamp is within range
		if (timestamp.Equal(startTime) || timestamp.After(startTime)) &&
			(timestamp.Equal(endTime) || timestamp.Before(endTime)) {
			results = append(results, item)
		}
	}

	return results, nil
}

func (fs *FileStore[T]) persist() error {
	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filepath, data, 0644)
}
