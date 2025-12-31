package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"upgo/internal/gerrit"
	"upgo/internal/tracker"

	"go.uber.org/zap"
)

// GerritSyncService coordinates the synchronization of Gerrit data (Changes, revisions, comments)
// into the local database. It handles fetching, storing, and managing Change data.
type GerritSyncService struct {
	db              *sql.DB
	gerritClient    *gerrit.Client
	changeFetcher   *gerrit.ChangeFetcher
	diffPolicy      *gerrit.DiffPolicy
	statusTracker   *tracker.StatusTracker
	logger          *zap.Logger
	project         string
	safetyWindow    time.Duration // Safety window for re-fetching (default: 10 minutes)
}

// NewGerritSyncService creates a new Gerrit sync service
func NewGerritSyncService(
	db *sql.DB,
	gerritClient *gerrit.Client,
	changeFetcher *gerrit.ChangeFetcher,
	diffPolicy *gerrit.DiffPolicy,
	statusTracker *tracker.StatusTracker,
	logger *zap.Logger,
	project string,
) *GerritSyncService {
	return &GerritSyncService{
		db:            db,
		gerritClient:  gerritClient,
		changeFetcher: changeFetcher,
		diffPolicy:    diffPolicy,
		statusTracker: statusTracker,
		logger:        logger,
		project:       project,
		safetyWindow:  10 * time.Minute,
	}
}

// Sync performs a full synchronization of Gerrit data for the configured project.
// It fetches and stores Changes (including their revisions, files, diffs, and comments),
// then updates the last_synced_at timestamp.
func (s *GerritSyncService) Sync(ctx context.Context) error {
	return s.SyncWithOptions(ctx, false)
}

// SyncWithOptions performs synchronization with options (e.g., force full sync).
func (s *GerritSyncService) SyncWithOptions(ctx context.Context, forceFullSync bool) error {
	syncStart := time.Now()
	s.logger.Info("Gerrit同期を開始しました", zap.Bool("force_full_sync", forceFullSync))

	repoID, err := s.getOrCreateRepository()
	if err != nil {
		return fmt.Errorf("リポジトリの取得に失敗しました: %w", err)
	}

	if err := s.syncChangesWithOptions(ctx, repoID, forceFullSync); err != nil {
		s.logger.Error("Change同期に失敗しました", zap.Error(err))
		return err
	}

	_, err = s.db.Exec(
		"UPDATE repositories SET last_synced_at = ? WHERE id = ?",
		time.Now(), repoID,
	)
	if err != nil {
		s.logger.Warn("最終同期時刻の更新に失敗しました", zap.Error(err))
	}

	totalDuration := time.Since(syncStart)
	s.logger.Info("Gerrit同期が完了しました",
		zap.Duration("total_duration", totalDuration),
	)
	return nil
}

// SyncChangeByID synchronizes a single Change by its database ID
func (s *GerritSyncService) SyncChangeByID(ctx context.Context, changeID int) error {
	s.logger.Info("Change同期を開始しました", zap.Int("change_id", changeID))

	// Get Change information from database
	var changeNumber int
	var repoID int
	var changeGerritID string
	err := s.db.QueryRow(
		"SELECT change_number, repository_id, change_id FROM changes WHERE id = ?",
		changeID,
	).Scan(&changeNumber, &repoID, &changeGerritID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("Changeが見つかりません: change_id=%d", changeID)
	}
	if err != nil {
		return fmt.Errorf("Change情報の取得に失敗しました: %w", err)
	}

	// Fetch Change from Gerrit
	changeInfo, err := s.changeFetcher.FetchChangeDetail(ctx, changeGerritID)
	if err != nil {
		return fmt.Errorf("GerritからのChange取得に失敗しました: %w", err)
	}

	// Save Change using existing saveChange method
	if _, _, _, err = s.saveChange(ctx, repoID, changeInfo); err != nil {
		return fmt.Errorf("Changeの保存に失敗しました: %w", err)
	}

	s.logger.Info("Change同期が完了しました", zap.Int("change_id", changeID), zap.Int("change_number", changeNumber))
	return nil
}

