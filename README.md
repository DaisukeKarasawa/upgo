# Upgo - Claude Code Skills for Go Development & Zellij Workflow

Claude Code 専用の Go 開発・並列ワークフロースキルセットです。

## 特徴

- **Go 開発スキル**: Go 言語の思想、エラーハンドリング、テスト、並行処理のベストプラクティス
- **Zellij 並列ワークフロー**: zellij を使った別ペインでのテスト実行、マルチエージェント構成

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
| `go-philosophy` | Go の設計思想と Go Proverbs |
| `go-error-handling` | エラーハンドリングパターン |
| `go-testing` | テスト戦略と TDD |
| `go-concurrency` | 並行処理パターン |
| `go-code-review` | コードレビュー観点 |
| `zellij-workflow` | zellij 並列ワークフロー |

### Slash Commands

| コマンド | 説明 |
|----------|------|
| `/zellij-test [pattern]` | 別ペインでテストを実行 |
| `/zellij-run <command>` | 別ペインでコマンドを実行 |
| `/orchestrator <task>` | タスクを分割して並列実行 |
| `/loop [prompt]` | ペアプログラミングパートナーを起動 |
| `/go-review <file>` | Go コードをレビュー |
| `/go-explain <code>` | Go コードを解説 |

### Agents

| エージェント | 説明 |
|--------------|------|
| `go-mentor` | Go 言語のメンター |
| `zellij-orchestrator` | 並列タスク実行オーケストレーター |

## 使い方

### Zellij ワークフロー

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
/orchestrator API を実装してテストも書いて
```

タスクが分割され、複数ペインで並列実行されます。

#### ペアプログラミング

```
/loop テストを担当してください
```

右側のペインでペアプログラミングパートナーが起動します。

### Go スキルの活用

Go に関する質問をすると、自動的に関連スキルが適用されます：

- 「エラー処理どうすればいい？」→ `go-error-handling` スキル適用
- 「テストの書き方教えて」→ `go-testing` スキル適用
- 「goroutine の使い方」→ `go-concurrency` スキル適用

## ディレクトリ構造

```
upgo/
├── .claude-plugin/       # Claude Code プラグインマニフェスト
│   └── plugin.json
├── skills/               # Skills
│   ├── go-*/             # Go 関連スキル
│   └── zellij-workflow/  # Zellij ワークフロースキル
├── agents/               # Agents
│   ├── go-mentor/        # Go メンター
│   └── zellij-orchestrator/ # Zellij オーケストレーター
└── commands/             # Slash Commands
    ├── go-*.md           # Go 関連コマンド
    ├── zellij-*.md       # Zellij 関連コマンド
    ├── orchestrator.md   # タスク並列実行
    └── loop.md           # ペアプログラミング
```

## ライセンス

MIT License

## 関連リンク

- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
- [Claude Code Skills](https://code.claude.com/docs/en/skills)
- [Zellij](https://zellij.dev/)
