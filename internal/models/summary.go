package models

import "time"

type PullRequestSummary struct {
	ID               int       `json:"id"`
	PRID             int       `json:"pr_id"`
	DescriptionSummary string   `json:"description_summary"`
	DiffSummary       string   `json:"diff_summary"`
	DiffExplanation   string   `json:"diff_explanation"`
	CommentsSummary   string   `json:"comments_summary"`
	DiscussionSummary string   `json:"discussion_summary"`
	MergeReason       string   `json:"merge_reason"`
	CloseReason       string   `json:"close_reason"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type IssueSummary struct {
	ID               int       `json:"id"`
	IssueID          int       `json:"issue_id"`
	DescriptionSummary string   `json:"description_summary"`
	CommentsSummary   string   `json:"comments_summary"`
	DiscussionSummary string   `json:"discussion_summary"`
	UpdatedAt         time.Time `json:"updated_at"`
}
