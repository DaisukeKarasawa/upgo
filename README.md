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

## 分析で得られる情報

- **変更の背景**: なぜこの変更が必要だったか
- **議論のポイント**: レビューで何が議論されたか
- **Go 思想との関連**: シンプルさ、明示性、直交性など
- **学べること**: 実践的なベストプラクティス

## ディレクトリ構造

```
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
