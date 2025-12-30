package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewHandlers(db *sql.DB, logger *zap.Logger) *Handlers {
	return &Handlers{
		db:     db,
		logger: logger,
	}
}

func (h *Handlers) GetPRs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	state := c.Query("state")
	author := c.Query("author")
	search := c.Query("search")

	offset := (page - 1) * limit

	query := "SELECT id, repository_id, github_id, title, body, state, author, created_at, updated_at, merged_at, closed_at, url FROM pull_requests WHERE 1=1"
	args := []interface{}{}

	if state != "" {
		query += " AND state = ?"
		args = append(args, state)
	}
	if author != "" {
		query += " AND author = ?"
		args = append(args, author)
	}
	if search != "" {
		query += " AND (title LIKE ? OR body LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.Error("PR一覧の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PR一覧の取得に失敗しました"})
		return
	}
	defer rows.Close()

	var prs []map[string]interface{}
	for rows.Next() {
		var pr struct {
			ID         int
			RepoID     int
			GitHubID   int
			Title      string
			Body       string
			State      string
			Author     string
			CreatedAt  string
			UpdatedAt  string
			MergedAt   sql.NullString
			ClosedAt   sql.NullString
			URL        string
		}
		err := rows.Scan(&pr.ID, &pr.RepoID, &pr.GitHubID, &pr.Title, &pr.Body, &pr.State, &pr.Author, &pr.CreatedAt, &pr.UpdatedAt, &pr.MergedAt, &pr.ClosedAt, &pr.URL)
		if err != nil {
			continue
		}

		prMap := map[string]interface{}{
			"id":         pr.ID,
			"repository_id": pr.RepoID,
			"github_id":  pr.GitHubID,
			"title":      pr.Title,
			"body":       pr.Body,
			"state":      pr.State,
			"author":     pr.Author,
			"created_at": pr.CreatedAt,
			"updated_at": pr.UpdatedAt,
			"url":        pr.URL,
		}
		if pr.MergedAt.Valid {
			prMap["merged_at"] = pr.MergedAt.String
		}
		if pr.ClosedAt.Valid {
			prMap["closed_at"] = pr.ClosedAt.String
		}

		prs = append(prs, prMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": prs,
		"page": page,
		"limit": limit,
	})
}

