---
description: 別ペインでClaude Codeペアプログラミングパートナーを起動
allowed-tools: Bash
argument-hint: [initial-prompt]
---

別ペインでClaude Codeを起動し、ペアプログラミングパートナーとして動作させます。

## 初期プロンプト

$ARGUMENTS

## 実行

```bash
# zellijセッション確認
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: zellijセッション内で実行してください"
  exit 1
fi

# 右側にペイン作成
zellij action new-pane --direction right --name "pair-programmer"

# Claude Codeを起動
INITIAL_PROMPT="$ARGUMENTS"
if [ -n "$INITIAL_PROMPT" ]; then
  zellij action write-chars "claude --print 'ペアプログラミングパートナーとして支援してください。$INITIAL_PROMPT'"
else
  zellij action write-chars "claude --print 'ペアプログラミングパートナーとして支援してください。'"
fi
zellij action write 10

# メインに戻る
zellij action move-focus left
```

## 使用パターン

### テスト担当

```bash
/loop テストの作成を担当してください
```

### レビュー担当

```bash
/loop コードレビューを担当してください
```

### ドキュメント担当

```bash
/loop ドキュメントの作成を担当してください
```

## 完了報告

「右側のペインでペアプログラミングパートナーが起動しました。」と報告してください。
