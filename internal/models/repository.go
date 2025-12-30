package models

import "time"

type Repository struct {
	ID           int       `json:"id"`
	Owner        string    `json:"owner"`
	Name         string    `json:"name"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
}
