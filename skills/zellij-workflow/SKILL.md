---
name: zellij-workflow
description: zellijを使った並列開発ワークフローのスキル。マルチペイン活用、タスク分散、CI/CD統合について質問されたときに使用。
---

# Zellij Workflow Patterns

zellijを使った効率的な並列開発ワークフローのパターン集。

## 基本コマンド

### ペイン操作

```bash
# 新しいペインを作成
zellij action new-pane                           # デフォルト（下）
zellij action new-pane --direction right         # 右側
zellij action new-pane --direction down          # 下側
zellij action new-pane --direction left          # 左側
zellij action new-pane --direction up            # 上側

# フローティングペイン
zellij action new-pane --floating --name "popup"

# ペインに名前を付ける
zellij action new-pane --name "test-runner"
```

### フォーカス移動

```bash
# 方向指定で移動
zellij action move-focus left
zellij action move-focus right
zellij action move-focus up
zellij action move-focus down

# 次/前のペインへ
zellij action focus-next-pane
zellij action focus-previous-pane
```

### コマンド送信

```bash
# 文字列を送信
zellij action write-chars "go test -v ./..."

# Enter キーを送信（ASCII 10 = LF）
zellij action write 10

# 組み合わせ（コマンド実行）
zellij action write-chars "make build" && zellij action write 10
```

### ペイン管理

```bash
# ペインを閉じる
zellij action close-pane

# ペインを最大化/戻す
zellij action toggle-fullscreen

# フレーム表示切替
zellij action toggle-pane-frames

# ペインの位置を交換
zellij action move-pane
```

## 開発ワークフローパターン

### パターン1: TDD ワークフロー

コードとテストを同時に見ながら開発：

```
┌─────────────────┬─────────────────┐
│                 │                 │
│   Main Editor   │   Test Runner   │
│   (開発作業)     │   (go test -v)  │
│                 │                 │
├─────────────────┴─────────────────┤
│                                   │
│         Test Output               │
│         (テスト結果)               │
│                                   │
└───────────────────────────────────┘
```

セットアップ:
```bash
# 右にテストランナー
zellij action new-pane --direction right --name "test-runner"
zellij action write-chars "go test -v ./... -count=1"
zellij action write 10
zellij action move-focus left

# 下にテスト出力
zellij action new-pane --direction down --name "test-output"
zellij action write-chars "watch -n 2 'go test ./... 2>&1 | tail -20'"
zellij action write 10
zellij action move-focus up
```

### パターン2: フルスタック開発

フロントエンドとバックエンドを同時開発：

```
┌─────────────────┬─────────────────┐
│                 │                 │
│   Backend Dev   │  Frontend Dev   │
│   (Go API)      │  (React/Vite)   │
│                 │                 │
├─────────────────┼─────────────────┤
│                 │                 │
│  Backend Server │ Frontend Server │
│  (go run ...)   │  (npm run dev)  │
│                 │                 │
└─────────────────┴─────────────────┘
```

セットアップ:
```bash
# 右にフロントエンド開発
zellij action new-pane --direction right --name "frontend"

# メインペインの下にバックエンドサーバー
zellij action move-focus left
zellij action new-pane --direction down --name "backend-server"
zellij action write-chars "go run cmd/server/main.go"
zellij action write 10

# フロントエンドの下にdevサーバー
zellij action move-focus right
zellij action new-pane --direction down --name "frontend-server"
zellij action write-chars "cd web && npm run dev"
zellij action write 10

# メイン開発ペインに戻る
zellij action move-focus up
zellij action move-focus left
```

### パターン3: CI パイプライン模擬

ローカルでCIパイプラインを並列実行：

```
┌─────────────────┬─────────────────┐
│   Lint Check    │   Type Check    │
│   (golangci)    │   (go vet)      │
├─────────────────┼─────────────────┤
│   Unit Tests    │  Build Check    │
│   (go test)     │   (go build)    │
└─────────────────┴─────────────────┘
```

