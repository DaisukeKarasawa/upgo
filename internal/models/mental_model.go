package models

import "time"

type MentalModelAnalysis struct {
	ID              int       `json:"id"`
	RepositoryID    int       `json:"repository_id"`
	AnalysisType    string    `json:"analysis_type"`
	AnalysisContent string    `json:"analysis_content"`
	AnalyzedPRIDs   string    `json:"analyzed_pr_ids"`   // JSON配列
	AnalyzedIssueIDs string  `json:"analyzed_issue_ids"` // JSON配列
	CreatedAt       time.Time `json:"created_at"`
}
