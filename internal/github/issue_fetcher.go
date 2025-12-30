package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"
)

type IssueFetcher struct {
	client *Client
	logger *zap.Logger
}

func NewIssueFetcher(client *Client, logger *zap.Logger) *IssueFetcher {
	return &IssueFetcher{
		client: client,
		logger: logger,
	}
}

func (f *IssueFetcher) FetchIssues(ctx context.Context, owner, repo string, state string) ([]*github.Issue, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return nil, err
	}

	opts := &github.IssueListByRepoOptions{
		State: state,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var allIssues []*github.Issue
	for {
		issues, resp, err := f.client.GetClient().Issues.ListByRepo(ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("Issue一覧の取得に失敗しました: %w", err)
		}

		// Filter out pull requests (they have a non-nil PullRequestLinks field)
		for _, issue := range issues {
			if issue.PullRequestLinks == nil {
				allIssues = append(allIssues, issue)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	f.logger.Info("Issue一覧を取得しました", zap.Int("count", len(allIssues)))
	return allIssues, nil
}

func (f *IssueFetcher) FetchIssue(ctx context.Context, owner, repo string, number int) (*github.Issue, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return nil, err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return nil, err
	}

	issue, _, err := f.client.GetClient().Issues.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("Issue詳細の取得に失敗しました: %w", err)
	}

	return issue, nil
}

func (f *IssueFetcher) FetchIssueComments(ctx context.Context, owner, repo string, number int) ([]*github.IssueComment, error) {
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
			return nil, fmt.Errorf("Issueコメントの取得に失敗しました: %w", err)
		}

		allComments = append(allComments, comments...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}
