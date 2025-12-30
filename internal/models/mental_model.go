package models

import "time"

// MentalModelAnalysis represents the analysis result of committer/reviewer mental models.
// It stores insights extracted from PR/Issue comments and discussions, including
// coding style preferences, review priorities, decision-making patterns, and technical philosophy.
type MentalModelAnalysis struct {
	// ID is the unique identifier for this analysis record.
	ID int `json:"id"`
	// RepositoryID is the ID of the repository this analysis belongs to.
	RepositoryID int `json:"repository_id"`
	// AnalysisType indicates the type of analysis performed (e.g., "committer", "reviewer").
	AnalysisType string `json:"analysis_type"`
	// AnalysisContent contains the detailed analysis result in text format.
	AnalysisContent string `json:"analysis_content"`
	// AnalyzedPRIDs is a JSON array string containing the IDs of PRs used in this analysis.
	AnalyzedPRIDs string `json:"analyzed_pr_ids"` // JSON array
	// AnalyzedIssueIDs is a JSON array string containing the IDs of issues used in this analysis.
	AnalyzedIssueIDs string `json:"analyzed_issue_ids"` // JSON array
	// CreatedAt is the timestamp when this analysis was created.
	CreatedAt time.Time `json:"created_at"`
}
