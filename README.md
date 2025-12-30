# UpGo - Goリポジトリ監視システム

GitHubリポジトリ（Go言語プロジェクト）のPR/Issueを監視し、変更差分やコメントを収集・日本語要約・分析するWebアプリケーションです。

## 機能

- PR/Issueの自動監視（ポーリング）
- 変更差分の取得と保存
- コメントと議論の収集
- ローカルLLM（Ollama）を使用した日本語要約・分析
- Merge/Close理由の自動分析
- コミッターのメンタルモデル分析
- Web UIでの閲覧

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

- **バックエンド**: `http://localhost:8080` (Goサーバー)
- **フロントエンド**: `http://localhost:5173` (Vite開発サーバー)

ブラウザで `http://localhost:5173` にアクセスしてください。フロントエンドの変更は自動的に反映されます（HMR）。

**注意**:

- フロントエンドの開発サーバーは`http://localhost:5173`で起動します
- バックエンドAPIは`http://localhost:8080`で起動します
- Viteの設定で、`/api`へのリクエストは自動的にバックエンドにプロキシされます

#### 本番モード（ビルド済み）

フロントエンドをビルドしてから起動する場合：

```bash
make run
```

これにより、フロントエンドがビルドされ、バックエンドサーバーが起動します。
ブラウザで `http://localhost:8080` にアクセスしてください。

## 使い方

### 手動同期

Web UIの「同期」ボタンをクリックするか、以下のAPIを呼び出します：

```bash
curl -X POST http://localhost:8080/api/v1/sync
```

### バックアップ

```bash
make backup
# または
curl -X POST http://localhost:8080/api/v1/backup
```

### ヘルスチェック

```bash
curl http://localhost:8080/health
```

## 設定

`config.yaml`で以下の設定が可能です：

- `repository`: 監視対象リポジトリ
- `scheduler`: ポーリング間隔と有効/無効
- `llm`: Ollamaの設定（モデル、タイムアウトなど）
- `database`: データベースのパス
- `server`: サーバーのポートとホスト
- `logging`: ログレベルと出力先
- `backup`: バックアップ設定

詳細は `config.yaml.example` を参照してください。

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

```bash
make test
```

### クリーンアップ

ビルド成果物を削除：

```bash
make clean
```

## ライセンス

MIT