// getOrCreateRepository retrieves the repository ID from the database, creating
// a new record if it doesn't exist. For Gerrit, we use project name as both owner and name.
func (s *GerritSyncService) getOrCreateRepository() (int, error) {
	var id int
	err := s.db.QueryRow(
		"SELECT id FROM repositories WHERE owner = ? AND name = ?",
		"go", s.project, // owner="go", name="go"
	).Scan(&id)

	if err == sql.ErrNoRows {
		result, err := s.db.Exec(
			"INSERT INTO repositories (owner, name, last_synced_at) VALUES (?, ?, ?)",
			"go", s.project, nil, // NULL for first sync to allow 30-day initial sync
		)
		if err != nil {
			return 0, err
		}
		repoID, _ := result.LastInsertId()
		return int(repoID), nil
	}

	if err != nil {
		return 0, err
	}

	return id, nil
}

// syncChanges fetches and saves Changes updated since the last sync.
// Uses repositories.last_synced_at as the threshold, falling back to configured days ago
// for the first sync. Adds a safety margin to avoid missing Changes that
// were updated just before the last sync completed.
func (s *GerritSyncService) syncChanges(ctx context.Context, repoID int) error {
	return s.syncChangesWithOptions(ctx, repoID, false)
}

// syncChangesWithOptions fetches and saves Changes with options (e.g., force full sync).
func (s *GerritSyncService) syncChangesWithOptions(ctx context.Context, repoID int, forceFullSync bool) error {
	var sinceTime time.Time

	if forceFullSync {
		// Force full sync: use configured days ago regardless of last_synced_at
		days := s.changeFetcher.Days()
		sinceTime = time.Now().AddDate(0, 0, -days)
		s.logger.Info("強制フル同期のため、設定日数前を起点に同期します",
			zap.Time("since_time", sinceTime),
			zap.Int("days", days),
		)
	} else {
		// Get last sync time from database
		var lastSyncedAt sql.NullTime
		err := s.db.QueryRow(
			"SELECT last_synced_at FROM repositories WHERE id = ?",
			repoID,
		).Scan(&lastSyncedAt)

		if err == nil && lastSyncedAt.Valid {
			// Use last sync time minus safety window
			sinceTime = lastSyncedAt.Time.Add(-s.safetyWindow)
			s.logger.Info("前回同期時刻を起点に同期します",
				zap.Time("last_synced_at", lastSyncedAt.Time),
				zap.Time("since_time", sinceTime),
			)
		} else {
			// First sync: use configured days ago as default
			// Get days from changeFetcher (already configured)
			days := s.changeFetcher.Days()
			sinceTime = time.Now().AddDate(0, 0, -days)
			s.logger.Info("初回同期のため、設定日数前を起点に同期します",
				zap.Time("since_time", sinceTime),
				zap.Int("days", days),
			)
		}
	}

	var totalChanges int
	var totalRevisions int
	var totalComments int
	var totalDiffs int
	changesFetchStart := time.Now()

	// Fetch Changes updated since the threshold
	changes, err := s.changeFetcher.FetchChangesUpdatedSince(ctx, sinceTime)
	if err != nil {
		return fmt.Errorf("Change一覧取得失敗: %w", err)
	}

	changesFetchDuration := time.Since(changesFetchStart)
	s.logger.Info("Change一覧取得完了",
		zap.Int("count", len(changes)),
		zap.Duration("duration", changesFetchDuration),
	)

	// Process Changes sequentially to avoid SQLite write conflicts
	for i := range changes {
		changeInfo := &changes[i]
		changeSaveStart := time.Now()
		revisionsCount, commentsCount, diffsCount, err := s.saveChange(ctx, repoID, changeInfo)
		if err != nil {
			s.logger.Error("Changeの保存に失敗しました", zap.Int("change_number", changeInfo.Number), zap.Error(err))
			continue
		}
		changeSaveDuration := time.Since(changeSaveStart)
		totalChanges++
		totalRevisions += revisionsCount
		totalComments += commentsCount
		totalDiffs += diffsCount
		s.logger.Debug("Change保存完了",
			zap.Int("change_number", changeInfo.Number),
			zap.Int("revisions_count", revisionsCount),
			zap.Int("comments_count", commentsCount),
			zap.Int("diffs_count", diffsCount),
			zap.Duration("duration", changeSaveDuration),
		)
	}

	changesSyncDuration := time.Since(changesFetchStart)
	s.logger.Info("Change同期完了",
		zap.Int("total_changes", totalChanges),
		zap.Int("total_revisions", totalRevisions),
		zap.Int("total_comments", totalComments),
		zap.Int("total_diffs", totalDiffs),
		zap.Duration("total_duration", changesSyncDuration),
	)

	return nil
}

