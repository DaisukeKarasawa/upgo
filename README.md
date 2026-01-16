# Upgo - Claude Code Skills for Go Development & Zellij Workflow

Claude Code専用のGo開発・並列ワークフロースキルセットです。

## 特徴

- **Go開発スキル**: Go言語の思想、エラーハンドリング、テスト、並行処理のベストプラクティス
- **Zellij並列ワークフロー**: zellijを使った別ペインでのテスト実行、マルチエージェント構成
- **PR分析ツール**: golang/goリポジトリのPRを分析し、Go思想を抽出

## インストール

### 方法1: プラグインとしてインストール（推奨）

```bash
# Claude Code でプラグインを追加
/plugin marketplace add DaisukeKarasawa/upgo
```

### 方法2: 手動コピー

```bash
# リポジトリをクローン
git clone https://github.com/DaisukeKarasawa/upgo.git

# Skills をコピー
cp -r upgo/skills/* ~/.claude/skills/

# Commands をコピー
cp -r upgo/commands/* ~/.claude/commands/

# Agents をコピー
cp -r upgo/agents/* ~/.claude/agents/
```

### 方法3: プロジェクトローカル

```bash
# プロジェクトディレクトリにコピー
cp -r upgo/skills/* your-project/.claude/skills/
cp -r upgo/commands/* your-project/.claude/commands/
cp -r upgo/agents/* your-project/.claude/agents/
```

## 含まれるコンポーネント

### Skills

| スキル | 説明 |
|--------|------|
| `go-philosophy` | Goの設計思想とGo Proverbs |
| `go-error-handling` | エラーハンドリングパターン |
| `go-testing` | テスト戦略とTDD |
| `go-concurrency` | 並行処理パターン |
| `go-code-review` | コードレビュー観点 |
| `zellij-workflow` | zellij並列ワークフロー |

### Slash Commands

| コマンド | 説明 |
|----------|------|
| `/zellij-test [pattern]` | 別ペインでテストを実行 |
| `/zellij-run <command>` | 別ペインでコマンドを実行 |
| `/orchestrator <task>` | タスクを分割して並列実行 |
| `/loop [prompt]` | ペアプログラミングパートナーを起動 |
| `/go-review <file>` | Goコードをレビュー |
| `/go-explain <code>` | Goコードを解説 |

### Agents

| エージェント | 説明 |
|--------------|------|
| `go-mentor` | Go言語のメンター |
| `zellij-orchestrator` | 並列タスク実行オーケストレーター |

## 使い方

### Zellijワークフロー

zellij セッション内で Claude Code を使用してください：

```bash
# zellij セッション開始
zellij

# Claude Code 起動
claude
```

#### 別ペインでテスト実行

```
/zellij-test
```

右側のペインでテストが実行されます。

#### 別ペインでサーバー起動

```
/zellij-run go run cmd/server/main.go
```

下側のペインでサーバーが起動します。

#### タスクを並列実行

```
/orchestrator APIを実装してテストも書いて
```

タスクが分割され、複数ペインで並列実行されます。

#### ペアプログラミング

```
/loop テストを担当してください
```

右側のペインでペアプログラミングパートナーが起動します。

### Goスキルの活用

Goに関する質問をすると、自動的に関連スキルが適用されます：

- 「エラー処理どうすればいい？」→ `go-error-handling` スキル適用
- 「テストの書き方教えて」→ `go-testing` スキル適用
- 「goroutineの使い方」→ `go-concurrency` スキル適用

## PR分析ツール（オプション）

golang/goリポジトリのPRを分析し、Go思想を抽出するCLIツールも含まれています。

### 必要な環境

- Go 1.21以上
- Ollama（https://ollama.ai/）
- GitHub Token

### セットアップ

```bash
# Ollamaモデル取得
ollama pull llama3.2

# 環境変数設定
export GITHUB_TOKEN=your_token

# 設定ファイル作成
cp config.yaml.example config.yaml
```

### 実行

```bash
# フルパイプライン
go run cmd/skillgen/main.go run

# 個別コマンド
go run cmd/skillgen/main.go sync      # PR同期
go run cmd/skillgen/main.go analyze   # PR分析
go run cmd/skillgen/main.go generate  # Skills生成
go run cmd/skillgen/main.go list      # 生成済み一覧
```

## ディレクトリ構造

```
upgo/
├── .claude-plugin/       # Claude Code プラグインマニフェスト
│   └── plugin.json
├── skills/               # Skills
│   ├── go-*/             # Go関連スキル
│   └── zellij-workflow/  # Zellijワークフロースキル
├── agents/               # Agents
│   ├── go-mentor/        # Goメンター
│   └── zellij-orchestrator/ # Zellijオーケストレーター
├── commands/             # Slash Commands
│   ├── go-*.md           # Go関連コマンド
│   ├── zellij-*.md       # Zellij関連コマンド
│   ├── orchestrator.md   # タスク並列実行
│   └── loop.md           # ペアプログラミング
├── cmd/skillgen/         # PR分析CLIツール
├── internal/             # 内部パッケージ
└── legacy/               # 旧Web UI
```

## ライセンス

MIT License

## 関連リンク

- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
- [Claude Code Skills](https://code.claude.com/docs/en/skills)
- [Zellij](https://zellij.dev/)
- [golang/go](https://github.com/golang/go)
