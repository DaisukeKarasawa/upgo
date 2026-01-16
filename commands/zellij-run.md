---
description: zellijの別ペインでコマンドを実行
allowed-tools: Bash, Read
argument-hint: <command>
---

# Zellij Run

zellijの別ペインで任意のコマンドを実行します。

## 引数

$ARGUMENTS

## 実行手順

### 1. コマンドの検証

引数が空の場合はエラーを返してください。

### 2. 新しいペインで実行

```bash
# 下部に新しいペインを作成
zellij action new-pane --direction down --name "runner"

# コマンドを送信
zellij action write-chars "$ARGUMENTS"
zellij action write 10  # Enter key
```

### 3. メインペインに戻る

```bash
zellij action move-focus up
```

## 使い方

- `/zellij-run make build` - ビルドを別ペインで実行
- `/zellij-run npm run dev` - devサーバーを別ペインで起動
- `/zellij-run go run cmd/server/main.go` - サーバーを別ペインで起動

## ペイン操作

### キーバインド（zellijデフォルト）

| 操作 | キー |
|------|------|
| ペイン間移動 | `Alt + 矢印` |
| ペインを閉じる | `Ctrl + p, x` |
| ペインモード | `Ctrl + p` |
| フローティング | `Ctrl + p, w` |

### 注意事項

- 長時間実行されるプロセス（dev server等）は別ペインで起動するのに適しています
- テストは `/zellij-test` コマンドを使用してください
