package storage

import (
	"encoding/json"
	"os"
	"sync"
)

// Store defines the interface for data storage
type Store[T any] interface {
	Save(data T) error
	Get() ([]T, error)
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

func (fs *FileStore[T]) persist() error {
	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filepath, data, 0644)
}
