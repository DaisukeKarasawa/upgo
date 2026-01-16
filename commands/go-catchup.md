---
description: golang/go の直近 PR をキャッチアップし、Go 思想を学ぶ
allowed-tools: Bash, WebFetch, Read
argument-hint: [件数] [カテゴリ]
---

# Go PR キャッチアップ

golang/go リポジトリの直近の PR を取得・分析し、Go の設計思想をキャッチアップします。

## 引数

- `$1`: 取得件数（デフォルト: 10）
- `$2`: カテゴリフィルタ（オプション: error-handling, performance, api-design, testing, runtime, compiler）

## 実行手順

### 1. 環境確認

```bash
# GitHub CLI 確認
which gh || echo "ERROR: gh コマンドが見つかりません。GitHub CLI をインストールしてください。"

# 認証確認
gh auth status
```

### 2. PR 一覧取得

```bash
# 直近のマージ済み PR を取得
LIMIT="${1:-10}"
gh pr list --repo golang/go --state merged --limit $LIMIT --json number,title,author,mergedAt,labels
```

### 3. 各 PR の分析

取得した各 PR について：

1. PR の詳細を取得
2. コメント・議論を確認
3. 変更内容を分析
4. Go 思想との関連を抽出

### 4. レポート作成

以下の形式でレポートを作成：

```markdown
# Go PR キャッチアップレポート

**期間**: <oldest_date> 〜 <newest_date>
**件数**: <count> 件

## サマリー

### カテゴリ別
- error-handling: X 件
- performance: Y 件
- ...

### 注目の PR
1. PR #XXXX: <title> - <一行説明>
2. PR #YYYY: <title> - <一行説明>

---

## 詳細分析

### PR #XXXX: <title>

**概要**: <要約>

**Go 思想**: <学べること>

---

## 今週の学び

<全体を通じて学べる Go の思想・ベストプラクティス>
```

## 使用例

```bash
# 直近 10 件をキャッチアップ
/go-catchup

# 直近 20 件をキャッチアップ
/go-catchup 20

# error-handling 関連のみ
/go-catchup 10 error-handling
```
