---
description: zellijの別ペインでテストを実行
allowed-tools: Bash
argument-hint: [test-pattern]
---

別ペインでテストを実行します。

## 手順

1. zellijセッション確認
2. 右側に新しいペインを作成
3. テストを実行
4. メインペインに戻る

## 実行

```bash
# 1. zellijセッション確認
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: zellijセッション内で実行してください"
  exit 1
fi

# 2. テストペイン作成
zellij action new-pane --direction right --name "test-runner"

# 3. テスト実行
TEST_PATTERN="$ARGUMENTS"
if [ -n "$TEST_PATTERN" ]; then
  zellij action write-chars "go test -v -run '$TEST_PATTERN' ./..."
else
  zellij action write-chars "go test -v ./..."
fi
zellij action write 10

# 4. メインに戻る
zellij action move-focus left
```

## 完了報告

「右側のペインでテストが実行されています。結果はそちらで確認できます。」と報告してください。
