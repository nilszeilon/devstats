package domain

import "time"

type KeypressData struct {
	Key       string    `json:"key" sql:"TEXT NOT NULL"`
	Timestamp time.Time `json:"timestamp" sql:"DATETIME NOT NULL"`
}

// TableName returns the custom table name for SQLite storage
func (KeypressData) TableName() string {
	return "keypresses"
}
