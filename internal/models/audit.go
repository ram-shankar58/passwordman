package models

import "time"

// AuditLog tracks vault entry access
type AuditLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"userId"`
	EntryID   int64     `json:"entryId"`
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}