func (h *Handlers) GetPR(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なIDです"})
		return
	}

	var pr struct {
		ID         int
		RepoID     int
		GitHubID   int
		Title      string
		Body       string
		State      string
		Author     string
		CreatedAt  string
		UpdatedAt  string
		MergedAt   sql.NullString
		ClosedAt   sql.NullString
		URL        string
	}

	err = h.db.QueryRow(
		"SELECT id, repository_id, github_id, title, body, state, author, created_at, updated_at, merged_at, closed_at, url FROM pull_requests WHERE id = ?",
		id,
	).Scan(&pr.ID, &pr.RepoID, &pr.GitHubID, &pr.Title, &pr.Body, &pr.State, &pr.Author, &pr.CreatedAt, &pr.UpdatedAt, &pr.MergedAt, &pr.ClosedAt, &pr.URL)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "PRが見つかりません"})
		return
	}
	if err != nil {
		h.logger.Error("PR詳細の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PR詳細の取得に失敗しました"})
		return
	}

	prMap := map[string]interface{}{
		"id":            pr.ID,
		"repository_id": pr.RepoID,
		"github_id":     pr.GitHubID,
		"title":         pr.Title,
		"body":          pr.Body,
		"state":         pr.State,
		"author":        pr.Author,
		"created_at":    pr.CreatedAt,
		"updated_at":    pr.UpdatedAt,
		"url":           pr.URL,
	}
	if pr.MergedAt.Valid {
		prMap["merged_at"] = pr.MergedAt.String
	}
	if pr.ClosedAt.Valid {
		prMap["closed_at"] = pr.ClosedAt.String
	}

	// 要約データの取得
	var summary struct {
		DescriptionSummary sql.NullString
		DiffSummary        sql.NullString
		DiffExplanation    sql.NullString
		CommentsSummary    sql.NullString
		DiscussionSummary  sql.NullString
		MergeReason        sql.NullString
		CloseReason        sql.NullString
	}
	err = h.db.QueryRow(
		"SELECT description_summary, diff_summary, diff_explanation, comments_summary, discussion_summary, merge_reason, close_reason FROM pull_request_summaries WHERE pr_id = ?",
		id,
	).Scan(&summary.DescriptionSummary, &summary.DiffSummary, &summary.DiffExplanation, &summary.CommentsSummary, &summary.DiscussionSummary, &summary.MergeReason, &summary.CloseReason)

	if err != nil && err != sql.ErrNoRows {
		h.logger.Warn("PR要約データの取得に失敗しました", zap.Error(err))
	}

	summaryMap := make(map[string]interface{})
	if summary.DescriptionSummary.Valid {
		summaryMap["description_summary"] = summary.DescriptionSummary.String
	}
	if summary.DiffSummary.Valid {
		summaryMap["diff_summary"] = summary.DiffSummary.String
	}
	if summary.DiffExplanation.Valid {
		summaryMap["diff_explanation"] = summary.DiffExplanation.String
	}
	if summary.CommentsSummary.Valid {
		summaryMap["comments_summary"] = summary.CommentsSummary.String
	}
	if summary.DiscussionSummary.Valid {
		summaryMap["discussion_summary"] = summary.DiscussionSummary.String
	}
	if summary.MergeReason.Valid {
		summaryMap["merge_reason"] = summary.MergeReason.String
	}
	if summary.CloseReason.Valid {
		summaryMap["close_reason"] = summary.CloseReason.String
	}
	if len(summaryMap) > 0 {
		prMap["summary"] = summaryMap
	}

	// 差分データの取得
	diffs := []map[string]interface{}{}
	diffRows, err := h.db.Query(
		"SELECT diff_text, file_path FROM pull_request_diffs WHERE pr_id = ? ORDER BY id",
		id,
	)
	if err == nil {
		defer diffRows.Close()
		for diffRows.Next() {
			var diffText, filePath string
			if err := diffRows.Scan(&diffText, &filePath); err == nil {
				diffs = append(diffs, map[string]interface{}{
					"diff_text": diffText,
					"file_path":  filePath,
				})
			}
		}
	}
	if len(diffs) > 0 {
		prMap["diffs"] = diffs
	}

	// コメント一覧の取得
	comments := []map[string]interface{}{}
	commentRows, err := h.db.Query(
		"SELECT github_id, body, author, created_at, updated_at FROM pull_request_comments WHERE pr_id = ? ORDER BY created_at",
		id,
	)
	if err == nil {
		defer commentRows.Close()
		for commentRows.Next() {
			var comment struct {
				GitHubID  int
				Body      string
				Author    string
				CreatedAt string
				UpdatedAt string
			}
			if err := commentRows.Scan(&comment.GitHubID, &comment.Body, &comment.Author, &comment.CreatedAt, &comment.UpdatedAt); err == nil {
				comments = append(comments, map[string]interface{}{
					"github_id":  comment.GitHubID,
					"body":       comment.Body,
					"author":     comment.Author,
					"created_at": comment.CreatedAt,
					"updated_at": comment.UpdatedAt,
				})
			}
		}
	}
	if len(comments) > 0 {
		prMap["comments"] = comments
	}

	c.JSON(http.StatusOK, prMap)
}

func (h *Handlers) GetIssues(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	state := c.Query("state")
	author := c.Query("author")
	search := c.Query("search")

	offset := (page - 1) * limit

	query := "SELECT id, repository_id, github_id, title, body, state, author, created_at, updated_at, closed_at, url FROM issues WHERE 1=1"
	args := []interface{}{}

	if state != "" {
		query += " AND state = ?"
		args = append(args, state)
	}
	if author != "" {
		query += " AND author = ?"
		args = append(args, author)
	}
	if search != "" {
		query += " AND (title LIKE ? OR body LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.Error("Issue一覧の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Issue一覧の取得に失敗しました"})
		return
	}
	defer rows.Close()

	var issues []map[string]interface{}
	for rows.Next() {
		var issue struct {
			ID        int
			RepoID    int
			GitHubID  int
			Title     string
			Body      string
			State     string
			Author    string
			CreatedAt string
			UpdatedAt string
			ClosedAt  sql.NullString
			URL       string
		}
		err := rows.Scan(&issue.ID, &issue.RepoID, &issue.GitHubID, &issue.Title, &issue.Body, &issue.State, &issue.Author, &issue.CreatedAt, &issue.UpdatedAt, &issue.ClosedAt, &issue.URL)
		if err != nil {
			continue
		}

		issueMap := map[string]interface{}{
			"id":         issue.ID,
			"repository_id": issue.RepoID,
			"github_id":  issue.GitHubID,
			"title":      issue.Title,
			"body":       issue.Body,
			"state":      issue.State,
			"author":     issue.Author,
			"created_at": issue.CreatedAt,
			"updated_at": issue.UpdatedAt,
			"url":        issue.URL,
		}
		if issue.ClosedAt.Valid {
			issueMap["closed_at"] = issue.ClosedAt.String
		}

		issues = append(issues, issueMap)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": issues,
		"page": page,
		"limit": limit,
	})
}

