---
description: zellijの別ペインでテストを実行し、結果を監視
allowed-tools: Bash, Read
argument-hint: [test-pattern]
---

# Zellij Test Runner

zellijの別ペインでテストを実行します。

## 引数

$ARGUMENTS

## 実行手順

### 1. 新しいペインでテストを起動

```bash
# テストコマンドを構築
TEST_CMD="go test -v ./..."
if [ -n "$ARGUMENTS" ]; then
    TEST_CMD="go test -v -run '$ARGUMENTS' ./..."
fi

# 右側に新しいペインを作成してテスト実行
zellij action new-pane --direction right --name "test-runner"
zellij action write-chars "$TEST_CMD"
zellij action write 10  # Enter key
```

### 2. メインペインに戻る

```bash
zellij action move-focus left
```

### 3. テスト結果の確認

テストが完了したら、ユーザーに結果を報告してください。
以下のコマンドで別ペインの出力を確認できます：

```bash
# ペインの出力をスクロールバッファから取得
zellij action toggle-pane-frames
```

## 使い方

- `/zellij-test` - 全テストを別ペインで実行
- `/zellij-test TestUserService` - 特定のテストのみ実行
- `/zellij-test Integration` - 名前にIntegrationを含むテストを実行

## 注意事項

- zellijセッション内で実行する必要があります
- テスト実行中もメインペインで作業を継続できます
- テストの結果は別ペインに表示されます
