package models

import (
	"time"
)

// HistoryEntry represents a point in an entity's history
type HistoryEntry struct {
	Timestamp time.Time    `json:"timestamp"`
	Entity    *Entity      `json:"entity"`
	Changes   []string     `json:"changes"` // Tags that changed at this timestamp
}