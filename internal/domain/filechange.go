package domain

import "time"

type FileChangeData struct {
	Language  string    `json:"language"`
	Timestamp time.Time `json:"timestamp"`
}
