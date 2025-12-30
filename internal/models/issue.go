package models

import "time"

type Issue struct {
	ID            int        `json:"id"`
	RepositoryID  int        `json:"repository_id"`
	GitHubID      int        `json:"github_id"`
	Title         string     `json:"title"`
	Body          string     `json:"body"`
	State         string     `json:"state"`
	PreviousState string     `json:"previous_state"`
	Author        string     `json:"author"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ClosedAt      *time.Time `json:"closed_at"`
	URL           string     `json:"url"`
	LastSyncedAt  *time.Time `json:"last_synced_at"`
}