セットアップ:
```bash
# 2x2グリッド作成
zellij action new-pane --direction right --name "type-check"
zellij action move-focus left
zellij action new-pane --direction down --name "unit-tests"
zellij action move-focus right
zellij action new-pane --direction down --name "build"

# 各ペインにコマンド送信
# ペイン1: lint
zellij action move-focus up
zellij action move-focus left
zellij action write-chars "golangci-lint run ./..."
zellij action write 10

# ペイン2: type check
zellij action move-focus right
zellij action write-chars "go vet ./..."
zellij action write 10

# ペイン3: unit tests
zellij action move-focus down
zellij action move-focus left
zellij action write-chars "go test -v ./..."
zellij action write 10

# ペイン4: build
zellij action move-focus right
zellij action write-chars "go build -o /dev/null ./..."
zellij action write 10
```

### パターン4: ログ監視 + 開発

アプリケーションログを監視しながら開発：

```
┌───────────────────────────────────┐
│                                   │
│         Main Editor               │
│         (開発作業)                 │
│                                   │
├─────────────────┬─────────────────┤
│   App Server    │   Log Viewer    │
│   (running)     │   (tail -f)     │
└─────────────────┴─────────────────┘
```

### パターン5: マルチエージェント Claude Code

複数のClaude Codeセッションで並列作業：

```
┌─────────────────┬─────────────────┐
│   Claude Code   │   Claude Code   │
│   (Feature A)   │   (Feature B)   │
├─────────────────┼─────────────────┤
│   Claude Code   │   Claude Code   │
│   (Tests)       │   (Docs)        │
└─────────────────┴─────────────────┘
```

セットアップ:
```bash
# 4つのペインを作成
zellij action new-pane --direction right --name "agent-2"
zellij action move-focus left
zellij action new-pane --direction down --name "agent-3"
zellij action move-focus right
zellij action new-pane --direction down --name "agent-4"

# 各ペインでClaude Codeを起動
# ペイン1 (現在位置)
zellij action move-focus up
zellij action move-focus left
zellij action write-chars "claude"
zellij action write 10

# ペイン2
zellij action move-focus right
zellij action write-chars "claude"
zellij action write 10

# ペイン3
zellij action move-focus down
zellij action move-focus left
zellij action write-chars "claude"
zellij action write 10

# ペイン4
zellij action move-focus right
zellij action write-chars "claude"
zellij action write 10
```

## タスク指示の送信

Claude Codeセッションにタスクを送信：

```bash
# --print オプションで直接タスクを渡す
zellij action write-chars "claude --print 'ユニットテストを作成してください'"
zellij action write 10

# または、起動後にプロンプトを送信
zellij action write-chars "Implement the user authentication feature"
zellij action write 10
```

## キーバインド設定（~/.config/zellij/config.kdl）

```kdl
keybinds {
    normal {
        // Ctrl+t でテストペインを開く
        bind "Ctrl t" {
            NewPane "Right" { name "tests"; }
            Write "go test -v ./...\n"
            MoveFocus "Left"
        }

        // Ctrl+s でサーバーペインを開く
        bind "Ctrl s" {
            NewPane "Down" { name "server"; }
            Write "go run cmd/server/main.go\n"
            MoveFocus "Up"
        }
    }
}
```

## ベストプラクティス

### 1. ペイン数は2-4が最適

多すぎると管理が困難になる。一般的な構成:
- 2ペイン: エディタ + テスト/サーバー
- 3ペイン: エディタ + テスト + サーバー
- 4ペイン: 並列タスク（CIパイプライン模擬）

### 2. 命名で識別性を高める

```bash
zellij action new-pane --name "test-runner"
zellij action new-pane --name "api-server"
zellij action new-pane --name "log-viewer"
```

### 3. フローティングペインの活用

一時的な作業にはフローティングペインが便利:

```bash
# フローティングペインでクイックテスト
zellij action new-pane --floating --name "quick-test"
zellij action write-chars "go test -v -run TestSpecific ./..."
zellij action write 10
```

### 4. セッション保存

作業構成をレイアウトとして保存:

```bash
# 現在のレイアウトを保存
zellij action dump-layout > ~/.config/zellij/layouts/dev.kdl
```

## トラブルシューティング

### ペインが作成できない

```bash
# zellijセッション内か確認
echo $ZELLIJ
# 空ならzellijセッション外

# セッション開始
zellij
```

### コマンドが送信されない

```bash
# write-chars は文字列を送信するだけ
# 実行にはEnterが必要
zellij action write-chars "command"
zellij action write 10  # Enter (LF)
```

### フォーカスが移動しない

```bash
# move-focusは方向指定が必要
zellij action move-focus left   # OK
zellij action move-focus        # NG
```
