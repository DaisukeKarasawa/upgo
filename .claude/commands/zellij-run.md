---
description: zellijの別ペインでコマンドを実行
allowed-tools: Bash
argument-hint: <command>
---

別ペインで任意のコマンドを実行します。

## 引数

$ARGUMENTS

## 実行

```bash
# zellijセッション確認
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: zellijセッション内で実行してください"
  exit 1
fi

# 引数確認
if [ -z "$ARGUMENTS" ]; then
  echo "ERROR: 実行するコマンドを指定してください"
  exit 1
fi

# 下側にペイン作成
zellij action new-pane --direction down --name "runner"

# コマンド実行
zellij action write-chars "$ARGUMENTS"
zellij action write 10

# メインに戻る
zellij action move-focus up
```

## 完了報告

「下側のペインでコマンドが実行されています。」と報告してください。
