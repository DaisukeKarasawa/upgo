---
name: zellij-workflow
description: |
  zellijターミナルマルチプレクサを使った並列開発ワークフロースキル。
  ユーザーが「別ペインで実行」「並列で」「zellij」「マルチペイン」「テストを別窓で」などと言った場合に使用。
  テスト実行、サーバー起動、マルチエージェント構成をサポート。
allowed-tools:
  - Bash
  - Read
---

# Zellij Workflow for Claude Code

zellijを使ってClaude Codeの作業を並列化するためのスキルです。

## 重要: zellijセッション確認

まず、zellijセッション内かどうかを確認してください：

```bash
echo $ZELLIJ
```

空の場合は `zellij` で新しいセッションを開始するよう促してください。

## コマンドリファレンス

### 新しいペイン作成

```bash
# 右側に新しいペイン
zellij action new-pane --direction right --name "ペイン名"

# 下側に新しいペイン
zellij action new-pane --direction down --name "ペイン名"

# フローティングペイン（ポップアップ）
zellij action new-pane --floating --name "ペイン名"
```

### コマンド送信

```bash
# 文字を送信
zellij action write-chars "コマンド文字列"

# Enterキーを送信（実行）
zellij action write 10
```

### フォーカス移動

```bash
zellij action move-focus left
zellij action move-focus right
zellij action move-focus up
zellij action move-focus down
```

## よく使うパターン

### パターン1: 別ペインでテスト実行

```bash
# 1. 右側にテストペインを作成
zellij action new-pane --direction right --name "tests"

# 2. テストコマンドを送信
zellij action write-chars "go test -v ./..."
zellij action write 10

# 3. メインペインに戻る
zellij action move-focus left
```

### パターン2: 別ペインでサーバー起動

```bash
# 1. 下側にサーバーペインを作成
zellij action new-pane --direction down --name "server"

# 2. サーバーを起動
zellij action write-chars "go run cmd/server/main.go"
zellij action write 10

# 3. メインペインに戻る
zellij action move-focus up
```

### パターン3: 別ペインでClaude Codeエージェント起動

```bash
# 1. 右側にエージェントペインを作成
zellij action new-pane --direction right --name "agent"

# 2. Claude Codeをタスク付きで起動
zellij action write-chars "claude --print 'タスクの指示'"
zellij action write 10

# 3. メインペインに戻る
zellij action move-focus left
```

### パターン4: 2x2グリッドで並列タスク

```bash
# グリッド作成
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
zellij action new-pane --direction down --name "task-3"
zellij action move-focus right
zellij action new-pane --direction down --name "task-4"

# メインに戻る
zellij action move-focus up
zellij action move-focus left
```

## 実行例

ユーザーが「テストを別ペインで実行して」と言った場合：

```bash
# zellijセッション確認
echo $ZELLIJ

# 別ペインでテスト
zellij action new-pane --direction right --name "test-runner"
zellij action write-chars "go test -v ./..."
zellij action write 10
zellij action move-focus left
```

ユーザーに「右側のペインでテストが実行されています」と報告してください。

## 注意事項

1. **Enter送信を忘れない**: `write-chars` だけではコマンドは実行されません。必ず `write 10` でEnterを送信。
2. **フォーカスを戻す**: 作業後はメインペインにフォーカスを戻す。
3. **ペイン名を付ける**: `--name` で識別しやすい名前を付ける。
