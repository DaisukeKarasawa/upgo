# Go Change Insights - Claude Code Plugin

golang/go の Change (CL) を Claude Code で自動取得・分析し、Go の設計思想をキャッチアップするプラグインです。

## 特徴

- **Change 自動取得**: Gerrit REST API で golang/go の最新 Change を取得
- **議論の分析**: レビューコメント・議論のポイントを抽出
- **Go 思想の学習**: 変更の背景から Go の設計思想を学ぶ

## 重要な注意事項

### ネットワークアクセス

- **ネットワークアクセス**: すべてのスキルとコマンドは Gerrit サーバー（`https://go-review.googlesource.com`）へのネットワークアクセスを必要とします
- **匿名アクセス**: プラグインは認証なしで動作します（匿名 API アクセスを使用）。`go-review.googlesource.com` は現在匿名アクセスをサポートしていますが、一部の Gerrit インスタンスでは匿名アクセスが許可されていない場合があります（401/403 エラーが返る可能性）。また、匿名アクセスではレート制限がより厳しい場合があります。
- **自動起動**: スキル（`go-pr-fetcher`, `go-pr-analyzer`）は Claude が会話の流れから自動的に起動する可能性があります。ネットワークアクセスを事前に確認してください

### 必要なコマンド

- `curl`: HTTP クライアント
- `jq`: JSON 処理
- `sed`: テキスト処理（XSSI プレフィックス除去用）

## インストール

### 方法1: プラグインとしてインストール（推奨）

マーケットプレイスを追加してプラグインをインストールします：

```bash
# 1. マーケットプレイスを追加
/plugin marketplace add daisukekarasawa/upgo

# 2. プラグインをインストール
/plugin install go-pr-insights@daisukekarasawa-upgo
```

または、`/plugin` コマンドを実行してインタラクティブUIからインストールすることもできます：

1. `/plugin` を実行
2. **Discover** タブで `go-pr-insights` を検索
3. プラグインを選択してインストール

### 方法2: 手動コピー

```bash
git clone https://github.com/DaisukeKarasawa/upgo.git
cp -r upgo/skills/* ~/.claude/skills/
cp -r upgo/commands/* ~/.claude/commands/
```

## 必要な環境

### 必須コマンド

- `curl`: HTTP クライアント
- `jq`: JSON 処理
- `sed`: テキスト処理

### オプション環境変数

- `GERRIT_BASE_URL`: Gerrit サーバーURL（デフォルト: `https://go-review.googlesource.com`）

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

スキルは自動的に起動され、Change 情報を取得・分析します。

## 含まれるコンポーネント

### Skills（ユーザー向け）

| スキル                | 説明                                                        | 自動起動 |
| --------------------- | ----------------------------------------------------------- | -------- |
| `go-pr-fetcher`       | Gerrit REST API で Change を取得                            | あり     |
| `go-pr-analyzer`      | Change を分析し Go 思想を抽出                               | あり     |
| `go-gerrit-reference` | Gerrit API の共通参照（ヘルパー関数・エラーハンドリング等） | なし     |

### Commands（ユーザー向け）

#### オーケストレーターコマンド（ワークフロー）

| コマンド                 | 説明                                                         |
| ------------------------ | ------------------------------------------------------------ |
| `/go-catchup [カテゴリ]` | 直近 30 日間の Change をキャッチアップ（取得→分析→レポート） |

#### プリミティブコマンド（単一目的）

| コマンド                                    | 説明                                 |
| ------------------------------------------- | ------------------------------------ |
| `/go-changes-fetch [days] [status] [limit]` | Change 一覧を取得（JSON出力）        |
| `/go-change-analyze <change-id>`            | 単一 Change を分析して Go 思想を抽出 |

**命名規則**: `commands/NAMING.md` を参照。`go-` プレフィックスで名前空間化し、1コマンド=1目的を原則としています。

## コンポーネント間の相互作用

### ユーザーコマンドから Change 分析までのフロー

#### オーケストレーターコマンド（推奨）

1. **コマンド実行**: ユーザーが `/go-catchup` コマンドを実行
2. **Change 取得**: `go-pr-fetcher` スキルが Gerrit REST API を使用して golang/go リポジトリから Change 情報を取得
3. **分析処理**: `go-pr-analyzer` スキルが取得した Change 情報を分析し、Goの設計思想を抽出
4. **レポート生成**: 分析結果をまとめたレポートを生成

#### プリミティブコマンド（柔軟な組み合わせ）

- `/go-changes-fetch [days] [status] [limit]`: Change 一覧を取得（JSON出力）
- `/go-change-analyze <change-id>`: 特定の Change を分析

これらを組み合わせてカスタムワークフローを構築できます。

### Gerrit との統合

- **Change 情報取得**: Gerrit REST API（`curl` + 匿名アクセス）で Change 一覧・詳細・コメント・パッチを取得
- **共通参照**: `go-gerrit-reference` スキルが `gerrit_api()` ヘルパー関数、エラーハンドリングパターンを提供
- **匿名アクセス**: 認証なしで動作します。`go-review.googlesource.com` は現在匿名アクセスをサポートしていますが、一部の Gerrit インスタンスでは匿名アクセスが許可されていない場合があります（401/403 エラーが返る可能性）
- **エラーハンドリング**: `curl` が見つからない場合や、Gerrit サーバーが匿名アクセスを許可していない場合はエラーメッセージを表示

### SkillsとCommandsの関係

- **Commands**: ユーザーが直接実行するスラッシュコマンド（例: `/go-catchup`）
- **Skills**: Commandsから呼び出される再利用可能な機能モジュール
  - `go-pr-fetcher`: Change 情報の取得を担当
  - `go-pr-analyzer`: Change の分析と Go 思想の抽出を担当
  - `go-gerrit-reference`: Gerrit API の共通参照（他のスキルから参照される）
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

### 構造

```plaintext
upgo/
├── .claude-plugin/              # プラグインマニフェスト
│   ├── plugin.json
│   └── marketplace.json
├── skills/                       # ユーザー向け Skills
│   ├── go-pr-fetcher/            # Change 取得
│   │   └── SKILL.md
│   ├── go-pr-analyzer/           # Change 分析
│   │   └── SKILL.md
│   └── go-gerrit-reference/      # Gerrit API 共通参照
│       ├── SKILL.md
│       └── REFERENCE.md
└── commands/                     # ユーザー向け Commands
    ├── CLAUDE.md                 # エージェント向けドキュメント
    ├── NAMING.md                 # 命名規則ガイド
    ├── go-catchup.md             # キャッチアップコマンド
    ├── go-change-analyze.md      # Change 分析コマンド
    └── go-changes-fetch.md       # Change 取得コマンド
```

## ライセンス

MIT License
