package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetChanges returns a list of Changes (Gerrit CLs)
func (h *Handlers) GetChanges(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	// Cap limit at 100 to prevent excessive database load and response sizes
	if limit > 100 {
		limit = 100
	}

	status := c.Query("status")
	branch := c.Query("branch")
	author := c.Query("author")

	offset := (page - 1) * limit

	// Build query
	query := "SELECT id, repository_id, change_number, change_id, project, branch, subject, message, status, owner, created_at, updated_at, submitted_at, url FROM changes WHERE 1=1"
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if branch != "" {
		query += " AND branch = ?"
		args = append(args, branch)
	}
	if author != "" {
		query += " AND owner = ?"
		args = append(args, author)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.Error("Change一覧の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Change一覧の取得に失敗しました"})
		return
	}
	defer rows.Close()

	// Return [] (not null) when empty for better frontend UX.
	changes := make([]map[string]interface{}, 0)
	for rows.Next() {
		var change struct {
			ID          int
			RepoID      int
			ChangeNumber int
			ChangeID    string
			Project     string
			Branch      string
			Subject     string
			Message     sql.NullString
			Status      string
			Owner       string
			CreatedAt   string
			UpdatedAt   string
			SubmittedAt sql.NullString
			URL         string
		}
		err := rows.Scan(&change.ID, &change.RepoID, &change.ChangeNumber, &change.ChangeID, &change.Project, &change.Branch,
			&change.Subject, &change.Message, &change.Status, &change.Owner, &change.CreatedAt, &change.UpdatedAt,
			&change.SubmittedAt, &change.URL)
		if err != nil {
			continue
		}

		changeMap := map[string]interface{}{
			"id":             change.ID,
			"repository_id":  change.RepoID,
			"change_number":  change.ChangeNumber,
			"change_id":      change.ChangeID,
			"project":        change.Project,
			"branch":         change.Branch,
			"subject":        change.Subject,
			"status":         change.Status,
			"owner":          change.Owner,
			"created_at":     change.CreatedAt,
			"updated_at":     change.UpdatedAt,
			"url":            change.URL,
		}
		if change.Message.Valid {
			changeMap["message"] = change.Message.String
		}
		if change.SubmittedAt.Valid {
			changeMap["submitted_at"] = change.SubmittedAt.String
		}

		changes = append(changes, changeMap)
	}

	if err := rows.Err(); err != nil {
		h.logger.Error("Change一覧の取得中にエラーが発生しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Change一覧の取得中にエラーが発生しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  changes,
		"page":  page,
		"limit": limit,
	})
}

