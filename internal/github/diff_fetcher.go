package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v60/github"
)

type DiffFetcher struct {
	client *Client
}

func NewDiffFetcher(client *Client) *DiffFetcher {
	return &DiffFetcher{
		client: client,
	}
}

func (f *DiffFetcher) FetchCommitDiff(ctx context.Context, owner, repo, sha string) (string, error) {
	if err := f.client.waitForRateLimit(ctx); err != nil {
		return "", err
	}

	if err := f.client.checkRateLimit(ctx); err != nil {
		return "", err
	}

	commit, _, err := f.client.GetClient().Repositories.GetCommit(ctx, owner, repo, sha, nil)
	if err != nil {
		return "", fmt.Errorf("コミット詳細の取得に失敗しました: %w", err)
	}

	// コミットのdiffを取得
	diff, _, err := f.client.GetClient().Repositories.GetCommitRaw(ctx, owner, repo, sha, github.RawOptions{
		Type: github.Diff,
	})
	if err != nil {
		return "", fmt.Errorf("コミット差分の取得に失敗しました: %w", err)
	}

	_ = commit // 使用しないが将来の拡張のために保持
	return diff, nil
}
