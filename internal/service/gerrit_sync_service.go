package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"upgo/internal/config"
	"upgo/internal/gerrit"

	grt "golang.org/x/build/gerrit"
	"go.uber.org/zap"
)

type GerritSyncService struct {
	db           *sql.DB
	gerritClient *gerrit.Client
	logger       *zap.Logger
	cfg          *config.Config
}

func NewGerritSyncService(
	db *sql.DB,
	gerritClient *gerrit.Client,
	logger *zap.Logger,
	cfg *config.Config,
) *GerritSyncService {
	return &GerritSyncService{
		db:           db,
		gerritClient: gerritClient,
		logger:       logger,
		cfg:          cfg,
	}
}

func (s *GerritSyncService) Sync(ctx context.Context) error {
	syncStart := time.Now()
	s.logger.Info("Gerrit同期を開始しました")

	sinceTime := time.Now().AddDate(0, 0, -s.cfg.Sync.UpdatedDays)

	opts := gerrit.QueryOptions{
		Project:  s.cfg.Gerrit.Project,
		Branches: s.cfg.Gerrit.Branches,
		Status:   s.cfg.Gerrit.Status,
		After:    sinceTime,
		Limit:    100,
	}

	var totalChanges int
	var start int

	for {
		opts.Start = start
		changes, err := s.gerritClient.QueryChanges(ctx, opts)
		if err != nil {
			return fmt.Errorf("変更の取得に失敗しました: %w", err)
		}

		if len(changes) == 0 {
			break
		}

		for _, change := range changes {
			if err := s.saveChange(ctx, change); err != nil {
				s.logger.Error("変更の保存に失敗しました",
					zap.Int("change_number", change.ChangeNumber),
					zap.Error(err))
				continue
			}
			totalChanges++
		}

		if len(changes) < opts.Limit {
			break
		}
		start += len(changes)
	}

	totalDuration := time.Since(syncStart)
	s.logger.Info("Gerrit同期が完了しました",
		zap.Int("total_changes", totalChanges),
		zap.Duration("duration", totalDuration))

	return nil
}

func (s *GerritSyncService) SyncChangeByNumber(ctx context.Context, changeNumber int) error {
	s.logger.Info("変更の同期を開始しました", zap.Int("change_number", changeNumber))

	changeID := fmt.Sprintf("%s~%d", s.cfg.Gerrit.Project, changeNumber)
	change, err := s.gerritClient.GetChangeDetail(ctx, changeID)
	if err != nil {
		return fmt.Errorf("変更詳細の取得に失敗しました: %w", err)
	}

	if err := s.saveChange(ctx, change); err != nil {
		return fmt.Errorf("変更の保存に失敗しました: %w", err)
	}

	if err := s.syncChangeDetails(ctx, changeID, change); err != nil {
		s.logger.Warn("変更詳細の同期に失敗しました", zap.Error(err))
	}

	s.logger.Info("変更の同期が完了しました", zap.Int("change_number", changeNumber))
	return nil
}

