# Go Change Insights - Claude Code Plugin

golang/go の Change (CL) を Claude Code で自動取得・分析し、Go の設計思想をキャッチアップするプラグインです。

## 特徴

- **Change 自動取得**: Gerrit REST API で golang/go の最新 Change を取得
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

- `curl` がインストールされていること
- Gerrit の認証情報が設定されていること:
  - `GERRIT_USER`: Gerrit ユーザー名
  - `GERRIT_HTTP_PASSWORD`: Gerrit HTTP パスワード（[設定ページ](https://go-review.googlesource.com/settings/#HTTPCredentials)から取得）
  - `GERRIT_BASE_URL`: Gerrit サーバーURL（オプション、デフォルト: `https://go-review.googlesource.com`）

## 使い方

### Change キャッチアップ

```bash
/go-catchup
```

直近 30 日間のマージ済み Change を取得・分析し、Go の設計思想をレポートします。

```bash
/go-catchup compiler
```

カテゴリフィルタを指定して取得できます(例: compiler, runtime など)。

### 個別 Change の分析

Claude Code に直接依頼：

```
golang/go の Change #3965 を分析して、Go の思想を教えて
```

## 含まれるコンポーネント

### Skills（ユーザー向け）

| スキル           | 説明                             |
| ---------------- | -------------------------------- |
| `go-pr-fetcher`  | Gerrit REST API で Change を取得 |
| `go-pr-analyzer` | Change を分析し Go 思想を抽出    |

### Commands（ユーザー向け）

| コマンド                 | 説明                                   |
| ------------------------ | -------------------------------------- |
| `/go-catchup [カテゴリ]` | 直近 30 日間の Change をキャッチアップ |

## コンポーネント間の相互作用

### ユーザーコマンドから Change 分析までのフロー

1. **コマンド実行**: ユーザーが `/go-catchup` コマンドを実行
2. **Change 取得**: `go-pr-fetcher` スキルが Gerrit REST API を使用して golang/go リポジトリから Change 情報を取得
3. **分析処理**: `go-pr-analyzer` スキルが取得した Change 情報を分析し、Goの設計思想を抽出
4. **レポート生成**: 分析結果をまとめたレポートを生成

### Gerrit との統合

- **Change 情報取得**: Gerrit REST API（`curl` + 認証）で Change 一覧・詳細・コメント・パッチを取得。`go-pr-fetcher` スキル内の `gerrit_api` ヘルパーを参照
- **認証**: `GERRIT_USER` と `GERRIT_HTTP_PASSWORD` を設定（[設定ページ](https://go-review.googlesource.com/settings/#HTTPCredentials)）
- **エラーハンドリング**: `curl` または認証情報が未設定の場合はエラーメッセージを表示

### SkillsとCommandsの関係

- **Commands**: ユーザーが直接実行するスラッシュコマンド（例: `/go-catchup`）
- **Skills**: Commandsから呼び出される再利用可能な機能モジュール
  - `go-pr-fetcher`: Change 情報の取得を担当
  - `go-pr-analyzer`: Change の分析と Go 思想の抽出を担当
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
│   ├── go-pr-fetcher/    # Change 取得
│   └── go-pr-analyzer/   # Change 分析
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
