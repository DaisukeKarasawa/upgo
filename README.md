# Upgo - Goリポジトリ監視システム

GitHubリポジトリ（Go言語プロジェクト）のPRを監視し、変更差分やコメントを収集・日本語要約・分析するWebアプリケーションです。

## 機能

- PRの更新チェック（軽量ポーリング）
- 変更差分の取得と保存（手動同期）
- コメントと議論の収集（手動同期）
- ローカルLLM（Ollama）を使用した日本語要約・分析（手動同期）
- Merge/Close理由の自動分析（手動同期）
- Web UIでの閲覧
- 同期ボタンに更新通知（赤丸）表示
- メンタルモデル分析（今後実装予定）

## 必要な環境

- Go 1.24以上（`go.mod`で`toolchain go1.24.11`が指定されています）
- Node.js 18以上
- Ollama（https://ollama.ai/）

## セットアップ

### 1. 必要なソフトウェアのインストール

```bash
# Goのインストール（未インストールの場合）
# https://golang.org/dl/

# Node.jsのインストール（未インストールの場合）
# https://nodejs.org/

# Ollamaのインストール
# https://ollama.ai/
```

### 2. Ollamaのセットアップ

```bash
# モデルのダウンロード
ollama pull llama3.2
```

### 3. 環境変数の設定

```bash
export GITHUB_TOKEN=your_github_token_here
```

GitHubトークンは以下の権限が必要です：

- `repo` (プライベートリポジトリの場合)
- `public_repo` (パブリックリポジトリの場合)

### 4. 設定ファイルの作成

```bash
cp config.yaml.example config.yaml
# config.yamlを編集（必要に応じて）
```

### 5. データディレクトリの作成

```bash
mkdir -p data backups logs
```

### 6. 依存関係のインストール

```bash
# バックエンド
go mod download

# フロントエンド
cd web && npm install
```

### 7. 起動

#### 開発モード

開発モードでは、バックエンドとフロントエンドの開発サーバーが同時に起動します：

```bash
make dev
```

これにより以下が起動します：

- **バックエンド**: `http://localhost:8081` (Goサーバー)
- **フロントエンド**: `http://localhost:5173` (Vite開発サーバー)

ブラウザで `http://localhost:5173` にアクセスしてください。フロントエンドの変更は自動的に反映されます（HMR）。

**注意**:

- フロントエンドの開発サーバーは`http://localhost:5173`で起動します
- バックエンドAPIは`http://localhost:8081`で起動します
- Viteの設定で、`/api`へのリクエストは自動的にバックエンドにプロキシされます

#### 本番モード（ビルド済み）

フロントエンドをビルドしてから起動する場合：

```bash
make run
```

これにより、フロントエンドがビルドされ、バックエンドサーバーが起動します。
ブラウザで `http://localhost:8081` にアクセスしてください。

## 使い方

### 更新チェックと同期

#### 更新チェック（自動）

サーバー起動時に設定された間隔で、GitHub上に更新があるかどうかを軽量チェックします。重い同期処理（取得/要約/分析）は実行しません。

- **ダッシュボード**: 直近1ヶ月に作成されたPRでDBに存在しないものがあるかチェック
- **PR詳細ページ**: 各PRが前回同期以降にGitHubで更新されたかチェック

更新がある場合は、Web UIの同期ボタン左上に赤丸が表示されます。

#### 手動同期

Web UIの「同期」ボタンをクリックするか、以下のAPIを呼び出します：

```bash
# 全体同期
curl -X POST http://localhost:8081/api/v1/sync

# 特定のPRを同期
curl -X POST http://localhost:8081/api/v1/prs/{id}/sync
```

手動同期では、PRの取得、コメント・差分の収集、要約・分析が実行されます。

### バックアップ

```bash
make backup
# または
curl -X POST http://localhost:8081/api/v1/backup
```

### データベースのクリア

SQLiteの全データをクリアする方法です。

#### セキュリティ警告

⚠️ **重要**: `/api/v1/clear` エンドポイントは現在、認証・認可の保護がありません。以下の点に注意してください：

- **開発環境**: 信頼できるネットワークでのみ使用してください
- **本番環境**: 以下のいずれかの対策を実装してください：
  - 認証ミドルウェア（APIキー、JWT、Basic認証など）の実装
  - ファイアウォールやIP許可リストによるアクセス制限
  - エンドポイントの無効化（方法2の使用を推奨）

本番環境では、認証が実装されるまで、**方法2（データベースファイルの直接削除）**の使用を強く推奨します。

#### データベースクリアの概要

