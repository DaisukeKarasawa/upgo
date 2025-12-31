package api

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChangeHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewChangeHandlers(db *sql.DB, logger *zap.Logger) *ChangeHandlers {
	return &ChangeHandlers{
		db:     db,
		logger: logger,
	}
}

func (h *ChangeHandlers) GetChanges(c *gin.Context) {
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
	if limit > 100 {
		limit = 100
	}

	status := c.Query("status")
	branch := c.Query("branch")

	offset := (page - 1) * limit

	query := `SELECT id, change_id, change_number, project, branch, status, subject, 
		owner_name, owner_email, created, updated, submitted, last_synced_at 
		FROM changes WHERE 1=1`
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if branch != "" {
		query += " AND branch = ?"
		args = append(args, branch)
	}

	query += " ORDER BY updated DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		h.logger.Error("変更一覧の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "変更一覧の取得に失敗しました"})
		return
	}
	defer rows.Close()

	var changes []map[string]interface{}
	for rows.Next() {
		var change struct {
			ID           int
			ChangeID     string
			ChangeNumber int
			Project      string
			Branch       string
			Status       string
			Subject      string
			OwnerName    string
			OwnerEmail   sql.NullString
			Created      string
			Updated      string
			Submitted    sql.NullString
			LastSyncedAt sql.NullString
		}
		err := rows.Scan(&change.ID, &change.ChangeID, &change.ChangeNumber, &change.Project,
			&change.Branch, &change.Status, &change.Subject, &change.OwnerName, &change.OwnerEmail,
			&change.Created, &change.Updated, &change.Submitted, &change.LastSyncedAt)
		if err != nil {
			continue
		}

		changeMap := map[string]interface{}{
			"id":            change.ID,
			"change_id":     change.ChangeID,
			"change_number": change.ChangeNumber,
			"project":       change.Project,
			"branch":        change.Branch,
			"status":        change.Status,
			"subject":       change.Subject,
			"owner_name":    change.OwnerName,
			"created":       change.Created,
			"updated":       change.Updated,
		}
		if change.OwnerEmail.Valid {
			changeMap["owner_email"] = change.OwnerEmail.String
		}
		if change.Submitted.Valid {
			changeMap["submitted"] = change.Submitted.String
		}
		if change.LastSyncedAt.Valid {
			changeMap["last_synced_at"] = change.LastSyncedAt.String
		}

		changes = append(changes, changeMap)
	}

	if err := rows.Err(); err != nil {
		h.logger.Error("変更一覧の取得中にエラーが発生しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "変更一覧の取得中にエラーが発生しました"})
		return
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM changes WHERE 1=1"
	countArgs := []interface{}{}
	if status != "" {
		countQuery += " AND status = ?"
		countArgs = append(countArgs, status)
	}
	if branch != "" {
		countQuery += " AND branch = ?"
		countArgs = append(countArgs, branch)
	}
	h.db.QueryRow(countQuery, countArgs...).Scan(&total)

	c.JSON(http.StatusOK, gin.H{
		"data":  changes,
		"page":  page,
		"limit": limit,
		"total": total,
	})
}

func (h *ChangeHandlers) GetChange(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なIDです"})
		return
	}

	var change struct {
		ID           int
		ChangeID     string
		ChangeNumber int
		Project      string
		Branch       string
		Status       string
		Subject      string
		Message      sql.NullString
		OwnerName    string
		OwnerEmail   sql.NullString
		Created      string
		Updated      string
		Submitted    sql.NullString
		LastSyncedAt sql.NullString
	}

	err = h.db.QueryRow(`
		SELECT id, change_id, change_number, project, branch, status, subject, message,
			owner_name, owner_email, created, updated, submitted, last_synced_at
		FROM changes WHERE id = ?`, id).Scan(
		&change.ID, &change.ChangeID, &change.ChangeNumber, &change.Project,
		&change.Branch, &change.Status, &change.Subject, &change.Message,
		&change.OwnerName, &change.OwnerEmail, &change.Created, &change.Updated,
		&change.Submitted, &change.LastSyncedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "変更が見つかりません"})
		return
	}
	if err != nil {
		h.logger.Error("変更詳細の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "変更詳細の取得に失敗しました"})
		return
	}

	changeMap := map[string]interface{}{
		"id":            change.ID,
		"change_id":     change.ChangeID,
		"change_number": change.ChangeNumber,
		"project":       change.Project,
		"branch":        change.Branch,
		"status":        change.Status,
		"subject":       change.Subject,
		"owner_name":    change.OwnerName,
		"created":       change.Created,
		"updated":       change.Updated,
	}
	if change.Message.Valid {
		changeMap["message"] = change.Message.String
	}
	if change.OwnerEmail.Valid {
		changeMap["owner_email"] = change.OwnerEmail.String
	}
	if change.Submitted.Valid {
		changeMap["submitted"] = change.Submitted.String
	}
	if change.LastSyncedAt.Valid {
		changeMap["last_synced_at"] = change.LastSyncedAt.String
	}

	revisions := []map[string]interface{}{}
	revisionRows, err := h.db.Query(`
		SELECT id, revision_id, patchset_num, uploader_name, uploader_email, created, commit_message
		FROM revisions WHERE change_db_id = ? ORDER BY patchset_num DESC`, id)
	if err == nil {
		defer revisionRows.Close()
		for revisionRows.Next() {
			var rev struct {
				ID            int
				RevisionID    string
				PatchsetNum   int
				UploaderName  sql.NullString
				UploaderEmail sql.NullString
				Created       string
				CommitMessage sql.NullString
			}
			if err := revisionRows.Scan(&rev.ID, &rev.RevisionID, &rev.PatchsetNum,
				&rev.UploaderName, &rev.UploaderEmail, &rev.Created, &rev.CommitMessage); err == nil {
				revMap := map[string]interface{}{
					"id":           rev.ID,
					"revision_id":  rev.RevisionID,
					"patchset_num": rev.PatchsetNum,
					"created":      rev.Created,
				}
				if rev.UploaderName.Valid {
					revMap["uploader_name"] = rev.UploaderName.String
				}
				if rev.UploaderEmail.Valid {
					revMap["uploader_email"] = rev.UploaderEmail.String
				}
				if rev.CommitMessage.Valid {
					revMap["commit_message"] = rev.CommitMessage.String
				}

				files := []map[string]interface{}{}
				fileRows, err := h.db.Query(`
					SELECT id, file_path, status, lines_inserted, lines_deleted, size_delta
					FROM files WHERE revision_db_id = ? ORDER BY file_path`, rev.ID)
				if err == nil {
					defer fileRows.Close()
					for fileRows.Next() {
						var file struct {
							ID            int
							FilePath      string
							Status        sql.NullString
							LinesInserted int
							LinesDeleted  int
							SizeDelta     int
						}
						if err := fileRows.Scan(&file.ID, &file.FilePath, &file.Status,
							&file.LinesInserted, &file.LinesDeleted, &file.SizeDelta); err == nil {
							fileMap := map[string]interface{}{
								"id":             file.ID,
								"file_path":      file.FilePath,
								"lines_inserted": file.LinesInserted,
								"lines_deleted":  file.LinesDeleted,
								"size_delta":     file.SizeDelta,
							}
							if file.Status.Valid {
								fileMap["status"] = file.Status.String
							}
							files = append(files, fileMap)
						}
					}
				}
				revMap["files"] = files

				revisions = append(revisions, revMap)
			}
		}
	}
	changeMap["revisions"] = revisions

	comments := []map[string]interface{}{}
	commentRows, err := h.db.Query(`
		SELECT id, comment_id, file_path, line, author_name, author_email, message, created, updated, in_reply_to, unresolved
		FROM comments WHERE change_db_id = ? ORDER BY created`, id)
	if err == nil {
		defer commentRows.Close()
		for commentRows.Next() {
			var comment struct {
				ID          int
				CommentID   string
				FilePath    sql.NullString
				Line        sql.NullInt64
				AuthorName  string
				AuthorEmail sql.NullString
				Message     string
				Created     string
				Updated     string
				InReplyTo   sql.NullString
				Unresolved  int
			}
			if err := commentRows.Scan(&comment.ID, &comment.CommentID, &comment.FilePath,
				&comment.Line, &comment.AuthorName, &comment.AuthorEmail, &comment.Message,
				&comment.Created, &comment.Updated, &comment.InReplyTo, &comment.Unresolved); err == nil {
				commentMap := map[string]interface{}{
					"id":          comment.ID,
					"comment_id":  comment.CommentID,
					"author_name": comment.AuthorName,
					"message":     comment.Message,
					"created":     comment.Created,
					"updated":     comment.Updated,
					"unresolved":  comment.Unresolved == 1,
				}
				if comment.FilePath.Valid {
					commentMap["file_path"] = comment.FilePath.String
				}
				if comment.Line.Valid {
					commentMap["line"] = comment.Line.Int64
				}
				if comment.AuthorEmail.Valid {
					commentMap["author_email"] = comment.AuthorEmail.String
				}
				if comment.InReplyTo.Valid {
					commentMap["in_reply_to"] = comment.InReplyTo.String
				}
				comments = append(comments, commentMap)
			}
		}
	}
	changeMap["comments"] = comments

	labels := []map[string]interface{}{}
	labelRows, err := h.db.Query(`
		SELECT id, label_name, value, account_name, account_email, granted_on
		FROM labels WHERE change_db_id = ? ORDER BY label_name, granted_on`, id)
	if err == nil {
		defer labelRows.Close()
		for labelRows.Next() {
			var label struct {
				ID           int
				LabelName    string
				Value        int
				AccountName  sql.NullString
				AccountEmail sql.NullString
				GrantedOn    string
			}
			if err := labelRows.Scan(&label.ID, &label.LabelName, &label.Value,
				&label.AccountName, &label.AccountEmail, &label.GrantedOn); err == nil {
				labelMap := map[string]interface{}{
					"id":         label.ID,
					"label_name": label.LabelName,
					"value":      label.Value,
					"granted_on": label.GrantedOn,
				}
				if label.AccountName.Valid {
					labelMap["account_name"] = label.AccountName.String
				}
				if label.AccountEmail.Valid {
					labelMap["account_email"] = label.AccountEmail.String
				}
				labels = append(labels, labelMap)
			}
		}
	}
	changeMap["labels"] = labels

	messages := []map[string]interface{}{}
	messageRows, err := h.db.Query(`
		SELECT id, message_id, author_name, author_email, message, date, revision_number
		FROM messages WHERE change_db_id = ? ORDER BY date`, id)
	if err == nil {
		defer messageRows.Close()
		for messageRows.Next() {
			var msg struct {
				ID             int
				MessageID      string
				AuthorName     sql.NullString
				AuthorEmail    sql.NullString
				Message        string
				Date           string
				RevisionNumber sql.NullInt64
			}
			if err := messageRows.Scan(&msg.ID, &msg.MessageID, &msg.AuthorName,
				&msg.AuthorEmail, &msg.Message, &msg.Date, &msg.RevisionNumber); err == nil {
				msgMap := map[string]interface{}{
					"id":         msg.ID,
					"message_id": msg.MessageID,
					"message":    msg.Message,
					"date":       msg.Date,
				}
				if msg.AuthorName.Valid {
					msgMap["author_name"] = msg.AuthorName.String
				}
				if msg.AuthorEmail.Valid {
					msgMap["author_email"] = msg.AuthorEmail.String
				}
				if msg.RevisionNumber.Valid {
					msgMap["revision_number"] = msg.RevisionNumber.Int64
				}
				messages = append(messages, msgMap)
			}
		}
	}
	changeMap["messages"] = messages

	c.JSON(http.StatusOK, changeMap)
}

func (h *ChangeHandlers) GetBranches(c *gin.Context) {
	rows, err := h.db.Query("SELECT DISTINCT branch FROM changes ORDER BY branch")
	if err != nil {
		h.logger.Error("ブランチ一覧の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ブランチ一覧の取得に失敗しました"})
		return
	}
	defer rows.Close()

	var branches []string
	for rows.Next() {
		var branch string
		if err := rows.Scan(&branch); err == nil {
			branches = append(branches, branch)
		}
	}

	c.JSON(http.StatusOK, gin.H{"branches": branches})
}

func (h *ChangeHandlers) GetStatuses(c *gin.Context) {
	rows, err := h.db.Query("SELECT DISTINCT status FROM changes ORDER BY status")
	if err != nil {
		h.logger.Error("ステータス一覧の取得に失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ステータス一覧の取得に失敗しました"})
		return
	}
	defer rows.Close()

	var statuses []string
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err == nil {
			statuses = append(statuses, status)
		}
	}

	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}
