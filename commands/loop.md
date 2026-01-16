---
description: 別ペインでClaude Codeをループモードで起動しペアプログラミング
allowed-tools: Bash
argument-hint: [initial-prompt]
---

# Loop - Pair Programming Partner

別ペインでClaude Codeをループモードで起動し、継続的なペアプログラミングを行います。

## 引数

$ARGUMENTS

## 概要

Claude Codeの `--print` と `--allowedTools` オプションを使って、特定のタスクに特化したエージェントを別ペインで起動します。

## 実行手順

### 1. 新しいペインを作成

```bash
zellij action new-pane --direction right --name "pair-programmer"
```

### 2. Claude Codeを起動

引数に応じて適切な設定で起動:

**テスト特化:**
```bash
zellij action write-chars "claude --print 'テストの作成と実行に集中してください。TDDサイクルを回します。' --allowedTools 'Bash,Read,Write,Edit'"
zellij action write 10
```

**レビュー特化:**
```bash
zellij action write-chars "claude --print 'コードレビューに集中してください。品質とセキュリティを確認します。' --allowedTools 'Read,Grep,Glob'"
zellij action write 10
```

**汎用:**
```bash
zellij action write-chars "claude --print 'ペアプログラミングパートナーとして支援します。$ARGUMENTS'"
zellij action write 10
```

### 3. メインペインに戻る

```bash
zellij action move-focus left
```

## 使い方

- `/loop` - 汎用ペアプログラミングパートナー
- `/loop テストを書いて` - テスト作成に特化
- `/loop このコードをレビューして` - レビューに特化
- `/loop ドキュメントを更新して` - ドキュメント作成に特化

## ワークフロー

```
┌─────────────────┬─────────────────┐
│                 │                 │
│   Main Claude   │  Loop Claude    │
│   (メイン作業)   │  (特化タスク)    │
│                 │                 │
└─────────────────┴─────────────────┘

1. メインペインで機能を実装
2. 右ペインのClaude Codeにテスト作成を依頼
3. 両方のClaude Codeが並列で作業
4. 結果を統合
```

## 通信パターン

### ファイル経由の情報共有

```bash
# メインペインで実装を完了
# → ファイルに保存

# 右ペインのClaude Codeがファイルを読んでテスト作成
```

### タスク完了の通知

```bash
# 共有ファイルでステータスを管理
echo "DONE: feature-implementation" >> .task-status
```

## 注意事項

- 各Claude Codeセッションは独立したコンテキストを持ちます
- API使用量は各セッション分かかります
- 長時間の並列作業は適切に管理してください