- **What（何を）**: すべてのテーブルを削除し、マイグレーションを再実行して空のスキーマを再作成します
- **Where（どこで）**: SQLiteデータベースファイル（`data/upgo.db`）に対して実行されます
- **Why（なぜ）**: 開発環境でのリセット、テストシナリオ、データクリーンアップに使用されます
- **Technologies（技術）**:
  - SQLiteの `DROP TABLE` コマンド（[SQLite DROP TABLE](https://www.sqlite.org/lang_droptable.html)）
  - Goの `database/sql` パッケージ（[database/sql](https://pkg.go.dev/database/sql)）

#### 開発環境用

##### 方法1: APIエンドポイントを使用（推奨）

サーバーが起動している状態で：

```bash
curl -X POST -H "X-Confirm-Clear: yes" http://localhost:8081/api/v1/clear
```

**注意**: 確認ヘッダー `X-Confirm-Clear: yes` が必要です。これにより、誤操作を防ぎます。

これにより、すべてのテーブルが削除され、空の状態で再作成されます。

##### 方法2: データベースファイルを直接削除

サーバーを停止してから：

```bash
# データベースファイルを削除
rm data/upgo.db data/upgo.db-shm data/upgo.db-wal

# サーバーを再起動すると、自動的にマイグレーションが実行されて空のデータベースが作成されます
make dev
```

#### 本番環境用

##### 方法1: APIエンドポイントを使用

⚠️ **セキュリティ警告**: 認証が実装されるまで、本番環境ではこの方法の使用を避けてください。

本番サーバーが起動している状態で：

```bash
curl -X POST -H "X-Confirm-Clear: yes" http://your-server:8081/api/v1/clear
```

##### 方法2: データベースファイルを直接削除（より安全）

本番環境では、以下の手順を推奨します：

1. サーバーを停止
2. データベースファイルをバックアップ（念のため）
3. データベースファイルを削除
4. サーバーを再起動

```bash
# 1. サーバーを停止

# 2. バックアップ（念のため）
cp data/upgo.db data/upgo.db.backup.$(date +%Y%m%d_%H%M%S)

# 3. データベースファイルを削除
rm data/upgo.db data/upgo.db-shm data/upgo.db-wal

# 4. サーバーを再起動（自動的にマイグレーションが実行されます）
```

**注意**: データベースをクリアすると、すべてのPR、コメント、分析結果などのデータが失われます。本番環境で実行する場合は、事前にバックアップを取得することを強く推奨します。

### ヘルスチェック

```bash
curl http://localhost:8081/health
```

## 設定

`config.yaml`で以下の設定が可能です：

- `repository`: 監視対象リポジトリ
- `scheduler`: 更新チェックの間隔と有効/無効（重い同期処理は実行しません）
- `llm`: Ollamaの設定（モデル、タイムアウトなど）
- `database`: データベースのパス
- `server`: サーバーのポートとホスト
- `logging`: ログレベルと出力先
- `backup`: バックアップ設定

詳細は `config.yaml.example` を参照してください。

### 環境ごとのデータベース設定

本番環境と開発環境で異なるデータベースファイルを使用することを強く推奨します。

`config.yaml`で`database.dev`と`database.prd`の両方を設定すると、環境変数`UPGO_ENV`に基づいて自動的に適切なパスが選択されます。

#### 設定方法

`config.yaml`で以下のように設定：

```yaml
database:
  dev: "./data/upgo-dev.db" # 開発環境用
  prd: "./data/upgo.db" # 本番環境用
```

#### 使用技術

この機能は以下の技術を使用して実装されています：

- **SQLite**: データベースエンジン（[SQLite Documentation](https://www.sqlite.org/docs.html)）
- **Viper**: 設定管理ライブラリ（[github.com/spf13/viper](https://github.com/spf13/viper)）
- **環境変数**: Goの`os.Getenv`を使用してUPGO_ENVを読み取り、適切なデータベースパスを選択（[os package](https://pkg.go.dev/os#Getenv)）

#### 環境の切り替え

##### 開発環境（デフォルト）

環境変数を設定しない場合、自動的に`database.dev`が使用されます：

```bash
make dev
```

##### 本番環境

環境変数`UPGO_ENV=production`または`UPGO_ENV=prod`を設定すると、`database.prd`が使用されます：

```bash
export UPGO_ENV=production
make run
```

**注意**: 本番環境と開発環境で同じデータベースファイルを使用すると、開発中の操作が本番データに影響を与える可能性があります。必ず別々のデータベースファイルを使用してください。

## 開発

### 開発サーバーの起動

開発時は、`make dev`を実行するとバックエンドとフロントエンドの開発サーバーが同時に起動します：

```bash
make dev
```

**個別に起動する場合：**

バックエンドのみ起動：

```bash
make dev-backend
# または
go run cmd/server/main.go
```

フロントエンドのみ起動：

```bash
make dev-frontend
# または
cd web && npm run dev
```

### ビルド

本番用にビルドする場合：

```bash
make build
```

これにより、バックエンドとフロントエンドの両方がビルドされます。

### テスト

#### 通常のテスト

```bash
make test
```

#### ベンチマークテスト

パフォーマンスを計測するベンチマークテストを実行：

```bash
# ベンチマークテストを実行
make bench

# 詳細な出力付きでベンチマークテストを実行
make bench-verbose
```

#### パフォーマンス計測テスト

実行時間や処理量を詳細に計測するテストを実行：

```bash
# パフォーマンス計測テストを実行
make perf-test

# すべてのテスト（通常テスト + ベンチマーク + パフォーマンス）を実行
make test-all
```

計測できる項目：

- 同期処理の実行時間
- データベース操作のパフォーマンス
- LLM呼び出しの実行時間（Ollamaが起動している場合）
- 処理量（メモリ割り当て、操作回数など）

### クリーンアップ

ビルド成果物を削除：

```bash
make clean
```
