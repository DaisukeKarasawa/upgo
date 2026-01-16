---
name: go-pr-fetcher
description: |
  golang/go リポジトリから PR 情報を取得するスキル。
  ユーザーが「Go の PR を取得」「golang/go の最新 PR」「Go の変更を確認」などと言った場合に使用。
allowed-tools:
  - Bash
  - WebFetch
---

# Go PR Fetcher

golang/go リポジトリから PR 情報を GitHub API で取得します。

## 前提条件

環境変数 `GITHUB_TOKEN` が設定されていること。

```bash
echo $GITHUB_TOKEN
```

設定されていない場合は、ユーザーに設定を促してください。

## PR 一覧の取得

### 直近の PR を取得（デフォルト: 30件）

```bash
gh pr list --repo golang/go --state all --limit 30 --json number,title,state,author,createdAt,updatedAt,labels
```

### マージ済み PR のみ取得

```bash
gh pr list --repo golang/go --state merged --limit 30 --json number,title,author,mergedAt,labels
```

### 特定期間の PR を取得

```bash
# 直近1週間
gh pr list --repo golang/go --state all --search "updated:>=$(date -v-7d +%Y-%m-%d)" --limit 50 --json number,title,state,author,updatedAt
```

## 個別 PR の詳細取得

### PR の基本情報

```bash
gh pr view <PR_NUMBER> --repo golang/go --json number,title,body,state,author,labels,comments,reviews
```

### PR のコメント・議論

```bash
gh pr view <PR_NUMBER> --repo golang/go --comments
```

### PR の変更ファイル

```bash
gh pr diff <PR_NUMBER> --repo golang/go
```

## 出力フォーマット

取得した PR 情報は以下の形式で整理してください：

```markdown
## PR #<number>: <title>

**状態**: <state> | **作成者**: <author> | **更新日**: <updatedAt>

**ラベル**: <labels>

### 概要
<body の要約>

### 変更ファイル
- <file1>
- <file2>
```

## エラーハンドリング

- `gh` コマンドが見つからない場合: GitHub CLI のインストールを案内
- 認証エラーの場合: `gh auth login` を案内
- レート制限の場合: しばらく待ってから再試行を案内