// saveChange saves or updates a Change in the database. It handles both new Changes and updates
// to existing ones. After saving, it synchronizes revisions, files, diffs, and comments.
//
// Returns the number of revisions, comments, and diffs processed.
func (s *GerritSyncService) saveChange(ctx context.Context, repoID int, changeInfo *gerrit.ChangeInfo) (revisionsCount int, commentsCount int, diffsCount int, err error) {
	// Parse timestamps
	createdAt, err := parseGerritTime(changeInfo.Created)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("created_atの解析に失敗しました: %w", err)
	}

	updatedAt, err := parseGerritTime(changeInfo.Updated)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("updated_atの解析に失敗しました: %w", err)
	}

	var submittedAt *time.Time
	if changeInfo.Submitted != "" {
		submitted, err := parseGerritTime(changeInfo.Submitted)
		if err == nil {
			submittedAt = &submitted
		}
	}

	// Get existing Change info to check if it has changed
	var changeDBID int
	var dbUpdatedAt time.Time
	var dbStatus string
	err = s.db.QueryRow(
		"SELECT id, updated_at, status FROM changes WHERE repository_id = ? AND change_number = ?",
		repoID, changeInfo.Number,
	).Scan(&changeDBID, &dbUpdatedAt, &dbStatus)

	var isNewChange bool
	if err != nil {
		if err == sql.ErrNoRows {
			isNewChange = true
			err = nil
		} else {
			return 0, 0, 0, fmt.Errorf("既存Change情報の取得に失敗しました: %w", err)
		}
	}

	// Build Gerrit URL
	gerritURL := fmt.Sprintf("%s/c/%s/+/%d", s.gerritClient.BaseURL(), changeInfo.Project, changeInfo.Number)

	// Use UPSERT: Try UPDATE first, then INSERT if no rows affected
	if !isNewChange {
		// Update existing Change
		result, err := s.db.Exec(`
			UPDATE changes 
			SET subject = ?, message = ?, status = ?, updated_at = ?, submitted_at = ?, last_synced_at = ?
			WHERE repository_id = ? AND change_number = ?`,
			changeInfo.Subject, changeInfo.Message(), changeInfo.Status, updatedAt,
			submittedAt, time.Now(),
			repoID, changeInfo.Number,
		)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("Changeの更新に失敗しました: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return 0, 0, 0, fmt.Errorf("Changeの更新が影響を与えませんでした")
		}
		changeDBID = int(changeDBID)
	} else {
		// Insert new Change
		result, err := s.db.Exec(`
			INSERT INTO changes 
			(repository_id, change_number, change_id, project, branch, subject, message, status, owner, created_at, updated_at, submitted_at, url, last_synced_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			repoID, changeInfo.Number, changeInfo.ChangeID, changeInfo.Project, changeInfo.Branch,
			changeInfo.Subject, changeInfo.Message(), changeInfo.Status,
			changeInfo.Owner.Name, createdAt, updatedAt, submittedAt, gerritURL, time.Now(),
		)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("Changeの挿入に失敗しました: %w", err)
		}
		changeDBID64, _ := result.LastInsertId()
		changeDBID = int(changeDBID64)
	}

	// Check if Change has actually changed
	changeChanged := isNewChange || !updatedAt.Equal(dbUpdatedAt)

	// Track state changes for existing Changes
	if !isNewChange {
		_, err = s.statusTracker.TrackChangeState(changeDBID, changeInfo.Status)
		if err != nil {
			s.logger.Warn("Change状態の追跡に失敗しました", zap.Error(err))
		}
	}

	// Only sync revisions, files, diffs, and comments if Change has changed
	if !changeChanged {
		s.logger.Debug("Changeに変更がないため、詳細情報の取得をスキップします",
			zap.Int("change_id", changeDBID),
			zap.Int("change_number", changeInfo.Number),
		)
		return 0, 0, 0, nil
	}

	// Sync revisions, files, and diffs
	revisionsCount, err = s.syncRevisions(ctx, changeDBID, changeInfo)
	if err != nil {
		s.logger.Warn("Revision同期に失敗しました", zap.Error(err))
	}

	// Sync comments
	commentsCount, err = s.syncChangeComments(ctx, changeDBID, changeInfo.ChangeID)
	if err != nil {
		s.logger.Warn("Changeコメントの同期に失敗しました", zap.Error(err))
	}

	// Sync labels
	err = s.syncChangeLabels(ctx, changeDBID, changeInfo)
	if err != nil {
		s.logger.Warn("Changeラベルの同期に失敗しました", zap.Error(err))
	}

	// Sync messages
	err = s.syncChangeMessages(ctx, changeDBID, changeInfo)
	if err != nil {
		s.logger.Warn("Changeメッセージの同期に失敗しました", zap.Error(err))
	}

	return revisionsCount, commentsCount, 0, nil // diffsCount is handled in syncRevisions
}

// syncRevisions synchronizes revisions (patchsets) for a change
func (s *GerritSyncService) syncRevisions(ctx context.Context, changeDBID int, changeInfo *gerrit.ChangeInfo) (int, error) {
	if changeInfo.Revisions == nil || len(changeInfo.Revisions) == 0 {
		return 0, nil
	}

	var revisionsCount int
	for revisionSHA, revInfo := range changeInfo.Revisions {
		// Parse created timestamp
		createdAt, err := parseGerritTime(revInfo.Created)
		if err != nil {
			s.logger.Warn("Revision作成時刻の解析に失敗しました", zap.String("revision_sha", revisionSHA), zap.Error(err))
			continue
		}

		// Check if revision already exists
		var revDBID int
		err = s.db.QueryRow(
			"SELECT id FROM revisions WHERE change_id = ? AND patch_set_number = ?",
			changeDBID, revInfo.Number,
		).Scan(&revDBID)

		var isNewRevision bool
		if err == sql.ErrNoRows {
			isNewRevision = true
		} else if err != nil {
			return revisionsCount, fmt.Errorf("Revision情報の取得に失敗しました: %w", err)
		}

		commitMessage := ""
		authorName := ""
		authorEmail := ""
		committerName := ""
		committerEmail := ""
		if revInfo.Commit.Subject != "" {
			commitMessage = revInfo.Commit.Message
			authorName = revInfo.Commit.Author.Name
			authorEmail = revInfo.Commit.Author.Email
			committerName = revInfo.Commit.Committer.Name
			committerEmail = revInfo.Commit.Committer.Email
		}

		if isNewRevision {
			// Insert new revision
			result, err := s.db.Exec(`
				INSERT INTO revisions 
				(change_id, patch_set_number, revision_sha, uploader, created_at, kind, commit_message, author_name, author_email, committer_name, committer_email)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				changeDBID, revInfo.Number, revisionSHA, revInfo.Uploader.Name, createdAt,
				revInfo.Kind, commitMessage, authorName, authorEmail, committerName, committerEmail,
			)
			if err != nil {
				return revisionsCount, fmt.Errorf("Revisionの挿入に失敗しました: %w", err)
			}
			revDBID64, _ := result.LastInsertId()
			revDBID = int(revDBID64)
		} else {
			// Update existing revision
			_, err = s.db.Exec(`
				UPDATE revisions 
				SET revision_sha = ?, uploader = ?, kind = ?, commit_message = ?, author_name = ?, author_email = ?, committer_name = ?, committer_email = ?
				WHERE id = ?`,
				revisionSHA, revInfo.Uploader.Name, revInfo.Kind, commitMessage,
				authorName, authorEmail, committerName, committerEmail, revDBID,
			)
			if err != nil {
				return revisionsCount, fmt.Errorf("Revisionの更新に失敗しました: %w", err)
			}
		}

		revisionsCount++

		// Sync files and diffs for this revision
		if revInfo.Files != nil {
			if err := s.syncRevisionFiles(ctx, revDBID, changeInfo.ChangeID, revisionSHA, revInfo.Files); err != nil {
				s.logger.Warn("Revisionファイルの同期に失敗しました", zap.Error(err))
			}
		}
	}

	return revisionsCount, nil
}

