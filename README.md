# Go PR Insights - Claude Code Plugin

golang/go の PR を Claude Code で自動取得・分析し、Go の設計思想をキャッチアップするプラグインです。

## 特徴

- **PR 自動取得**: GitHub CLI (gh) で golang/go の最新 PR を取得
- **議論の分析**: レビューコメント・議論のポイントを抽出
- **Go 思想の学習**: 変更の背景から Go の設計思想を学ぶ

## インストール

### 方法1: プラグインとしてインストール（推奨）

```bash
/plugin marketplace add DaisukeKarasawa/upgo
```

### 方法2: 手動コピー

```bash
git clone https://github.com/DaisukeKarasawa/upgo.git
cp -r upgo/skills/* ~/.claude/skills/
cp -r upgo/commands/* ~/.claude/commands/
```

## 必要な環境

- GitHub CLI (`gh`) がインストールされていること
- `gh auth login` で認証済みであること

## 使い方

### PR キャッチアップ

```bash
/go-catchup
```

直近 30 日間のマージ済み PR を取得・分析し、Go の設計思想をレポートします。

```bash
/go-catchup compiler
```

カテゴリフィルタを指定して取得できます(例: compiler, runtime など)。

### 個別 PR の分析

Claude Code に直接依頼：

```plaintext
golang/go の PR #12345 を分析して、Go の思想を教えて
```

## 含まれるコンポーネント

### Skills（ユーザー向け）

| スキル | 説明 |
|--------|------|
| `go-pr-fetcher` | GitHub CLI (gh) で PR を取得 |
| `go-pr-analyzer` | PR を分析し Go 思想を抽出 |

### Commands（ユーザー向け）

| コマンド | 説明 |
|----------|------|
| `/go-catchup [カテゴリ]` | 直近 30 日間の PR をキャッチアップ |

## コンポーネント間の相互作用

### ユーザーコマンドからPR分析までのフロー

1. **コマンド実行**: ユーザーが `/go-catchup` コマンドを実行
2. **PR取得**: `go-pr-fetcher` スキルが GitHub CLI (`gh`) を使用して golang/go リポジトリからPR情報を取得
3. **分析処理**: `go-pr-analyzer` スキルが取得したPR情報を分析し、Goの設計思想を抽出
4. **レポート生成**: 分析結果をまとめたレポートを生成

### GitHub CLIとの統合

- **PR情報取得**: `gh pr list` および `gh pr view` コマンドを使用してPRの基本情報、コメント、レビュー、差分を取得
- **認証**: `gh auth login` で認証済みである必要があります
- **エラーハンドリング**: GitHub CLIが見つからない場合や認証エラーが発生した場合、適切なエラーメッセージを表示

### SkillsとCommandsの関係

- **Commands**: ユーザーが直接実行するスラッシュコマンド（例: `/go-catchup`）
- **Skills**: Commandsから呼び出される再利用可能な機能モジュール
  - `go-pr-fetcher`: PR情報の取得を担当
  - `go-pr-analyzer`: PRの分析とGo思想の抽出を担当
- **役割分担**: Commandsがワークフローを定義し、Skillsが具体的な処理を実行する構造

## 分析で得られる情報

- **変更の背景**: なぜこの変更が必要だったか
- **議論のポイント**: レビューで何が議論されたか
- **Go 思想との関連**: シンプルさ、明示性、直交性など
- **学べること**: 実践的なベストプラクティス

## ディレクトリ構造

### ディレクトリの役割

- **`.claude-plugin/`**: プラグインマニフェストとメタデータ（plugin.json）
- **`skills/`**: Claude Codeによって自動読み込みされるエンドユーザー向けSkills
- **`commands/`**: Markdownで定義されたユーザー向けスラッシュコマンド
- **`.claude/`**: 開発/テスト/デバッグ用の内部開発者向けツール

### 構造

```plaintext
upgo/
├── .claude-plugin/       # プラグインマニフェスト
│   └── plugin.json
├── skills/               # ユーザー向け Skills
│   ├── go-pr-fetcher/    # PR 取得
│   └── go-pr-analyzer/   # PR 分析
├── commands/             # ユーザー向け Commands
│   └── go-catchup.md     # キャッチアップコマンド
└── .claude/              # 開発者向けツール（内部用）
    ├── skills/           # zellij, Go 開発支援
    ├── commands/         # 開発用コマンド
    └── agents/           # 開発用エージェント
```

## 開発に貢献する

プラグインの開発に参加したい方は [CONTRIBUTING.md](CONTRIBUTING.md) をご覧ください。

## ライセンス

MIT License
