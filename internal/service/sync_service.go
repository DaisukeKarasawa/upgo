package service

import (
	"context"
	"database/sql"
	"fmt"
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
	prFetcher       *github.PRFetcher
	statusTracker   *tracker.StatusTracker
	analysisService *AnalysisService
	logger          *zap.Logger
	owner           string
	repo            string
}

func NewSyncService(
	db *sql.DB,
	githubClient *github.Client,
	prFetcher *github.PRFetcher,
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
	s.logger.Info("同期を開始しました")

	repoID, err := s.getOrCreateRepository()
	if err != nil {
		return fmt.Errorf("リポジトリの取得に失敗しました: %w", err)
	}

	if err := s.syncPRs(ctx, repoID); err != nil {
		s.logger.Error("PR同期に失敗しました", zap.Error(err))
	}

	_, err = s.db.Exec(
		"UPDATE repositories SET last_synced_at = ? WHERE id = ?",
		time.Now(), repoID,
	)
	if err != nil {
		s.logger.Warn("最終同期時刻の更新に失敗しました", zap.Error(err))
	}

	s.logger.Info("同期が完了しました")
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

// syncPRs fetches and saves all PRs for both open and closed states.
// We sync both states separately because GitHub's API requires filtering by state,
// and we want to capture the complete history including closed PRs.
func (s *SyncService) syncPRs(ctx context.Context, repoID int) error {
	states := []string{"open", "closed"}
	for _, state := range states {
		prs, err := s.prFetcher.FetchPRs(ctx, s.owner, s.repo, state)
		if err != nil {
			return err
		}

		for _, pr := range prs {
			if err := s.savePR(ctx, repoID, pr); err != nil {
				s.logger.Error("PRの保存に失敗しました", zap.Int("pr_number", pr.GetNumber()), zap.Error(err))
				continue
			}
		}
	}

	return nil
}

// savePR saves or updates a PR in the database. It handles both new PRs and updates
// to existing ones. The state is normalized to "merged" if the PR was merged,
// overriding GitHub's "closed" state to provide more semantic meaning.
//
// After saving, it synchronizes comments and diffs, and triggers analysis for
// new PRs or those that changed state (e.g., opened, closed, merged).
func (s *SyncService) savePR(ctx context.Context, repoID int, pr *ghub.PullRequest) error {
	state := pr.GetState()
	mergedAt := pr.GetMergedAt()
	// Override state to "merged" if merged_at is set, providing clearer semantics
	// than GitHub's generic "closed" state
	if !mergedAt.IsZero() {
		state = "merged"
	}

	var prID int
	err := s.db.QueryRow(
		"SELECT id FROM pull_requests WHERE repository_id = ? AND github_id = ?",
		repoID, pr.GetNumber(),
	).Scan(&prID)

	isNewPR := false
	var stateChanged bool
	if err == sql.ErrNoRows {
		isNewPR = true
		// Convert time pointers to interface{} for nullable database columns.
		// Using interface{} allows nil values when the time is zero, which
		// properly represents NULL in the database rather than a zero timestamp.
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

		result, err := s.db.Exec(`
			INSERT INTO pull_requests 
			(repository_id, github_id, title, body, state, author, created_at, updated_at, merged_at, closed_at, url, last_synced_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			repoID, pr.GetNumber(), pr.GetTitle(), pr.GetBody(), state,
			pr.GetUser().GetLogin(), createdAt, updatedAt,
			mergedAtInsert, closedAtInsert, pr.GetHTMLURL(), time.Now(),
		)
		if err != nil {
			return err
		}
		prID64, _ := result.LastInsertId()
		prID = int(prID64)
		stateChanged = false
	} else if err != nil {
		return err
	} else {
		stateChanged, err = s.statusTracker.TrackPRState(prID, state)
		if err != nil {
			s.logger.Warn("PR状態の追跡に失敗しました", zap.Error(err))
		}

		var mergedAt, closedAt interface{}
		mergedAtPtr := pr.GetMergedAt()
		if !mergedAtPtr.IsZero() {
			mergedAt = mergedAtPtr.Time
		}
		closedAtPtr := pr.GetClosedAt()
		if !closedAtPtr.IsZero() {
			closedAt = closedAtPtr.Time
		}

		updatedAt := pr.GetUpdatedAt().Time

		_, err = s.db.Exec(`
			UPDATE pull_requests 
			SET title = ?, body = ?, state = ?, updated_at = ?, merged_at = ?, closed_at = ?, last_synced_at = ?
			WHERE id = ?`,
			pr.GetTitle(), pr.GetBody(), state, updatedAt,
			mergedAt, closedAt, time.Now(), prID,
		)
		if err != nil {
			return err
		}
	}

	if err := s.syncPRComments(ctx, prID, pr.GetNumber()); err != nil {
		s.logger.Warn("PRコメントの同期に失敗しました", zap.Error(err))
	}

	if err := s.syncPRDiff(ctx, prID, pr.GetNumber()); err != nil {
		s.logger.Warn("PR差分の同期に失敗しました", zap.Error(err))
	}

	// Trigger analysis only for new PRs or when state changes, avoiding redundant
	// analysis on every sync for unchanged PRs
	if isNewPR || stateChanged {
		s.triggerAnalysis(ctx, prID, "PR")
	}

	return nil
}

// syncPRComments fetches and stores all comments for a PR.
// Uses INSERT OR REPLACE to handle updates to existing comments (e.g., edited comments)
// based on the github_id, ensuring we always have the latest version.
func (s *SyncService) syncPRComments(ctx context.Context, prID int, prNumber int) error {
	comments, err := s.prFetcher.FetchPRComments(ctx, s.owner, s.repo, prNumber)
	if err != nil {
		return err
	}

	for _, comment := range comments {
		createdAt := comment.GetCreatedAt().Time
		updatedAt := comment.GetUpdatedAt().Time

		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO pull_request_comments 
			(pr_id, github_id, body, author, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			prID, comment.GetID(), comment.GetBody(), comment.GetUser().GetLogin(),
			createdAt, updatedAt,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// syncPRDiff fetches and stores the complete diff for a PR.
// The file_path is set to "all" because we're storing the entire diff as a single
// text blob rather than per-file diffs. This simplifies storage but could be
// refactored to store per-file diffs if needed for better queryability.
func (s *SyncService) syncPRDiff(ctx context.Context, prID int, prNumber int) error {
	diff, err := s.prFetcher.FetchPRDiff(ctx, s.owner, s.repo, prNumber)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO pull_request_diffs 
		(pr_id, diff_text, file_path, created_at)
		VALUES (?, ?, ?, ?)`,
		prID, diff, "all", time.Now(),
	)

	return err
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
