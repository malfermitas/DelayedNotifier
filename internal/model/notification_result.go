package model

import "time"

type NotificationResult struct {
	ID     string             `json:"id"`
	Status NotificationStatus `json:"status"` // sent, failed, cancelled
	SentAt time.Time          `json:"sent_at"`
	Error  string             `json:"error,omitempty"`
}
