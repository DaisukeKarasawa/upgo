---
description: タスクを分析し、zellijの複数ペインで並列実行
allowed-tools: Bash, Read, Write
argument-hint: <task-description>
---

タスクを分析し、並列実行可能なサブタスクに分割して複数ペインで実行します。

## タスク

$ARGUMENTS

## 手順

### 1. タスク分析

ユーザーのタスクを分析し、2-4個の独立したサブタスクに分割してください。

例:
- 「APIを実装してテストも書いて」→ タスク1: API実装、タスク2: テスト作成
- 「コードレビューして」→ タスク1: セキュリティ、タスク2: パフォーマンス、タスク3: スタイル

### 2. ペインレイアウト

サブタスク数に応じてレイアウトを選択:

**2タスク (左右):**
```bash
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
```

**3タスク (T字):**
```bash
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
zellij action new-pane --direction down --name "task-3"
zellij action move-focus up
```

**4タスク (グリッド):**
```bash
zellij action new-pane --direction right --name "task-2"
zellij action move-focus left
zellij action new-pane --direction down --name "task-3"
zellij action move-focus right
zellij action new-pane --direction down --name "task-4"
zellij action move-focus up
zellij action move-focus left
```

### 3. エージェント起動

各ペインでClaude Codeを起動してタスクを割り当て:

```bash
# ペイン2にタスク送信
zellij action move-focus right
zellij action write-chars "claude --print 'サブタスク2の内容'"
zellij action write 10
zellij action move-focus left
```

### 4. メインタスク実行

メインペインでタスク1を直接実行してください。

## 完了報告

「タスクを N 個のサブタスクに分割し、各ペインで並列実行を開始しました。」と報告してください。