func (h *Handlers) GetIssue(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なIDです"})
		return
	}

	var issue struct {
		ID        int
		RepoID    int
		GitHubID  int
		Title     string
		Body      string
		State     string
		Author    string
		CreatedAt string
		UpdatedAt string
		ClosedAt  sql.NullString
		URL       string
	}

	err = h.db.QueryRow(
		"SELECT id, repository_id, github_id, title, body, state, author, created_at, updated_at, closed_at, url FROM issues WHERE id = ?",
		id,
	).Scan(&issue.ID, &issue.RepoID, &issue.GitHubID, &issue.Title, &issue.Body, &issue.State, &issue.Author, &issue.CreatedAt, &issue.UpdatedAt, &issue.ClosedAt, &issue.URL)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Issueが見つかりません"})
		return
	}
	if err != nil {
		h.logger.Error("Issue詳細の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Issue詳細の取得に失敗しました"})
		return
	}

	issueMap := map[string]interface{}{
		"id":            issue.ID,
		"repository_id": issue.RepoID,
		"github_id":     issue.GitHubID,
		"title":         issue.Title,
		"body":          issue.Body,
		"state":         issue.State,
		"author":        issue.Author,
		"created_at":    issue.CreatedAt,
		"updated_at":    issue.UpdatedAt,
		"url":           issue.URL,
	}
	if issue.ClosedAt.Valid {
		issueMap["closed_at"] = issue.ClosedAt.String
	}

	// 要約データの取得
	var summary struct {
		DescriptionSummary sql.NullString
		CommentsSummary    sql.NullString
		DiscussionSummary  sql.NullString
	}
	err = h.db.QueryRow(
		"SELECT description_summary, comments_summary, discussion_summary FROM issue_summaries WHERE issue_id = ?",
		id,
	).Scan(&summary.DescriptionSummary, &summary.CommentsSummary, &summary.DiscussionSummary)

	if err != nil && err != sql.ErrNoRows {
		h.logger.Warn("Issue要約データの取得に失敗しました", zap.Error(err))
	}

	summaryMap := make(map[string]interface{})
	if summary.DescriptionSummary.Valid {
		summaryMap["description_summary"] = summary.DescriptionSummary.String
	}
	if summary.CommentsSummary.Valid {
		summaryMap["comments_summary"] = summary.CommentsSummary.String
	}
	if summary.DiscussionSummary.Valid {
		summaryMap["discussion_summary"] = summary.DiscussionSummary.String
	}
	if len(summaryMap) > 0 {
		issueMap["summary"] = summaryMap
	}

	// コメント一覧の取得
	comments := []map[string]interface{}{}
	commentRows, err := h.db.Query(
		"SELECT github_id, body, author, created_at, updated_at FROM issue_comments WHERE issue_id = ? ORDER BY created_at",
		id,
	)
	if err == nil {
		defer commentRows.Close()
		for commentRows.Next() {
			var comment struct {
				GitHubID  int
				Body      string
				Author    string
				CreatedAt string
				UpdatedAt string
			}
			if err := commentRows.Scan(&comment.GitHubID, &comment.Body, &comment.Author, &comment.CreatedAt, &comment.UpdatedAt); err == nil {
				comments = append(comments, map[string]interface{}{
					"github_id":  comment.GitHubID,
					"body":       comment.Body,
					"author":     comment.Author,
					"created_at": comment.CreatedAt,
					"updated_at": comment.UpdatedAt,
				})
			}
		}
	}
	if len(comments) > 0 {
		issueMap["comments"] = comments
	}

	c.JSON(http.StatusOK, issueMap)
}
