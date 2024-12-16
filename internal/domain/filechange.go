package domain

import "time"

type FileChangeData struct {
	Filepath  string    `json:"filepath"`
	Action    string    `json:"action"` // "opened", "modified", "closed"
	Timestamp time.Time `json:"timestamp"`
}
