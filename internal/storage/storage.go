package storage

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// DataPoint represents a single piece of collected data
type DataPoint struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// Store defines the interface for data storage
type Store interface {
	Save(dataType string, data interface{}) error
	Get(dataType string) ([]DataPoint, error)
}

// FileStore implements Store interface using file storage
type FileStore struct {
	filepath string
	mu       sync.RWMutex
	data     []DataPoint
}

func NewFileStore(filepath string) (*FileStore, error) {
	fs := &FileStore{
		filepath: filepath,
		data:     make([]DataPoint, 0),
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

func (fs *FileStore) Save(dataType string, data interface{}) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	dp := DataPoint{
		ID:        generateID(),
		Type:      dataType,
		Timestamp: time.Now(),
		Data:      data,
	}

	fs.data = append(fs.data, dp)
	return fs.persist()
}

func (fs *FileStore) Get(dataType string) ([]DataPoint, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var result []DataPoint
	for _, dp := range fs.data {
		if dp.Type == dataType {
			result = append(result, dp)
		}
	}
	return result, nil
}

func (fs *FileStore) persist() error {
	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.filepath, data, 0644)
}

func generateID() string {
	return time.Now().Format("20060102150405.000")
}
