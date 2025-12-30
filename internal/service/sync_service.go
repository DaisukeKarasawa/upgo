package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"upgo/internal/github"
	"upgo/internal/tracker"

	ghub "github.com/google/go-github/v60/github"
	"go.uber.org/zap"
)

// SyncService coordinates the synchronization of GitHub data (PRs, comments)
// into the local database. It handles fetching, storing, and triggering analysis
// for newly created or state-changed items.
type SyncService struct {
	db              *sql.DB
	githubClient    *github.Client
	prFetcher       github.PRFetcherInterface
	statusTracker   *tracker.StatusTracker
	analysisService *AnalysisService
	logger          *zap.Logger
	owner           string
	repo            string
}

func NewSyncService(
	db *sql.DB,
	githubClient *github.Client,
	prFetcher github.PRFetcherInterface,
	statusTracker *tracker.StatusTracker,
	analysisService *AnalysisService,
	logger *zap.Logger,
	owner, repo string,
) *SyncService {
	return &SyncService{
		db:              db,
		githubClient:    githubClient,
		prFetcher:       prFetcher,
		statusTracker:   statusTracker,
		analysisService: analysisService,
		logger:          logger,
		owner:           owner,
		repo:            repo,
	}
}

// Sync performs a full synchronization of GitHub data for the configured repository.
// It fetches and stores PRs (including their comments and diffs),
// then updates the last_synced_at timestamp. Errors in individual sync operations
// are logged but don't stop the overall process, allowing partial success.
func (s *SyncService) Sync(ctx context.Context) error {
	syncStart := time.Now()
	s.logger.Info("同期を開始しました")

	repoID, err := s.getOrCreateRepository()
	if err != nil {
		return fmt.Errorf("リポジトリの取得に失敗しました: %w", err)
	}

	if err := s.syncPRs(ctx, repoID); err != nil {
		s.logger.Error("PR同期に失敗しました", zap.Error(err))
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
	s.logger.Info("同期が完了しました",
		zap.Duration("total_duration", totalDuration),
	)
	return nil
}

// SyncPRByID synchronizes a single PR by its database ID.
// It retrieves the PR's github_id and repository_id from the database,
// fetches the latest PR data from GitHub, and saves it using the existing savePR method.
// This ensures that comments, diffs, and analysis are also updated for the PR.
func (s *SyncService) SyncPRByID(ctx context.Context, prID int) error {
	s.logger.Info("PR同期を開始しました", zap.Int("pr_id", prID))

	// Get PR information from database
	var githubID, repoID int
	err := s.db.QueryRow(
		"SELECT github_id, repository_id FROM pull_requests WHERE id = ?",
		prID,
	).Scan(&githubID, &repoID)

	if err == sql.ErrNoRows {
		return fmt.Errorf("PRが見つかりません: pr_id=%d", prID)
	}
	if err != nil {
		return fmt.Errorf("PR情報の取得に失敗しました: %w", err)
	}

	// Fetch PR from GitHub
	pr, err := s.prFetcher.FetchPR(ctx, s.owner, s.repo, githubID)
	if err != nil {
		return fmt.Errorf("GitHubからのPR取得に失敗しました: %w", err)
	}

	// Save PR using existing savePR method (handles comments, diffs, and analysis)
	_, _, err = s.savePR(ctx, repoID, pr)
	if err != nil {
		return fmt.Errorf("PRの保存に失敗しました: %w", err)
	}

	s.logger.Info("PR同期が完了しました", zap.Int("pr_id", prID), zap.Int("github_id", githubID))
	return nil
}

// getOrCreateRepository retrieves the repository ID from the database, creating
// a new record if it doesn't exist. This ensures we have a valid repository ID
// for foreign key relationships before syncing PRs.
func (s *SyncService) getOrCreateRepository() (int, error) {
	var id int
	err := s.db.QueryRow(
		"SELECT id FROM repositories WHERE owner = ? AND name = ?",
		s.owner, s.repo,
	).Scan(&id)

	if err == sql.ErrNoRows {
		result, err := s.db.Exec(
			"INSERT INTO repositories (owner, name, last_synced_at) VALUES (?, ?, ?)",
			s.owner, s.repo, time.Now(),
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

// syncPRs fetches and saves PRs updated since the last sync.
// Uses repositories.last_synced_at as the threshold, falling back to 30 days ago
// for the first sync. Adds a 5-minute safety margin to avoid missing PRs that
// were updated just before the last sync completed.
//
// Important: We stop fetching older PR pages at the GitHub API layer
// (sorted by updated desc) to avoid retrieving PRs outside the sync window.
func (s *SyncService) syncPRs(ctx context.Context, repoID int) error {
	// Get last sync time from database
	var lastSyncedAt sql.NullTime
	err := s.db.QueryRow(
		"SELECT last_synced_at FROM repositories WHERE id = ?",
		repoID,
	).Scan(&lastSyncedAt)

	var sinceTime time.Time
	if err == nil && lastSyncedAt.Valid {
		// Use last sync time minus 5 minutes as safety margin
		sinceTime = lastSyncedAt.Time.Add(-5 * time.Minute)
		s.logger.Info("前回同期時刻を起点に同期します",
			zap.Time("last_synced_at", lastSyncedAt.Time),
			zap.Time("since_time", sinceTime),
		)
	} else {
		// First sync: use 30 days ago as default
		sinceTime = time.Now().AddDate(0, 0, -30)
		s.logger.Info("初回同期のため、30日前を起点に同期します",
			zap.Time("since_time", sinceTime),
		)
	}

	var totalPRs int
	var totalComments int
	var totalDiffs int
	prsFetchStart := time.Now()

	// Channel to collect PRs from parallel fetchers
	prChan := make(chan *ghub.PullRequest, 100)
	errChan := make(chan error, 2)
	var wg sync.WaitGroup

	// Fetch PRs updated since the threshold in parallel for each state
	states := []string{"open", "closed"}
	for _, state := range states {
		wg.Add(1)
		go func(state string) {
			defer wg.Done()
			stateFetchStart := time.Now()
			prs, err := s.prFetcher.FetchPRsUpdatedSince(ctx, s.owner, s.repo, state, sinceTime)
			if err != nil {
				errChan <- fmt.Errorf("PR一覧取得失敗 (%s): %w", state, err)
				return
			}
			stateFetchDuration := time.Since(stateFetchStart)
			s.logger.Info("PR一覧取得完了",
				zap.String("state", state),
				zap.Int("count", len(prs)),
				zap.Duration("duration", stateFetchDuration),
			)

			for _, pr := range prs {
				prChan <- pr
			}
		}(state)
	}

	// Close channel when all fetchers are done
	go func() {
		wg.Wait()
		close(prChan)
	}()

	// Process PRs sequentially to avoid SQLite write conflicts
	// This allows parallel fetching while maintaining safe sequential writes
	done := make(chan bool)
	go func() {
		defer close(done)
		for pr := range prChan {
			prSaveStart := time.Now()
			commentsCount, diffsCount, err := s.savePR(ctx, repoID, pr)
			if err != nil {
				s.logger.Error("PRの保存に失敗しました", zap.Int("pr_number", pr.GetNumber()), zap.Error(err))
				continue
			}
			prSaveDuration := time.Since(prSaveStart)
			totalPRs++
			totalComments += commentsCount
			totalDiffs += diffsCount
			s.logger.Debug("PR保存完了",
				zap.Int("pr_number", pr.GetNumber()),
				zap.Int("comments_count", commentsCount),
				zap.Int("diffs_count", diffsCount),
				zap.Duration("duration", prSaveDuration),
			)
		}
	}()

	// Wait for completion or error
	select {
	case err := <-errChan:
		// Wait for processing to finish before returning error
		<-done
		return err
	case <-done:
		// All processing completed successfully
	}

	prsFetchDuration := time.Since(prsFetchStart)
	s.logger.Info("PR同期完了",
		zap.Int("total_prs", totalPRs),
		zap.Int("total_comments", totalComments),
		zap.Int("total_diffs", totalDiffs),
		zap.Duration("total_duration", prsFetchDuration),
	)

	return nil
}

// savePR saves or updates a PR in the database. It handles both new PRs and updates
// to existing ones. The state is normalized to "merged" if the PR was merged,
// overriding GitHub's "closed" state to provide more semantic meaning.
//
// After saving, it synchronizes comments and diffs, and triggers analysis for
// new PRs or those that changed state (e.g., opened, closed, merged).
//
// Returns the number of comments and diffs processed.
func (s *SyncService) savePR(ctx context.Context, repoID int, pr *ghub.PullRequest) (commentsCount int, diffsCount int, err error) {
	state := pr.GetState()
	mergedAt := pr.GetMergedAt()
	// Override state to "merged" if merged_at is set, providing clearer semantics
	// than GitHub's generic "closed" state
	if !mergedAt.IsZero() {
		state = "merged"
	}

	// Get existing PR info to check if it has changed
	var prID int
	var dbUpdatedAt time.Time
	var dbHeadSha sql.NullString
	var dbState string
	err = s.db.QueryRow(
		"SELECT id, updated_at, head_sha, state FROM pull_requests WHERE repository_id = ? AND github_id = ?",
		repoID, pr.GetNumber(),
	).Scan(&prID, &dbUpdatedAt, &dbHeadSha, &dbState)

	// Handle database errors properly
	var isNewPR bool
	if err != nil {
		if err == sql.ErrNoRows {
			// PR doesn't exist yet, this is expected for new PRs
			isNewPR = true
			err = nil // Reset error since this is a valid case
		} else {
			// Other database errors (connection issues, SQL errors, etc.)
			return 0, 0, fmt.Errorf("既存PR情報の取得に失敗しました: %w", err)
		}
	}
	prUpdatedAt := pr.GetUpdatedAt().Time
	prHeadSha := pr.GetHead().GetSHA()

	// Convert time pointers to interface{} for nullable database columns.
	var mergedAtInsert, closedAtInsert interface{}
	mergedAtInsertPtr := pr.GetMergedAt()
	if !mergedAtInsertPtr.IsZero() {
		mergedAtInsert = mergedAtInsertPtr.Time
	}
	closedAtInsertPtr := pr.GetClosedAt()
	if !closedAtInsertPtr.IsZero() {
		closedAtInsert = closedAtInsertPtr.Time
	}

	createdAt := pr.GetCreatedAt().Time
	updatedAt := pr.GetUpdatedAt().Time

	// Use UPSERT: Try UPDATE first, then INSERT if no rows affected
	// This preserves the existing ID and avoids AUTOINCREMENT issues
	if !isNewPR {
		// Update existing PR
		result, err := s.db.Exec(`
			UPDATE pull_requests 
			SET title = ?, body = ?, state = ?, updated_at = ?, merged_at = ?, closed_at = ?, last_synced_at = ?, head_sha = ?
			WHERE repository_id = ? AND github_id = ?`,
			pr.GetTitle(), pr.GetBody(), state, updatedAt,
			mergedAtInsert, closedAtInsert, time.Now(), prHeadSha,
			repoID, pr.GetNumber(),
		)
		if err != nil {
			return 0, 0, fmt.Errorf("PRの更新に失敗しました: %w", err)
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			// Should not happen, but handle gracefully
			return 0, 0, fmt.Errorf("PRの更新が影響を与えませんでした")
		}
	} else {
		// Insert new PR
		result, err := s.db.Exec(`
			INSERT INTO pull_requests 
			(repository_id, github_id, title, body, state, author, created_at, updated_at, merged_at, closed_at, url, last_synced_at, head_sha)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			repoID, pr.GetNumber(), pr.GetTitle(), pr.GetBody(), state,
			pr.GetUser().GetLogin(), createdAt, updatedAt,
			mergedAtInsert, closedAtInsert, pr.GetHTMLURL(), time.Now(), prHeadSha,
		)
		if err != nil {
			return 0, 0, fmt.Errorf("PRの挿入に失敗しました: %w", err)
		}
		prID64, _ := result.LastInsertId()
		prID = int(prID64)
	}

	// Check if PR has actually changed
	prChanged := isNewPR || !prUpdatedAt.Equal(dbUpdatedAt) ||
		(dbHeadSha.Valid && dbHeadSha.String != prHeadSha) ||
		(!dbHeadSha.Valid && prHeadSha != "")

	// Track state changes for existing PRs
	var stateChanged bool
	if !isNewPR {
		stateChanged, err = s.statusTracker.TrackPRState(prID, state)
		if err != nil {
			s.logger.Warn("PR状態の追跡に失敗しました", zap.Error(err))
		}
	} else {
		stateChanged = false
	}

	// Only sync comments and diffs if PR has changed
	if !prChanged {
		s.logger.Debug("PRに変更がないため、コメントと差分の取得をスキップします",
			zap.Int("pr_id", prID),
			zap.Int("pr_number", pr.GetNumber()),
		)
		return 0, 0, nil
	}

	commentsCount, err = s.syncPRComments(ctx, prID, pr.GetNumber())
	if err != nil {
		s.logger.Warn("PRコメントの同期に失敗しました", zap.Error(err))
	}

	// Only fetch diff if head_sha changed
	headShaChanged := false
	if dbHeadSha.Valid {
		headShaChanged = dbHeadSha.String != prHeadSha
	} else {
		headShaChanged = prHeadSha != ""
	}

	if headShaChanged {
		diffsCount, err = s.syncPRDiff(ctx, prID, pr.GetNumber())
		if err != nil {
			s.logger.Warn("PR差分の同期に失敗しました", zap.Error(err))
		}
	} else {
		s.logger.Debug("head_shaに変更がないため、差分の取得をスキップします",
			zap.Int("pr_id", prID),
			zap.Int("pr_number", pr.GetNumber()),
		)
		diffsCount = 0
	}

	// Trigger analysis only for new PRs or when state changes, avoiding redundant
	// analysis on every sync for unchanged PRs
	if isNewPR || stateChanged {
		s.triggerAnalysis(ctx, prID, "PR")
	}

	return commentsCount, diffsCount, nil
}

// syncPRComments fetches and stores comments for a PR updated since the last sync.
// Uses the latest comment's updated_at as the threshold to fetch only new/updated comments.
// Uses INSERT OR REPLACE to handle updates to existing comments (e.g., edited comments)
// based on the github_id, ensuring we always have the latest version.
// Uses a transaction with prepared statement for batch insertion.
// Returns the number of comments processed.
func (s *SyncService) syncPRComments(ctx context.Context, prID int, prNumber int) (int, error) {
	// Get the latest comment's updated_at to use as 'since' parameter
	var latestCommentUpdatedAt sql.NullTime
	err := s.db.QueryRow(
		"SELECT MAX(updated_at) FROM pull_request_comments WHERE pr_id = ?",
		prID,
	).Scan(&latestCommentUpdatedAt)

	var sinceTime time.Time
	if err == nil && latestCommentUpdatedAt.Valid {
		// Use latest comment time minus 1 minute as safety margin
		sinceTime = latestCommentUpdatedAt.Time.Add(-1 * time.Minute)
	}

	commentsFetchStart := time.Now()
	comments, err := s.prFetcher.FetchPRCommentsSince(ctx, s.owner, s.repo, prNumber, sinceTime)
	if err != nil {
		return 0, err
	}
	commentsFetchDuration := time.Since(commentsFetchStart)

	if len(comments) == 0 {
		return 0, nil
	}

	commentsSaveStart := time.Now()
	// Use transaction for batch insertion
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("トランザクションの開始に失敗しました: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO pull_request_comments 
		(pr_id, github_id, body, author, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("ステートメントの準備に失敗しました: %w", err)
	}
	defer stmt.Close()

	for _, comment := range comments {
		createdAt := comment.GetCreatedAt().Time
		updatedAt := comment.GetUpdatedAt().Time

		_, err := stmt.Exec(
			prID, comment.GetID(), comment.GetBody(), comment.GetUser().GetLogin(),
			createdAt, updatedAt,
		)
		if err != nil {
			return 0, fmt.Errorf("コメントの保存に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}
	commentsSaveDuration := time.Since(commentsSaveStart)

	s.logger.Debug("PRコメント同期完了",
		zap.Int("pr_id", prID),
		zap.Int("pr_number", prNumber),
		zap.Int("count", len(comments)),
		zap.Duration("fetch_duration", commentsFetchDuration),
		zap.Duration("save_duration", commentsSaveDuration),
	)

	return len(comments), nil
}

// syncPRDiff fetches and stores the complete diff for a PR.
// The file_path is set to "all" because we're storing the entire diff as a single
// text blob rather than per-file diffs. This simplifies storage but could be
// refactored to store per-file diffs if needed for better queryability.
// Returns 1 if diff was saved, 0 if not (e.g., error or empty diff).
func (s *SyncService) syncPRDiff(ctx context.Context, prID int, prNumber int) (int, error) {
	diffFetchStart := time.Now()
	diff, err := s.prFetcher.FetchPRDiff(ctx, s.owner, s.repo, prNumber)
	if err != nil {
		return 0, err
	}
	diffFetchDuration := time.Since(diffFetchStart)

	diffSaveStart := time.Now()
	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO pull_request_diffs 
		(pr_id, diff_text, file_path, created_at)
		VALUES (?, ?, ?, ?)`,
		prID, diff, "all", time.Now(),
	)
	diffSaveDuration := time.Since(diffSaveStart)

	if err == nil {
		s.logger.Debug("PR差分同期完了",
			zap.Int("pr_id", prID),
			zap.Int("pr_number", prNumber),
			zap.Int("diff_size_bytes", len(diff)),
			zap.Duration("fetch_duration", diffFetchDuration),
			zap.Duration("save_duration", diffSaveDuration),
		)
		return 1, nil
	}

	return 0, err
}

// triggerAnalysis starts an asynchronous analysis task for a PR.
// The analysis runs in a background goroutine to avoid blocking the sync operation.
// A separate context with a 5-minute timeout is used to prevent analysis from
// hanging indefinitely. If analysis fails, it triggers an exponential backoff retry.
func (s *SyncService) triggerAnalysis(ctx context.Context, id int, itemType string) {
	go func() {
		analysisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err := s.analysisService.AnalyzePR(analysisCtx, id)

		if err != nil {
			s.logger.Warn("要約・分析の実行に失敗しました。リトライします",
				zap.String("type", itemType),
				zap.Int("id", id),
				zap.Error(err),
			)
			s.retryAnalysis(id, itemType, 1)
		} else {
			s.logger.Info("要約・分析が完了しました",
				zap.String("type", itemType),
				zap.Int("id", id),
			)
		}
	}()
}

// retryAnalysis implements exponential backoff retry logic for failed analysis operations.
// The backoff delay doubles with each attempt (1s, 2s, 4s) to avoid overwhelming
// the analysis service while giving transient failures time to resolve.
// After maxRetries (3), the operation is abandoned to prevent infinite retry loops.
func (s *SyncService) retryAnalysis(id int, itemType string, attempt int) {
	const maxRetries = 3
	if attempt > maxRetries {
		s.logger.Error("要約・分析のリトライが最大回数に達しました",
			zap.String("type", itemType),
			zap.Int("id", id),
			zap.Int("attempts", attempt),
		)
		return
	}

	// Exponential backoff: 2^(attempt-1) seconds (1s, 2s, 4s)
	backoff := time.Duration(1<<uint(attempt-1)) * time.Second
	time.Sleep(backoff)

	analysisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := s.analysisService.AnalyzePR(analysisCtx, id)

	if err != nil {
		s.logger.Warn("要約・分析のリトライに失敗しました",
			zap.String("type", itemType),
			zap.Int("id", id),
			zap.Int("attempt", attempt),
			zap.Error(err),
		)
		s.retryAnalysis(id, itemType, attempt+1)
	} else {
		s.logger.Info("要約・分析のリトライが成功しました",
			zap.String("type", itemType),
			zap.Int("id", id),
			zap.Int("attempt", attempt),
		)
	}
}
