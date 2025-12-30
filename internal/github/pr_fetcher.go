package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"
)

type PRFetcher struct {
	client *Client
	logger *zap.Logger
}

func NewPRFetcher(client *Client, logger *zap.Logger) *PRFetcher {
	return &PRFetcher{
		client: client,
		logger: logger,
	}
}

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

func (f *PRFetcher) FetchPRComments(ctx context.Context, owner, repo string, number int) ([]*github.IssueComment, error) {
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

	return allComments, nil
}

func (f *PRFetcher) FetchPRDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return "", err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return "", err
	}

	// PRのdiffを取得
	diff, _, err := f.client.GetClient().PullRequests.GetRaw(ctx, owner, repo, number, github.RawOptions{
		Type: github.Diff,
	})
	if err != nil {
		return "", fmt.Errorf("PR差分の取得に失敗しました: %w", err)
	}

	return diff, nil
}

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
