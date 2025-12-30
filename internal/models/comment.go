package models

import "time"

type PullRequestComment struct {
	ID        int       `json:"id"`
	PRID      int       `json:"pr_id"`
	GitHubID  int       `json:"github_id"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type IssueComment struct {
	ID        int       `json:"id"`
	IssueID   int       `json:"issue_id"`
	GitHubID  int       `json:"github_id"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