func (s *GerritSyncService) saveChange(ctx context.Context, change *grt.ChangeInfo) error {
	var submitted interface{}
	if change.Submitted != "" {
		t, err := parseGerritTime(change.Submitted)
		if err == nil {
			submitted = t
		}
	}

	created, _ := parseGerritTime(change.Created)
	updated, _ := parseGerritTime(change.Updated)

	ownerName := ""
	ownerEmail := ""
	if change.Owner != nil {
		ownerName = change.Owner.Name
		ownerEmail = change.Owner.Email
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO changes (change_id, change_number, project, branch, status, subject, message, owner_name, owner_email, created, updated, submitted, last_synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(change_id) DO UPDATE SET
			status = excluded.status,
			subject = excluded.subject,
			message = excluded.message,
			updated = excluded.updated,
			submitted = excluded.submitted,
			last_synced_at = excluded.last_synced_at
	`, change.ChangeID, change.ChangeNumber, change.Project, change.Branch, change.Status,
		change.Subject, "", ownerName, ownerEmail, created, updated, submitted, time.Now())

	if err != nil {
		return fmt.Errorf("変更の保存に失敗しました: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	s.logger.Debug("変更を保存しました",
		zap.Int("change_number", change.ChangeNumber),
		zap.Int64("rows_affected", rowsAffected))

	return nil
}

func (s *GerritSyncService) syncChangeDetails(ctx context.Context, changeID string, change *grt.ChangeInfo) error {
	var changeDBID int
	err := s.db.QueryRowContext(ctx,
		"SELECT id FROM changes WHERE change_id = ?", change.ChangeID).Scan(&changeDBID)
	if err != nil {
		return fmt.Errorf("変更IDの取得に失敗しました: %w", err)
	}

	if err := s.syncRevisions(ctx, changeDBID, changeID, change); err != nil {
		s.logger.Warn("リビジョンの同期に失敗しました", zap.Error(err))
	}

	if err := s.syncComments(ctx, changeDBID, changeID); err != nil {
		s.logger.Warn("コメントの同期に失敗しました", zap.Error(err))
	}

	if err := s.syncMessages(ctx, changeDBID, change); err != nil {
		s.logger.Warn("メッセージの同期に失敗しました", zap.Error(err))
	}

	if err := s.syncLabels(ctx, changeDBID, change); err != nil {
		s.logger.Warn("ラベルの同期に失敗しました", zap.Error(err))
	}

	return nil
}

func (s *GerritSyncService) syncRevisions(ctx context.Context, changeDBID int, changeID string, change *grt.ChangeInfo) error {
	for revisionID, revision := range change.Revisions {
		uploaderName := ""
		uploaderEmail := ""
		if revision.Uploader != nil {
			uploaderName = revision.Uploader.Name
			uploaderEmail = revision.Uploader.Email
		}

		created, _ := parseGerritTime(revision.Created)
		commitMessage := ""
		if revision.Commit != nil {
			commitMessage = revision.Commit.Message
		}

		result, err := s.db.ExecContext(ctx, `
			INSERT INTO revisions (change_db_id, revision_id, patchset_num, uploader_name, uploader_email, created, commit_message)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(change_db_id, patchset_num) DO UPDATE SET
				revision_id = excluded.revision_id,
				uploader_name = excluded.uploader_name,
				uploader_email = excluded.uploader_email,
				commit_message = excluded.commit_message
		`, changeDBID, revisionID, revision.PatchSetNumber, uploaderName, uploaderEmail, created, commitMessage)

		if err != nil {
			s.logger.Warn("リビジョンの保存に失敗しました",
				zap.String("revision_id", revisionID),
				zap.Error(err))
			continue
		}

		var revisionDBID int64
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			revisionDBID, _ = result.LastInsertId()
			if revisionDBID == 0 {
				s.db.QueryRowContext(ctx,
					"SELECT id FROM revisions WHERE change_db_id = ? AND patchset_num = ?",
					changeDBID, revision.PatchSetNumber).Scan(&revisionDBID)
			}
		}

		if revisionDBID > 0 {
			if err := s.syncFiles(ctx, int(revisionDBID), changeID, revisionID); err != nil {
				s.logger.Warn("ファイルの同期に失敗しました",
					zap.String("revision_id", revisionID),
					zap.Error(err))
			}
		}
	}

	return nil
}

func (s *GerritSyncService) syncFiles(ctx context.Context, revisionDBID int, changeID, revisionID string) error {
	files, err := s.gerritClient.ListFiles(ctx, changeID, revisionID)
	if err != nil {
		return fmt.Errorf("ファイル一覧の取得に失敗しました: %w", err)
	}

	for filePath, fileInfo := range files {
		if filePath == "/COMMIT_MSG" || filePath == "/MERGE_LIST" {
			continue
		}

		if s.shouldExcludeFile(filePath) {
			continue
		}

		status := fileInfo.Status
		if status == "" {
			status = "M"
		}

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO files (revision_db_id, file_path, status, lines_inserted, lines_deleted, size_delta)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(revision_db_id, file_path) DO UPDATE SET
				status = excluded.status,
				lines_inserted = excluded.lines_inserted,
				lines_deleted = excluded.lines_deleted,
				size_delta = excluded.size_delta
		`, revisionDBID, filePath, status, fileInfo.LinesInserted, fileInfo.LinesDeleted, fileInfo.SizeDelta)

		if err != nil {
			s.logger.Warn("ファイルの保存に失敗しました",
				zap.String("file_path", filePath),
				zap.Error(err))
		}
	}

	return nil
}

func (s *GerritSyncService) shouldExcludeFile(filePath string) bool {
	for _, excludePath := range s.cfg.Diff.ExcludePaths {
		if strings.HasPrefix(filePath, excludePath) {
			return true
		}
	}

	for _, pattern := range s.cfg.Diff.ExcludePatterns {
		matched, err := regexp.MatchString(pattern, filePath)
		if err == nil && matched {
			return true
		}
	}

	return false
}

