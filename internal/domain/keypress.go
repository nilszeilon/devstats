package domain

import "time"

type KeypressData struct {
	Key       string    `json:"key"`
	Timestamp time.Time `json:"timestamp"`
}
