# Upgo - Go 公式リポジトリ PR 分析 → Claude Code Skills

golang/go リポジトリの PR を分析し、Go の設計思想・哲学・ベストプラクティスを Claude Code Skills として生成するツールです。

## 特徴

- **リアルタイムキャッチアップ**: 直近1ヶ月の PR 情報を取得・分析
- **議論の可視化**: PR のレビューコメントや議論のポイントを抽出
- **Go 思想の理解**: 変更の背景にある Go の設計思想を学べる
- **Skills 形式で配布**: Claude Code ですぐに使える形式で出力

## ワークフロー

```
golang/go の PR 取得
        ↓
  議論・レビュー・変更内容を収集
        ↓
  LLM（Ollama）で分析
        ↓
  Skills 形式で出力
        ↓
  Claude Code でキャッチアップ
```

## 必要な環境

- Go 1.21 以上
- Ollama（https://ollama.ai/）
- GitHub Token

## セットアップ

### 1. Ollama のインストールとモデル取得

```bash
# Ollama をインストール後
ollama pull llama3.2
```

### 2. 環境変数の設定

```bash
export GITHUB_TOKEN=your_github_token_here
```

### 3. 設定ファイルの作成

```bash
cp config.yaml.example config.yaml
```

### 4. 依存関係のインストール

```bash
go mod download
```

## 使い方

### フルパイプライン（推奨）

```bash
# PR 取得 → 分析 → Skills 生成を一括実行
go run cmd/skillgen/main.go run
```

### 個別コマンド

```bash
# PR データを同期（直近1ヶ月）
go run cmd/skillgen/main.go sync

# PR を分析
go run cmd/skillgen/main.go analyze

# Skills を生成
go run cmd/skillgen/main.go generate

# 生成された Skills を一覧表示
go run cmd/skillgen/main.go list
```

### 生成された Skills を使う

```bash
# 個人の Claude Code に追加
cp -r skills/* ~/.claude/skills/

# または、プロジェクトに追加
cp -r skills/* your-project/.claude/skills/
```

## 生成される Skills

実行すると、以下のようなスキルが生成されます：

| スキル | 説明 |
|--------|------|
| `go-error-handling` | エラーハンドリング関連の PR から抽出した知見 |
| `go-testing` | テスト関連の PR から抽出した知見 |
| `go-performance` | パフォーマンス関連の PR から抽出した知見 |
| `go-concurrency` | 並行処理関連の PR から抽出した知見 |
| `go-weekly-digest` | 全カテゴリのサマリー |

## Skills の内容例

```markdown
# Go エラーハンドリング の最新動向

## 注目のPR

### PR #12345: errors: add ErrUnsupported

**状態**: merged | **作成者**: rsc

**概要**: 新しい標準エラー ErrUnsupported を追加...

**議論のポイント**: os.ErrNotExist との一貫性について議論...

**Go思想**: シンプルで予測可能な API 設計を重視...
```

## ディレクトリ構造

```
upgo/
├── cmd/skillgen/          # CLI ツール
├── internal/
│   ├── analyzer/         # PR 分析サービス
│   ├── skillgen/         # Skills 生成サービス
│   ├── github/           # GitHub API クライアント
│   ├── llm/              # LLM クライアント（Ollama）
│   └── database/         # SQLite データストア
├── skills/               # Skills
│   ├── go-*/             # Go 関連スキル
│   └── zellij-workflow/  # Zellij ワークフロースキル
├── agents/               # サブエージェント
│   ├── go-mentor/        # Go メンター
│   └── zellij-orchestrator/ # Zellij オーケストレーター
├── commands/             # スラッシュコマンド
│   ├── go-review.md      # Go コードレビュー
│   ├── zellij-test.md    # 別ペインでテスト実行
│   ├── zellij-run.md     # 別ペインでコマンド実行
│   ├── orchestrator.md   # タスク並列実行
│   └── loop.md           # ペアプログラミング
└── legacy/               # 旧 Web UI
```

## 設定

`config.yaml` で以下を設定できます：

```yaml
repository:
  owner: "golang"  # 分析対象リポジトリ
  name: "go"

llm:
  base_url: "http://localhost:11434"
  model: "llama3.2"
  timeout: 300

skillgen:
  output_dir: "skills"  # Skills 出力先
```

## 定期実行

cron などで定期的に実行することで、最新の PR 情報をキャッチアップできます：

```bash
# 毎日朝9時に実行
0 9 * * * cd /path/to/upgo && go run cmd/skillgen/main.go run
```

## Zellij ワークフロー

zellij を使った並列開発ワークフローのための Skills / コマンド / サブエージェントを提供しています。

### スラッシュコマンド

| コマンド | 説明 |
|----------|------|
| `/zellij-test [pattern]` | 別ペインでテストを実行 |
| `/zellij-run <command>` | 別ペインで任意のコマンドを実行 |
| `/orchestrator <task>` | タスクを分割して複数ペインで並列実行 |
| `/loop [prompt]` | 別ペインでペアプログラミングパートナーを起動 |

### 使用例

```bash
# 別ペインでテスト実行
/zellij-test

# 特定のテストパターンのみ
/zellij-test TestUserService

# 開発サーバーを別ペインで起動
/zellij-run go run cmd/server/main.go

# タスクを並列実行
/orchestrator APIエンドポイントを実装してテストも書いて

# ペアプログラミング
/loop テストを書いてください
```

### Skills

`zellij-workflow` スキルに詳細なパターンとコマンドリファレンスがあります。

### セットアップ

zellij セッション内で Claude Code を使用してください：

```bash
# zellij セッション開始
zellij

# Claude Code 起動
claude
```

## Legacy: Web UI 版

以前の Web UI 版は `legacy/` ディレクトリに保存されています。

```bash
make legacy-dev
```

## ライセンス

MIT License

## 関連リンク

- [golang/go リポジトリ](https://github.com/golang/go)
- [Claude Code Skills ドキュメント](https://code.claude.com/docs/en/skills)
- [Ollama](https://ollama.ai/)
