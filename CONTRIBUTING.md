# 開発者向けガイド

upgo プラグインの開発に貢献していただきありがとうございます。このドキュメントでは、プラグイン開発に必要な情報をまとめています。

## 目次

- [開発環境のセットアップ](#開発環境のセットアップ)
- [開発用ツール](#開発用ツール)
- [開発ワークフロー](#開発ワークフロー)
- [プラグイン開発のベストプラクティス](#プラグイン開発のベストプラクティス)
- [ディレクトリ構造](#ディレクトリ構造)
- [コミット規約](#コミット規約)

## 開発環境のセットアップ

### 必要なツール

```bash
# 必須
- curl (HTTP リクエスト用)
- Claude Code CLI

# オプション（推奨）
- zellij (並列開発用)
```

### インストール手順

```bash
# curl (通常は既にインストール済み)
# macOS
# 通常は標準インストール済み

# Linux
# 通常は標準インストール済み、ない場合は:
# sudo apt-get install curl  # Debian/Ubuntu
# sudo yum install curl       # RHEL/CentOS

# Gerrit 認証情報の設定
export GERRIT_USER="your-username"
export GERRIT_HTTP_PASSWORD="your-http-password"
# HTTP パスワードは以下から取得:
# https://go-review.googlesource.com/settings/#HTTPCredentials

# zellij（オプション）
# macOS
brew install zellij

# Linux
cargo install zellij
```

### ローカル開発用セットアップ

```bash
# リポジトリをクローン
git clone https://github.com/DaisukeKarasawa/upgo.git
cd upgo

# 開発用ツールをローカルの Claude Code にリンク（オプション）
ln -s $(pwd)/.claude/skills/* ~/.claude/skills/
ln -s $(pwd)/.claude/commands/* ~/.claude/commands/
```

## 開発用ツール

`.claude/` ディレクトリには、プラグイン開発を効率化するための内部ツールが含まれています。

### 開発用 Skills

| スキル              | 説明                                   |
| ------------------- | -------------------------------------- |
| `plugin-test`       | プラグインの構造と機能をテスト         |
| `go-code-review`    | Go コードレビューのチェックリスト      |
| `go-philosophy`     | Go の設計思想と哲学                    |
| `go-error-handling` | エラーハンドリングのベストプラクティス |
| `go-testing`        | テスト戦略と TDD                       |
| `go-concurrency`    | 並行処理パターン                       |
| `zellij-workflow`   | zellij での並列開発ワークフロー        |

### 開発用 Commands

| コマンド        | 説明                               | 引数             |
| --------------- | ---------------------------------- | ---------------- |
| `/test-plugin`  | プラグインテストを別ペインで実行   | なし             |
| `/go-review`    | Go コードをレビュー                | `<file-path>`    |
| `/go-explain`   | Go コードを解説                    | `<file-path>`    |
| `/zellij-test`  | テストを別ペインで実行             | `[test-pattern]` |
| `/zellij-run`   | コマンドを別ペインで実行           | `<command>`      |
| `/loop`         | ペアプログラミングパートナーを起動 | なし             |
| `/orchestrator` | タスクを並列実行                   | `<task>`         |

### 開発用 Agents

| エージェント          | 説明                                            |
| --------------------- | ----------------------------------------------- |
| `go-mentor`           | Go メンター（設計思想・ベストプラクティス指導） |
| `zellij-orchestrator` | 並列タスク実行のオーケストレータ                |

## 開発ワークフロー

### プラグイン構造のテスト

プラグインの構造と機能を検証します。

```bash
/test-plugin
```

別ペインで以下のテストが実行されます:

- ファイル構造の検証
- `plugin.json` の妥当性チェック
- 環境チェック（`curl` コマンド、Gerrit 認証情報）
- Skills/Commands の定義フォーマット検証
- 基本的な機能テスト（Change 取得）

### コードレビュー

Go コードレビューのチェックリストに基づいて改善点を提案します。

```bash
/go-review path/to/file.go
```

レビュー観点:

- エラーハンドリング
- リソース管理
- 並行処理の安全性
- コードスタイル
- パフォーマンス
- セキュリティ

### 並列でテスト実行

zellij セッション内で:

```bash
# 全テスト実行
/zellij-test

# 特定のテストのみ実行
/zellij-test TestFunctionName
```

テストが別ペインで実行されるため、メインペインで作業を続けられます。

### 並列開発のワークフロー

複数のタスクを同時に実行:

```bash
# タスクを別ペインで実行
/zellij-run go run main.go

# ペアプログラミングモード
/loop

# 複数タスクのオーケストレーション
/orchestrator "複数のタスクを並列実行"
```

## プラグイン開発のベストプラクティス

### ファイル構造

必須ファイル:

- `.claude-plugin/plugin.json` - プラグインメタデータ
- `skills/*/SKILL.md` - ユーザー向け Skills
- `commands/*.md` - ユーザー向け Commands
- `README.md` - ユーザー向けドキュメント

開発用ツール（オプション）:

- `.claude/skills/` - 開発用 Skills
- `.claude/commands/` - 開発用 Commands
- `.claude/agents/` - 開発用 Agents

### テスト駆動開発（TDD）

1. 機能を追加する前に `/test-plugin` で現在の状態を確認
2. 新機能を実装
3. 再度 `/test-plugin` で検証
4. すべてのテストが通ることを確認してからコミット

### コードレビュー

コミット前に必ず:

```bash
/go-review path/to/modified/file.go
```

### プラグインマニフェスト（plugin.json）

必須フィールド:

- `name`: プラグイン名
- `version`: バージョン（セマンティックバージョニング）
- `description`: 簡潔な説明
- `author`: 作者情報
- `repository`: リポジトリ URL
- `license`: ライセンス

例:

```json
{
  "name": "upgo",
  "version": "1.0.0",
  "description": "golang/go の Change (CL) を分析し Go の設計思想を学ぶ",
  "author": {
    "name": "Your Name",
    "url": "https://github.com/yourname"
  },
  "repository": "https://github.com/yourname/upgo",
  "license": "MIT"
}
```

## ディレクトリ構造

```
upgo/
├── .claude-plugin/           # プラグインマニフェスト
│   └── plugin.json          # プラグインメタデータ
│
├── skills/                   # ユーザー向け Skills
│   ├── go-pr-fetcher/       # Change 取得スキル
│   │   └── SKILL.md
│   └── go-pr-analyzer/      # Change 分析スキル
│       └── SKILL.md
│
├── commands/                 # ユーザー向け Commands
│   └── go-catchup.md        # Change キャッチアップコマンド
│
├── .claude/                  # 開発者向けツール（内部用）
│   ├── skills/              # 開発用 Skills
│   │   ├── plugin-test/     # プラグインテストスキル
│   │   │   └── SKILL.md
│   │   ├── go-code-review/  # コードレビュースキル
│   │   │   └── SKILL.md
│   │   ├── go-philosophy/   # Go 設計思想スキル
│   │   │   └── SKILL.md
│   │   ├── go-error-handling/ # エラー処理スキル
│   │   │   └── SKILL.md
│   │   ├── go-testing/      # テスト戦略スキル
│   │   │   └── SKILL.md
│   │   ├── go-concurrency/  # 並行処理スキル
│   │   │   └── SKILL.md
│   │   └── zellij-workflow/ # zellij ワークフロースキル
│   │       └── SKILL.md
│   │
│   ├── commands/            # 開発用 Commands
│   │   ├── test-plugin.md   # プラグインテスト実行
│   │   ├── go-review.md     # コードレビュー実行
│   │   ├── go-explain.md    # コード解説
│   │   ├── zellij-test.md   # テスト並列実行
│   │   ├── zellij-run.md    # コマンド並列実行
│   │   ├── loop.md          # ペアプログラミング
│   │   └── orchestrator.md  # タスクオーケストレーション
│   │
│   ├── agents/              # 開発用 Agents
│   │   ├── go-mentor/       # Go メンターエージェント
│   │   │   └── AGENT.md
│   │   └── zellij-orchestrator/ # 並列タスクエージェント
│   │       └── AGENT.md
│   │
│   └── CLAUDE.md            # プロジェクト固有の Claude 指示
│
├── README.md                # ユーザー向けドキュメント
├── CONTRIBUTING.md          # 開発者向けガイド（このファイル）
└── LICENSE                  # ライセンスファイル
```

### ディレクトリの役割

#### ユーザー向け（配布対象）

- `.claude-plugin/`: プラグインメタデータ
- `skills/`: エンドユーザーが利用する Skills
- `commands/`: エンドユーザーが利用する Commands

#### 開発者向け（内部用）

- `.claude/skills/`: プラグイン開発を支援する Skills
- `.claude/commands/`: 開発タスクを効率化する Commands
- `.claude/agents/`: 特殊な開発タスク用の Agents

## コミット規約

[gitmoji](https://gist.github.com/parmentf/035de27d6ed1dce0b36a) を使用してコミットメッセージを記述します。

### よく使う gitmoji

| Gitmoji              | 使用場面           | 例                                            |
| -------------------- | ------------------ | --------------------------------------------- |
| `:sparkles:`         | 新機能追加         | `:sparkles: Add Change filtering by label`    |
| `:bug:`              | バグ修正           | `:bug: Fix error handling in Change fetcher`  |
| `:recycle:`          | リファクタリング   | `:recycle: Refactor Change analysis logic`    |
| `:white_check_mark:` | テスト追加・更新   | `:white_check_mark: Add tests for go-catchup` |
| `:memo:`             | ドキュメント更新   | `:memo: Update README with new features`      |
| `:art:`              | コード構造改善     | `:art: Improve SKILL.md formatting`           |
| `:zap:`              | パフォーマンス改善 | `:zap: Optimize Change fetching logic`        |
| `:fire:`             | コード削除         | `:fire: Remove deprecated skill`              |
| `:construction:`     | WIP                | `:construction: Work in progress on analyzer` |

### コミットメッセージの形式

```
<gitmoji> <簡潔な説明>

<詳細な説明（オプション）>
```

### 例

```bash
git commit -m ":sparkles: Add label filtering to go-pr-fetcher

Users can now filter Changes by specific labels when using
the go-pr-fetcher skill."
```

### コミットサイズ

- 小さく頻繁にコミット
- 1つのコミットで1つの論理的な変更
- WIPコミットは避ける（完成してからコミット）

## プルリクエストのガイドライン

### PR を作成する前に

1. `/test-plugin` で全テストが通ることを確認
2. 変更したファイルを `/go-review` でレビュー
3. コミットメッセージが gitmoji 規約に従っているか確認

### PR タイトル

コミットメッセージと同様に gitmoji を使用:

```
:sparkles: Add new feature for Change filtering
```

### PR 説明

以下を含めてください:

- **変更内容**: 何を変更したか
- **理由**: なぜこの変更が必要か
- **テスト**: どのようにテストしたか
- **スクリーンショット**: UI変更がある場合

## サポート

質問や問題がある場合:

- [Issues](https://github.com/DaisukeKarasawa/upgo/issues) で報告
- 開発用ツール（`/test-plugin`, `/go-review` など）を活用
- `go-mentor` エージェントに質問

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) をご覧ください。
