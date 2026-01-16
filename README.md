# Upgo - Go Development Skills for Claude Code

Go言語の設計思想・哲学・ベストプラクティスを学べる Claude Code Skills セットです。

## 特徴

- **Go の思想を理解**: 「なぜ Go はこう設計されているか」を学べる
- **実践的なパターン**: エラーハンドリング、テスト、並行処理の実践的なパターン
- **コードレビュー観点**: Go らしいコードを書くためのチェックリスト
- **すぐに使える**: コピーするだけで導入完了

## スキル一覧

| スキル | 説明 |
|--------|------|
| `go-philosophy` | Go の設計思想と哲学（Go Proverbs、シンプルさ、明示性） |
| `go-error-handling` | エラーハンドリングパターン（ラッピング、カスタムエラー、sentinel errors） |
| `go-testing` | テスト戦略（TDD、テーブル駆動テスト、モック、ベンチマーク） |
| `go-concurrency` | 並行処理パターン（goroutine、channel、context、sync） |
| `go-code-review` | コードレビューガイド（チェックリスト、セキュリティ、パフォーマンス） |

## インストール方法

### 方法1: 手動コピー（推奨）

```bash
# リポジトリをクローン
git clone https://github.com/DaisukeKarasawa/upgo.git

# スキルをコピー（個人用）
cp -r upgo/skills/* ~/.claude/skills/

# または、プロジェクト単位で
cp -r upgo/skills/* your-project/.claude/skills/
```

### 方法2: エージェントとコマンドも含める

```bash
# スキル、エージェント、コマンドをすべてコピー
cp -r upgo/skills/* ~/.claude/skills/
cp -r upgo/agents/* ~/.claude/agents/
cp -r upgo/commands/* ~/.claude/commands/
```

### 方法3: プラグインマーケットプレイス（将来対応予定）

```bash
# Claude Code で実行
/plugin marketplace add DaisukeKarasawa/upgo
/plugin install go-skills@upgo
```

## 使い方

### スキル（自動適用）

インストール後、Claude Code が自動的にスキルを活用します。

```
# 例: Go のエラーハンドリングについて質問
> Go でエラーをどう処理すればいい？

# → go-error-handling スキルが自動的に適用され、
#   Go の思想に基づいた回答が得られます
```

### コマンド

```bash
# Go コードをレビュー
/go-review path/to/file.go

# Go コードを解説
/go-explain path/to/file.go
```

### エージェント

Go メンターエージェントが自動的に起動し、Go の学習をサポートします。

## ディレクトリ構造

```
upgo/
├── skills/                      # Claude Code Skills
│   ├── go-philosophy/          # Go の設計思想
│   ├── go-error-handling/      # エラーハンドリング
│   ├── go-testing/             # テスト戦略
│   ├── go-concurrency/         # 並行処理
│   └── go-code-review/         # コードレビュー
├── agents/                      # サブエージェント
│   └── go-mentor/              # Go メンター
├── commands/                    # スラッシュコマンド
│   ├── go-review.md            # コードレビュー
│   └── go-explain.md           # コード解説
├── .claude-plugin/              # プラグイン設定
└── legacy/                      # 旧 Web UI（参考用）
```

## 対象ユーザー

- Go を始めたばかりの開発者
- Go のベストプラクティスを学びたい中級者
- チームの Go コード品質を向上させたいリーダー
- Claude Code を Go 開発に活用したい人

## 参考資料

このスキルセットは以下の資料を参考に作成されています：

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Practical Go](https://dave.cheney.net/practical-go)

## Legacy: Web UI 版

以前の Web UI 版（GitHub PR 監視・分析システム）は `legacy/` ディレクトリに保存されています。

```bash
# Legacy Web UI を起動する場合
make legacy-dev
```

詳細は [legacy/README.md](legacy/README.md) を参照してください。

## ライセンス

MIT License

## コントリビューション

Issue や Pull Request を歓迎します。

- スキルの改善提案
- 新しいスキルの追加
- ドキュメントの改善

## 関連リンク

- [Claude Code Skills ドキュメント](https://code.claude.com/docs/en/skills)
- [anthropics/skills](https://github.com/anthropics/skills) - 公式 Skills リポジトリ
- [awesome-claude-skills](https://github.com/travisvn/awesome-claude-skills) - Skills キュレーションリスト