func (s *GerritSyncService) syncComments(ctx context.Context, changeDBID int, changeID string) error {
	comments, err := s.gerritClient.ListChangeComments(ctx, changeID)
	if err != nil {
		return fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	for filePath, fileComments := range comments {
		for _, comment := range fileComments {
			authorName := ""
			authorEmail := ""
			if comment.Author != nil {
				authorName = comment.Author.Name
				authorEmail = comment.Author.Email
			}

			created, _ := parseGerritTime(comment.Updated)
			updated := created

			var line interface{}
			if comment.Line > 0 {
				line = comment.Line
			}

			var inReplyTo interface{}
			if comment.InReplyTo != "" {
				inReplyTo = comment.InReplyTo
			}

			var filePathVal interface{}
			if filePath != "/PATCHSET_LEVEL" {
				filePathVal = filePath
			}

			_, err := s.db.ExecContext(ctx, `
				INSERT INTO comments (change_db_id, comment_id, file_path, line, author_name, author_email, message, created, updated, in_reply_to, unresolved)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(change_db_id, comment_id) DO UPDATE SET
					message = excluded.message,
					updated = excluded.updated,
					unresolved = excluded.unresolved
			`, changeDBID, comment.ID, filePathVal, line, authorName, authorEmail, comment.Message, created, updated, inReplyTo, boolToInt(comment.Unresolved))

			if err != nil {
				s.logger.Warn("コメントの保存に失敗しました",
					zap.String("comment_id", comment.ID),
					zap.Error(err))
			}
		}
	}

	return nil
}

func (s *GerritSyncService) syncMessages(ctx context.Context, changeDBID int, change *grt.ChangeInfo) error {
	for _, msg := range change.Messages {
		authorName := ""
		authorEmail := ""
		if msg.Author != nil {
			authorName = msg.Author.Name
			authorEmail = msg.Author.Email
		}

		date, _ := parseGerritTime(msg.Date)

		var revisionNumber interface{}
		if msg.RevisionNumber > 0 {
			revisionNumber = msg.RevisionNumber
		}

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO messages (change_db_id, message_id, author_name, author_email, message, date, revision_number)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(change_db_id, message_id) DO UPDATE SET
				message = excluded.message
		`, changeDBID, msg.ID, authorName, authorEmail, msg.Message, date, revisionNumber)

		if err != nil {
			s.logger.Warn("メッセージの保存に失敗しました",
				zap.String("message_id", msg.ID),
				zap.Error(err))
		}
	}

	return nil
}

func (s *GerritSyncService) syncLabels(ctx context.Context, changeDBID int, change *grt.ChangeInfo) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM labels WHERE change_db_id = ?", changeDBID)
	if err != nil {
		s.logger.Warn("既存ラベルの削除に失敗しました", zap.Error(err))
	}

	for labelName, labelInfo := range change.Labels {
		if labelInfo.All == nil {
			continue
		}

		for _, approval := range labelInfo.All {
			if approval.Value == 0 {
				continue
			}

			accountName := ""
			accountEmail := ""
			if approval.AccountInfo != nil {
				accountName = approval.AccountInfo.Name
				accountEmail = approval.AccountInfo.Email
			}

			grantedOn, _ := parseGerritTime(approval.Date)
			if grantedOn.IsZero() {
				grantedOn = time.Now()
			}

			_, err := s.db.ExecContext(ctx, `
				INSERT INTO labels (change_db_id, label_name, value, account_name, account_email, granted_on)
				VALUES (?, ?, ?, ?, ?, ?)
			`, changeDBID, labelName, approval.Value, accountName, accountEmail, grantedOn)

			if err != nil {
				s.logger.Warn("ラベルの保存に失敗しました",
					zap.String("label_name", labelName),
					zap.Error(err))
			}
		}
	}

	return nil
}

func (s *GerritSyncService) CheckUpdates(ctx context.Context) (bool, error) {
	sinceTime := time.Now().AddDate(0, 0, -1)

	opts := gerrit.QueryOptions{
		Project:  s.cfg.Gerrit.Project,
		Branches: s.cfg.Gerrit.Branches,
		Status:   s.cfg.Gerrit.Status,
		After:    sinceTime,
		Limit:    1,
	}

	changes, err := s.gerritClient.QueryChanges(ctx, opts)
	if err != nil {
		return false, fmt.Errorf("更新チェックに失敗しました: %w", err)
	}

	return len(changes) > 0, nil
}

func parseGerritTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, fmt.Errorf("空の時刻文字列")
	}

	formats := []string{
		"2006-01-02 15:04:05.000000000",
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("時刻のパースに失敗しました: %s", timeStr)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