// syncRevisionFiles synchronizes files and diffs for a revision
func (s *GerritSyncService) syncRevisionFiles(ctx context.Context, revisionDBID int, changeID, revisionSHA string, files map[string]gerrit.FileInfo) error {
	for filePath, fileInfo := range files {
		// Skip COMMIT_MSG and other magic files for diff storage
		if filePath == "/COMMIT_MSG" || filePath == "/MERGE_LIST" {
			continue
		}

		// Check if file already exists
		var fileDBID int
		err := s.db.QueryRow(
			"SELECT id FROM change_files WHERE revision_id = ? AND file_path = ?",
			revisionDBID, filePath,
		).Scan(&fileDBID)

		var isNewFile bool
		if err == sql.ErrNoRows {
			isNewFile = true
		} else if err != nil {
			return fmt.Errorf("ファイル情報の取得に失敗しました: %w", err)
		}

		// Get file stats
		fileStats := s.diffPolicy.GetFileDiffStats(&fileInfo)
		fileStats.FilePath = filePath

		if isNewFile {
			// Insert new file
			_, err = s.db.Exec(`
				INSERT INTO change_files 
				(revision_id, file_path, status, old_path, lines_inserted, lines_deleted, size_delta, size, binary)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				revisionDBID, fileStats.FilePath, fileStats.Status, fileStats.OldPath,
				fileStats.LinesInserted, fileStats.LinesDeleted, fileStats.SizeDelta, fileStats.Size,
				fileStats.Binary,
			)
			if err != nil {
				return fmt.Errorf("ファイルの挿入に失敗しました: %w", err)
			}
		} else {
			// Update existing file
			_, err = s.db.Exec(`
				UPDATE change_files 
				SET status = ?, old_path = ?, lines_inserted = ?, lines_deleted = ?, size_delta = ?, size = ?, binary = ?
				WHERE id = ?`,
				fileStats.Status, fileStats.OldPath, fileStats.LinesInserted, fileStats.LinesDeleted,
				fileStats.SizeDelta, fileStats.Size, fileStats.Binary, fileDBID,
			)
			if err != nil {
				return fmt.Errorf("ファイルの更新に失敗しました: %w", err)
			}
		}

		// Fetch and store diff if not binary and within size limit
		if !fileInfo.Binary {
			diff, err := s.changeFetcher.FetchFileDiff(ctx, changeID, revisionSHA, filePath)
			if err != nil {
				s.logger.Debug("ファイル差分の取得に失敗しました（スキップ）", zap.String("file_path", filePath), zap.Error(err))
				continue
			}

			shouldStore, diffText, statsOnly := s.diffPolicy.ProcessDiff(diff, filePath)
			if shouldStore && !statsOnly {
				// Store diff
				diffSize := len([]byte(diffText))
				_, err = s.db.Exec(`
					INSERT OR REPLACE INTO change_diffs 
					(revision_id, file_path, diff_text, diff_size, created_at)
					VALUES (?, ?, ?, ?, ?)`,
					revisionDBID, filePath, diffText, diffSize, time.Now(),
				)
				if err != nil {
					s.logger.Warn("差分の保存に失敗しました", zap.String("file_path", filePath), zap.Error(err))
				}
			} else {
				s.logger.Debug("差分サイズが上限を超えているため、統計のみ保存します", zap.String("file_path", filePath))
			}
		}
	}

	return nil
}

// syncChangeComments synchronizes comments for a change
func (s *GerritSyncService) syncChangeComments(ctx context.Context, changeDBID int, changeID string) (int, error) {
	// Fetch all comments for the change
	commentsMap, err := s.changeFetcher.FetchChangeComments(ctx, changeID)
	if err != nil {
		return 0, err
	}

	var totalComments int
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO change_comments 
		(change_id, revision_id, comment_id, file_path, line, patch_set_number, message, author, created_at, updated_at, in_reply_to, unresolved)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("ステートメントの準備に失敗しました: %w", err)
	}
	defer stmt.Close()

	// Get revision mapping (revision SHA -> revision DB ID)
	revisionMap := make(map[string]int)
	rows, err := tx.Query("SELECT id, revision_sha FROM revisions WHERE change_id = ?", changeDBID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var revID int
			var revSHA string
			if err := rows.Scan(&revID, &revSHA); err == nil {
				revisionMap[revSHA] = revID
			}
		}
	}

	for filePath, comments := range commentsMap {
		for _, comment := range comments {
			createdAt, err := parseGerritTime(comment.Updated) // Use updated as created fallback
			if err != nil {
				createdAt = time.Now()
			}
			updatedAt, err := parseGerritTime(comment.Updated)
			if err != nil {
				updatedAt = time.Now()
			}

			// Determine revision_id from patch_set_number
			var revisionDBID *int
			if comment.PatchSet > 0 {
				var revID int
				err = tx.QueryRow(
					"SELECT id FROM revisions WHERE change_id = ? AND patch_set_number = ?",
					changeDBID, comment.PatchSet,
				).Scan(&revID)
				if err == nil {
					revisionDBID = &revID
				}
			}

			var linePtr *int
			if comment.Line > 0 {
				linePtr = &comment.Line
			}

			_, err = stmt.Exec(
				changeDBID, revisionDBID, comment.ID, filePath, linePtr, comment.PatchSet,
				comment.Message, comment.Author.Name, createdAt, updatedAt,
				comment.InReplyTo, comment.Unresolved,
			)
			if err != nil {
				return totalComments, fmt.Errorf("コメントの保存に失敗しました: %w", err)
			}
			totalComments++
		}
	}

	if err := tx.Commit(); err != nil {
		return totalComments, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return totalComments, nil
}

// syncChangeLabels synchronizes labels for a change
func (s *GerritSyncService) syncChangeLabels(ctx context.Context, changeDBID int, changeInfo *gerrit.ChangeInfo) error {
	if changeInfo.Labels == nil {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	// Delete existing labels for this change
	_, err = tx.Exec("DELETE FROM change_labels WHERE change_id = ?", changeDBID)
	if err != nil {
		return fmt.Errorf("既存ラベルの削除に失敗しました: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO change_labels 
		(change_id, label_name, account, value, date)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("ステートメントの準備に失敗しました: %w", err)
	}
	defer stmt.Close()

	for labelName, labelInfo := range changeInfo.Labels {
		if labelInfo.All != nil {
			for _, approval := range labelInfo.All {
				date, err := parseGerritTime(approval.Date)
				if err != nil {
					date = time.Now()
				}

				accountName := approval.Name
				if accountName == "" {
					accountName = approval.Username
				}

				_, err = stmt.Exec(
					changeDBID, labelName, accountName, approval.Value, date,
				)
				if err != nil {
					return fmt.Errorf("ラベルの保存に失敗しました: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}

// syncChangeMessages synchronizes messages for a change
func (s *GerritSyncService) syncChangeMessages(ctx context.Context, changeDBID int, changeInfo *gerrit.ChangeInfo) error {
	if changeInfo.Messages == nil || len(changeInfo.Messages) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO change_messages 
		(change_id, message_id, author, message, date, revision_number)
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("ステートメントの準備に失敗しました: %w", err)
	}
	defer stmt.Close()

	for _, msg := range changeInfo.Messages {
		date, err := parseGerritTime(msg.Date)
		if err != nil {
			date = time.Now()
		}

		var revisionNumber *int
		if msg.RevisionNumber > 0 {
			revNum := msg.RevisionNumber
			revisionNumber = &revNum
		}

		authorName := msg.Author.Name
		if authorName == "" {
			authorName = msg.Author.Username
		}

		_, err = stmt.Exec(
			changeDBID, msg.ID, authorName, msg.Message, date, revisionNumber,
		)
		if err != nil {
			return fmt.Errorf("メッセージの保存に失敗しました: %w", err)
		}
	}

	return tx.Commit()
}

// parseGerritTime parses Gerrit timestamp format
func parseGerritTime(timeStr string) (time.Time, error) {
	// Gerrit uses format: "2024-01-01 12:00:00.000000000"
	// Try multiple formats
	formats := []string{
		"2006-01-02 15:04:05.000000000",
		"2006-01-02 15:04:05",
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("時刻の解析に失敗しました: %s", timeStr)
}
