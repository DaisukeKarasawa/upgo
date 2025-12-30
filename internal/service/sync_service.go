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

type SyncService struct {
	db              *sql.DB
	githubClient    *github.Client
	prFetcher       *github.PRFetcher
	issueFetcher    *github.IssueFetcher
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
	issueFetcher *github.IssueFetcher,
	statusTracker *tracker.StatusTracker,
	analysisService *AnalysisService,
	logger *zap.Logger,
	owner, repo string,
) *SyncService {
	return &SyncService{
		db:              db,
		githubClient:    githubClient,
		prFetcher:       prFetcher,
		issueFetcher:    issueFetcher,
		statusTracker:   statusTracker,
		analysisService: analysisService,
		logger:          logger,
		owner:           owner,
		repo:            repo,
	}
}

func (s *SyncService) Sync(ctx context.Context) error {
	s.logger.Info("同期を開始しました")

	// リポジトリの取得または作成
	repoID, err := s.getOrCreateRepository()
	if err != nil {
		return fmt.Errorf("リポジトリの取得に失敗しました: %w", err)
	}

	// PRの同期
	if err := s.syncPRs(ctx, repoID); err != nil {
		s.logger.Error("PR同期に失敗しました", zap.Error(err))
	}

	// Issueの同期
	if err := s.syncIssues(ctx, repoID); err != nil {
		s.logger.Error("Issue同期に失敗しました", zap.Error(err))
	}

	// 最終同期時刻の更新
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

func (s *SyncService) syncPRs(ctx context.Context, repoID int) error {
	// すべての状態のPRを取得
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

func (s *SyncService) savePR(ctx context.Context, repoID int, pr *ghub.PullRequest) error {
	// PRの状態を決定
	state := pr.GetState()
	mergedAt := pr.GetMergedAt()
	if !mergedAt.IsZero() {
		state = "merged"
	}

	// PRの取得または更新
	var prID int
	err := s.db.QueryRow(
		"SELECT id FROM pull_requests WHERE repository_id = ? AND github_id = ?",
		repoID, pr.GetNumber(),
	).Scan(&prID)

	isNewPR := false
	if err == sql.ErrNoRows {
		// 新規作成
		isNewPR = true
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
	} else if err != nil {
		return err
	} else {
		// 更新
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

	// 状態変更の追跡
	stateChanged, err := s.statusTracker.TrackPRState(prID, state)
	if err != nil {
		s.logger.Warn("PR状態の追跡に失敗しました", zap.Error(err))
	}

	// コメントの取得と保存
	if err := s.syncPRComments(ctx, prID, pr.GetNumber()); err != nil {
		s.logger.Warn("PRコメントの同期に失敗しました", zap.Error(err))
	}

	// 差分の取得と保存
	if err := s.syncPRDiff(ctx, prID, pr.GetNumber()); err != nil {
		s.logger.Warn("PR差分の同期に失敗しました", zap.Error(err))
	}

	// 要約・分析をトリガー（初回保存時または状態変更時）
	if isNewPR || stateChanged {
		s.triggerAnalysis(ctx, prID, "PR")
	}

	return nil
}

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

func (s *SyncService) syncPRDiff(ctx context.Context, prID int, prNumber int) error {
	diff, err := s.prFetcher.FetchPRDiff(ctx, s.owner, s.repo, prNumber)
	if err != nil {
		return err
	}

	// 差分をファイル単位で分割して保存（簡易実装）
	// 実際の実装では、より高度な解析が必要かもしれません
	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO pull_request_diffs 
		(pr_id, diff_text, file_path, created_at)
		VALUES (?, ?, ?, ?)`,
		prID, diff, "all", time.Now(),
	)

	return err
}

func (s *SyncService) syncIssues(ctx context.Context, repoID int) error {
	states := []string{"open", "closed"}
	for _, state := range states {
		issues, err := s.issueFetcher.FetchIssues(ctx, s.owner, s.repo, state)
		if err != nil {
			return err
		}

		for _, issue := range issues {
			if err := s.saveIssue(ctx, repoID, issue); err != nil {
				s.logger.Error("Issueの保存に失敗しました", zap.Int("issue_number", issue.GetNumber()), zap.Error(err))
				continue
			}
		}
	}

	return nil
}

func (s *SyncService) saveIssue(ctx context.Context, repoID int, issue *ghub.Issue) error {
	var issueID int
	err := s.db.QueryRow(
		"SELECT id FROM issues WHERE repository_id = ? AND github_id = ?",
		repoID, issue.GetNumber(),
	).Scan(&issueID)

	isNewIssue := false
	if err == sql.ErrNoRows {
		// 新規作成
		isNewIssue = true
		var closedAtInsert interface{}
		closedAtInsertPtr := issue.GetClosedAt()
		if !closedAtInsertPtr.IsZero() {
			closedAtInsert = closedAtInsertPtr.Time
		}

		createdAt := issue.GetCreatedAt().Time
		updatedAt := issue.GetUpdatedAt().Time

		result, err := s.db.Exec(`
			INSERT INTO issues 
			(repository_id, github_id, title, body, state, author, created_at, updated_at, closed_at, url, last_synced_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			repoID, issue.GetNumber(), issue.GetTitle(), issue.GetBody(), issue.GetState(),
			issue.GetUser().GetLogin(), createdAt, updatedAt,
			closedAtInsert, issue.GetHTMLURL(), time.Now(),
		)
		if err != nil {
			return err
		}
		issueID64, _ := result.LastInsertId()
		issueID = int(issueID64)
	} else if err != nil {
		return err
	} else {
		var closedAt interface{}
		closedAtPtr := issue.GetClosedAt()
		if !closedAtPtr.IsZero() {
			closedAt = closedAtPtr.Time
		}

		updatedAt := issue.GetUpdatedAt().Time

		_, err = s.db.Exec(`
			UPDATE issues 
			SET title = ?, body = ?, state = ?, updated_at = ?, closed_at = ?, last_synced_at = ?
			WHERE id = ?`,
			issue.GetTitle(), issue.GetBody(), issue.GetState(), updatedAt,
			closedAt, time.Now(), issueID,
		)
		if err != nil {
			return err
		}
	}

	// 状態変更の追跡
	stateChanged, err := s.statusTracker.TrackIssueState(issueID, issue.GetState())
	if err != nil {
		s.logger.Warn("Issue状態の追跡に失敗しました", zap.Error(err))
	}

	// コメントの取得と保存
	if err := s.syncIssueComments(ctx, issueID, issue.GetNumber()); err != nil {
		s.logger.Warn("Issueコメントの同期に失敗しました", zap.Error(err))
	}

	// 要約・分析をトリガー（初回保存時または状態変更時）
	if isNewIssue || stateChanged {
		s.triggerAnalysis(ctx, issueID, "Issue")
	}

	return nil
}

// triggerAnalysis は要約・分析を非同期で実行し、失敗時はリトライします
func (s *SyncService) triggerAnalysis(ctx context.Context, id int, itemType string) {
	go func() {
		analysisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		var err error
		if itemType == "PR" {
			err = s.analysisService.AnalyzePR(analysisCtx, id)
		} else {
			err = s.analysisService.AnalyzeIssue(analysisCtx, id)
		}

		if err != nil {
			s.logger.Warn("要約・分析の実行に失敗しました。リトライします",
				zap.String("type", itemType),
				zap.Int("id", id),
				zap.Error(err),
			)
			// 非同期リトライ（最大3回、指数バックオフ）
			// リトライでは新しいコンテキストを作成する
			s.retryAnalysis(id, itemType, 1)
		} else {
			s.logger.Info("要約・分析が完了しました",
				zap.String("type", itemType),
				zap.Int("id", id),
			)
		}
	}()
}

// retryAnalysis は要約・分析をリトライします
// 各リトライのたびに新しいコンテキストを作成します
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

	// 指数バックオフ: 1秒、2秒、4秒
	backoff := time.Duration(1<<uint(attempt-1)) * time.Second
	time.Sleep(backoff)

	// 各リトライのたびに新しいコンテキストを作成
	analysisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var err error
	if itemType == "PR" {
		err = s.analysisService.AnalyzePR(analysisCtx, id)
	} else {
		err = s.analysisService.AnalyzeIssue(analysisCtx, id)
	}

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

func (s *SyncService) syncIssueComments(ctx context.Context, issueID int, issueNumber int) error {
	comments, err := s.issueFetcher.FetchIssueComments(ctx, s.owner, s.repo, issueNumber)
	if err != nil {
		return err
	}

	for _, comment := range comments {
		createdAt := comment.GetCreatedAt().Time
		updatedAt := comment.GetUpdatedAt().Time

		_, err := s.db.Exec(`
			INSERT OR REPLACE INTO issue_comments 
			(issue_id, github_id, body, author, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			issueID, comment.GetID(), comment.GetBody(), comment.GetUser().GetLogin(),
			createdAt, updatedAt,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
