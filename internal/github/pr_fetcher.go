package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"
)

// PRFetcher provides methods to fetch pull request data from GitHub.
// It wraps the GitHub client and handles rate limiting and pagination automatically.
type PRFetcher struct {
	client *Client
	logger *zap.Logger
}

// NewPRFetcher creates a new PRFetcher instance with the given GitHub client and logger.
// Returns nil if either client or logger is nil.
func NewPRFetcher(client *Client, logger *zap.Logger) *PRFetcher {
	if client == nil || logger == nil {
		return nil
	}
	return &PRFetcher{
		client: client,
		logger: logger,
	}
}

// FetchPRs retrieves all pull requests for a repository with the specified state.
// State can be "open", "closed", or "all". Returns all PRs across paginated results.
// Returns an error if rate limiting fails or the API call encounters an error.
func (f *PRFetcher) FetchPRs(ctx context.Context, owner, repo string, state string) ([]*github.PullRequest, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return nil, err
	}

	opts := &github.PullRequestListOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allPRs []*github.PullRequest
	for {
		prs, resp, err := f.client.GetClient().PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("PR一覧の取得に失敗しました: %w", err)
		}

		allPRs = append(allPRs, prs...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	f.logger.Info("PR一覧を取得しました", zap.Int("count", len(allPRs)))
	return allPRs, nil
}

// FetchPRsUpdatedSince retrieves pull requests updated since the given time.
//
// Implementation notes:
// - GitHub's Pull Requests List API doesn't support a "since" parameter.
// - Instead, we sort by "updated" in descending order and stop pagination once we
//   hit PRs older than the threshold. This prevents fetching older PR pages.
func (f *PRFetcher) FetchPRsUpdatedSince(ctx context.Context, owner, repo string, state string, since time.Time) ([]*github.PullRequest, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return nil, err
	}

	opts := &github.PullRequestListOptions{
		State:     state,
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var recentPRs []*github.PullRequest
	for {
		prs, resp, err := f.client.GetClient().PullRequests.List(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("PR一覧の取得に失敗しました: %w", err)
		}

		// Because results are sorted by updated desc, once updated_at is older than
		// "since", everything after is also older, and we can stop pagination.
		for _, pr := range prs {
			if pr.GetUpdatedAt().Time.Before(since) {
				f.logger.Info(
					"PR一覧を取得しました（期間で打ち切り）",
					zap.Int("count", len(recentPRs)),
					zap.Time("since", since),
				)
				return recentPRs, nil
			}
			recentPRs = append(recentPRs, pr)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	f.logger.Info(
		"PR一覧を取得しました（全ページ走査完了）",
		zap.Int("count", len(recentPRs)),
		zap.Time("since", since),
	)
	return recentPRs, nil
}

// FetchPR retrieves a single pull request by its number.
// Returns an error if rate limiting fails or the API call encounters an error.
func (f *PRFetcher) FetchPR(ctx context.Context, owner, repo string, number int) (*github.PullRequest, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return nil, err
	}

	pr, _, err := f.client.GetClient().PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("PR詳細の取得に失敗しました: %w", err)
	}

	return pr, nil
}

// FetchPRComments retrieves all comments for a pull request.
// Returns all comments across paginated results.
// Returns an error if rate limiting fails or the API call encounters an error.
func (f *PRFetcher) FetchPRComments(ctx context.Context, owner, repo string, number int) ([]*github.IssueComment, error) {
	return f.FetchPRCommentsSince(ctx, owner, repo, number, time.Time{})
}

// FetchPRCommentsSince retrieves comments for a pull request updated since the given time.
// If since is zero time, retrieves all comments (same as FetchPRComments).
// Returns all comments across paginated results.
// Returns an error if rate limiting fails or the API call encounters an error.
func (f *PRFetcher) FetchPRCommentsSince(ctx context.Context, owner, repo string, number int, since time.Time) ([]*github.IssueComment, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return nil, err
	}

	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	if !since.IsZero() {
		opts.Since = &since
	}

	var allComments []*github.IssueComment
	for {
		comments, resp, err := f.client.GetClient().Issues.ListComments(ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("PRコメントの取得に失敗しました: %w", err)
		}

		allComments = append(allComments, comments...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	f.logger.Info("PRコメントを取得しました", zap.Int("count", len(allComments)))
	return allComments, nil
}

// FetchPRDiff retrieves the unified diff for a pull request.
// Returns the diff as a string in unified diff format.
// Returns an error if rate limiting fails or the API call encounters an error.
func (f *PRFetcher) FetchPRDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return "", err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return "", err
	}

	// Get PR diff
	diff, _, err := f.client.GetClient().PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{
		Type: github.Diff,
	})
	if err != nil {
		return "", fmt.Errorf("PR差分の取得に失敗しました: %w", err)
	}

	return diff, nil
}

// FetchMergedCommits retrieves all commits to a repository since the specified time.
// Returns all commits across paginated results.
// Returns an error if rate limiting fails or the API call encounters an error.
func (f *PRFetcher) FetchMergedCommits(ctx context.Context, owner, repo string, since time.Time) ([]*github.RepositoryCommit, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return nil, err
	}

	opts := &github.CommitsListOptions{
		Since: since,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allCommits []*github.RepositoryCommit
	for {
		commits, resp, err := f.client.GetClient().Repositories.ListCommits(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("コミット一覧の取得に失敗しました: %w", err)
		}

		allCommits = append(allCommits, commits...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCommits, nil
}
