package domain

import "time"

type FileChangeData struct {
	Language  string    `json:"language" sql:"TEXT NOT NULL"`
	Timestamp time.Time `json:"timestamp" sql:"DATETIME NOT NULL"`
}

// TableName returns the custom table name for SQLite storage
func (FileChangeData) TableName() string {
	return "file_changes"
}