// GetChange returns a single Change with details
func (h *Handlers) GetChange(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なIDです"})
		return
	}

	var change struct {
		ID          int
		RepoID      int
		ChangeNumber int
		ChangeID    string
		Project     string
		Branch      string
		Subject     string
		Message     sql.NullString
		Status      string
		Owner       string
		CreatedAt   string
		UpdatedAt   string
		SubmittedAt sql.NullString
		URL         string
	}

	err = h.db.QueryRow(
		"SELECT id, repository_id, change_number, change_id, project, branch, subject, message, status, owner, created_at, updated_at, submitted_at, url FROM changes WHERE id = ?",
		id,
	).Scan(&change.ID, &change.RepoID, &change.ChangeNumber, &change.ChangeID, &change.Project, &change.Branch,
		&change.Subject, &change.Message, &change.Status, &change.Owner, &change.CreatedAt, &change.UpdatedAt,
		&change.SubmittedAt, &change.URL)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Changeが見つかりません"})
		return
	}
	if err != nil {
		h.logger.Error("Change詳細の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Change詳細の取得に失敗しました"})
		return
	}

	changeMap := map[string]interface{}{
		"id":             change.ID,
		"repository_id":  change.RepoID,
		"change_number":  change.ChangeNumber,
		"change_id":      change.ChangeID,
		"project":        change.Project,
		"branch":         change.Branch,
		"subject":       change.Subject,
		"status":         change.Status,
		"owner":          change.Owner,
		"created_at":     change.CreatedAt,
		"updated_at":     change.UpdatedAt,
		"url":            change.URL,
	}
	if change.Message.Valid {
		changeMap["message"] = change.Message.String
	}
	if change.SubmittedAt.Valid {
		changeMap["submitted_at"] = change.SubmittedAt.String
	}

	// Get revisions
	revisions := []map[string]interface{}{}
	revRows, err := h.db.Query(
		"SELECT id, patch_set_number, revision_sha, uploader, created_at, kind, commit_message FROM revisions WHERE change_id = ? ORDER BY patch_set_number",
		id,
	)
	if err == nil {
		defer revRows.Close()
		for revRows.Next() {
			var rev struct {
				ID            int
				PatchSetNumber int
				RevisionSHA   string
				Uploader      string
				CreatedAt     string
				Kind          sql.NullString
				CommitMessage sql.NullString
			}
			if err := revRows.Scan(&rev.ID, &rev.PatchSetNumber, &rev.RevisionSHA, &rev.Uploader, &rev.CreatedAt, &rev.Kind, &rev.CommitMessage); err == nil {
				revMap := map[string]interface{}{
					"id":               rev.ID,
					"patch_set_number": rev.PatchSetNumber,
					"revision_sha":     rev.RevisionSHA,
					"uploader":         rev.Uploader,
					"created_at":       rev.CreatedAt,
				}
				if rev.Kind.Valid {
					revMap["kind"] = rev.Kind.String
				}
				if rev.CommitMessage.Valid {
					revMap["commit_message"] = rev.CommitMessage.String
				}
				revisions = append(revisions, revMap)
			}
		}
	}
	if len(revisions) > 0 {
		changeMap["revisions"] = revisions
	}

	// Get files for latest revision
	if len(revisions) > 0 {
		latestRevID := revisions[len(revisions)-1]["id"].(int)
		files := []map[string]interface{}{}
		fileRows, err := h.db.Query(
			"SELECT file_path, status, old_path, lines_inserted, lines_deleted, size_delta, size, binary FROM change_files WHERE revision_id = ? ORDER BY file_path",
			latestRevID,
		)
		if err == nil {
			defer fileRows.Close()
			for fileRows.Next() {
				var file struct {
					FilePath      string
					Status        sql.NullString
					OldPath       sql.NullString
					LinesInserted int
					LinesDeleted  int
					SizeDelta     int64
					Size          int64
					Binary        bool
				}
				if err := fileRows.Scan(&file.FilePath, &file.Status, &file.OldPath, &file.LinesInserted, &file.LinesDeleted, &file.SizeDelta, &file.Size, &file.Binary); err == nil {
					fileMap := map[string]interface{}{
						"file_path":      file.FilePath,
						"lines_inserted": file.LinesInserted,
						"lines_deleted":  file.LinesDeleted,
						"size_delta":     file.SizeDelta,
						"size":           file.Size,
						"binary":         file.Binary,
					}
					if file.Status.Valid {
						fileMap["status"] = file.Status.String
					}
					if file.OldPath.Valid {
						fileMap["old_path"] = file.OldPath.String
					}
					files = append(files, fileMap)
				}
			}
		}
		if len(files) > 0 {
			changeMap["files"] = files
		}
	}

	// Get comments
	comments := []map[string]interface{}{}
	commentRows, err := h.db.Query(
		"SELECT comment_id, file_path, line, patch_set_number, message, author, created_at, updated_at, in_reply_to, unresolved FROM change_comments WHERE change_id = ? ORDER BY created_at",
		id,
	)
	if err == nil {
		defer commentRows.Close()
		for commentRows.Next() {
			var comment struct {
				CommentID      string
				FilePath       sql.NullString
				Line           sql.NullInt64
				PatchSetNumber int
				Message        string
				Author         string
				CreatedAt      string
				UpdatedAt      string
				InReplyTo      sql.NullString
				Unresolved     bool
			}
			if err := commentRows.Scan(&comment.CommentID, &comment.FilePath, &comment.Line, &comment.PatchSetNumber,
				&comment.Message, &comment.Author, &comment.CreatedAt, &comment.UpdatedAt, &comment.InReplyTo, &comment.Unresolved); err == nil {
				commentMap := map[string]interface{}{
					"comment_id":       comment.CommentID,
					"patch_set_number": comment.PatchSetNumber,
					"message":          comment.Message,
					"author":           comment.Author,
					"created_at":       comment.CreatedAt,
					"updated_at":       comment.UpdatedAt,
					"unresolved":       comment.Unresolved,
				}
				if comment.FilePath.Valid {
					commentMap["file_path"] = comment.FilePath.String
				}
				if comment.Line.Valid {
					commentMap["line"] = comment.Line.Int64
				}
				if comment.InReplyTo.Valid {
					commentMap["in_reply_to"] = comment.InReplyTo.String
				}
				comments = append(comments, commentMap)
			}
		}
	}
	if len(comments) > 0 {
		changeMap["comments"] = comments
	}

	// Get labels
	labels := []map[string]interface{}{}
	labelRows, err := h.db.Query(
		"SELECT label_name, account, value, date FROM change_labels WHERE change_id = ? ORDER BY label_name, date",
		id,
	)
	if err == nil {
		defer labelRows.Close()
		for labelRows.Next() {
			var label struct {
				LabelName string
				Account   string
				Value     int
				Date      string
			}
			if err := labelRows.Scan(&label.LabelName, &label.Account, &label.Value, &label.Date); err == nil {
				labels = append(labels, map[string]interface{}{
					"label_name": label.LabelName,
					"account":    label.Account,
					"value":      label.Value,
					"date":       label.Date,
				})
			}
		}
	}
	if len(labels) > 0 {
		changeMap["labels"] = labels
	}

	// Get messages
	messages := []map[string]interface{}{}
	msgRows, err := h.db.Query(
		"SELECT message_id, author, message, date, revision_number FROM change_messages WHERE change_id = ? ORDER BY date",
		id,
	)
	if err == nil {
		defer msgRows.Close()
		for msgRows.Next() {
			var msg struct {
				MessageID      string
				Author         string
				Message        string
				Date           string
				RevisionNumber sql.NullInt64
			}
			if err := msgRows.Scan(&msg.MessageID, &msg.Author, &msg.Message, &msg.Date, &msg.RevisionNumber); err == nil {
				msgMap := map[string]interface{}{
					"message_id": msg.MessageID,
					"author":     msg.Author,
					"message":    msg.Message,
					"date":       msg.Date,
				}
				if msg.RevisionNumber.Valid {
					msgMap["revision_number"] = msg.RevisionNumber.Int64
				}
				messages = append(messages, msgMap)
			}
		}
	}
	if len(messages) > 0 {
		changeMap["messages"] = messages
	}

	c.JSON(http.StatusOK, changeMap)
}
