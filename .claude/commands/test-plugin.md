---
description: Run plugin tests in a separate zellij pane
allowed-tools: Bash
---

Run comprehensive plugin tests in a separate pane.

## Test Details

### Test 1: File Structure Validation

**Purpose（目的）**: プラグインの基本構造が正しく存在することを確認します。プラグインが正常に動作するためには、必要なファイルとディレクトリがすべて存在している必要があります。

**Expected Result（期待される結果）**: すべての必須ファイルが存在する場合、各ファイル名の前に `✓` マーカーが表示されます。

**Example Output（サンプル出力）**:
```
1. File Structure Validation
✓ plugin.json
✓ go-pr-fetcher
✓ go-pr-analyzer
✓ go-catchup
✓ README.md
```

**Troubleshooting（トラブルシューティング）**:
- **ファイルが見つからない場合**: 
  - ファイルが削除された場合は、gitから復元: `git checkout -- <file-path>`
  - ファイルが移動された場合は、正しいパスに戻すか、gitで確認: `git status`
  - 新規作成が必要な場合は、既存のファイルを参考に作成

**Why it matters（重要性）**: これらのファイルが欠落していると、プラグインがClaude Codeで正しく認識されず、機能が動作しません。

---

### Test 2: Plugin Manifest Validation

**Purpose（目的）**: `.claude-plugin/plugin.json` が有効なJSON形式で、プラグインの動作に必要な必須フィールド（name、version、description）を含むことを確認します。

**Expected Result（期待される結果）**: JSONが有効で、必須フィールドがすべて存在する場合、各項目の前に `✓` マーカーが表示されます。

**Example Output（サンプル出力）**:
```
2. Plugin Manifest Validation
✓ Valid JSON
✓ name field
✓ version field
✓ description field
```

**Troubleshooting（トラブルシューティング）**:
- **無効なJSONの場合**: 
  - JSON構文エラーを確認: `cat .claude-plugin/plugin.json | python3 -m json.tool`
  - カンマの欠落、引用符の不一致、閉じ括弧の欠落などを修正
  - オンラインJSONバリデーター（例: jsonlint.com）で検証
- **必須フィールドが欠落している場合**: 
  - `name`: プラグイン名を追加（例: `"name": "go-pr-insights"`）
  - `version`: バージョン番号を追加（例: `"version": "1.0.0"`）
  - `description`: プラグインの説明を追加

**Why it matters（重要性）**: 無効なマニフェストファイルは、プラグインのインストールや認識を妨げます。Claude Codeはこのファイルを読み込んでプラグインのメタデータを取得します。

---

### Test 3: Environment Check

**Purpose（目的）**: プラグインが動作するために必要な環境（GitHub CLI `gh` コマンドと認証状態）が整っていることを確認します。

**Expected Result（期待される結果）**: `gh` コマンドがインストールされており、認証済みの場合、両方の項目の前に `✓` マーカーが表示されます。

**Example Output（サンプル出力）**:
```
3. Environment Check
✓ gh command found
✓ gh authenticated
```

**Troubleshooting（トラブルシューティング）**:
- **`gh` コマンドが見つからない場合**: 
  - macOS: `brew install gh`
  - Linux: パッケージマネージャーでインストール（例: `sudo apt install gh` または `sudo dnf install gh`）
  - Windows: [GitHub CLI公式サイト](https://cli.github.com/)からインストーラーをダウンロード
  - インストール後、シェルを再起動して `which gh` で確認
- **認証されていない場合**: 
  - `gh auth login` を実行
  - ブラウザで認証を完了するか、トークンを入力
  - 認証状態を確認: `gh auth status`
  - 認証が期限切れの場合は再認証: `gh auth refresh`

**Why it matters（重要性）**: `gh` コマンドと認証は、GitHub APIからPR情報を取得するために必須です。これらが整っていないと、プラグインの主要機能が動作しません。

---

### Test 4: Skill Definition Validation

**Purpose（目的）**: Skillファイル（`skills/*/SKILL.md`）が正しいフォーマットを持ち、必須フィールド（`description`、`allowed-tools`）が含まれていることを確認します。

**Expected Result（期待される結果）**: 各Skillファイルの必須フィールドが存在する場合、各項目の前に `✓` マーカーが表示されます。

**Example Output（サンプル出力）**:
```
4. Skill Definition Validation
✓ go-pr-fetcher description
✓ go-pr-fetcher allowed-tools
✓ go-pr-analyzer description
✓ go-pr-analyzer allowed-tools
```

**Troubleshooting（トラブルシューティング）**:
- **フィールドが欠落している場合**: 
  - Skillファイルの先頭にfrontmatterセクションを追加:
    ```yaml
    ---
    description: |
      Skillの説明をここに記載
    allowed-tools:
      - Bash
      - WebFetch
    ---
    ```
  - YAMLフォーマットが正しいか確認（インデント、コロンの後にスペース）
  - 既存のSkillファイルを参考にフォーマットを確認

**Why it matters（重要性）**: Skillファイルのフォーマットが正しくないと、Claude CodeがSkillを認識できず、ユーザーがSkillを使用できません。

---

### Test 5: Command Definition Validation

**Purpose（目的）**: Commandファイル（`commands/*.md`）が正しいフォーマットを持ち、必須フィールド（`description`、`allowed-tools`）が含まれていることを確認します。

**Expected Result（期待される結果）**: Commandファイルの必須フィールドが存在する場合、各項目の前に `✓` マーカーが表示されます。

**Example Output（サンプル出力）**:
```
5. Command Definition Validation
✓ go-catchup description
✓ go-catchup allowed-tools
```

**Troubleshooting（トラブルシューティング）**:
- **フィールドが欠落している場合**: 
  - Commandファイルの先頭にfrontmatterセクションを追加:
    ```yaml
    ---
    description: Commandの説明をここに記載
    allowed-tools: Bash
    ---
    ```
  - YAMLフォーマットが正しいか確認
  - 既存のCommandファイルを参考にフォーマットを確認

**Why it matters（重要性）**: Commandファイルのフォーマットが正しくないと、Claude Codeがコマンドを認識できず、ユーザーがコマンドを実行できません。

---

### Test 6: Basic Functionality Test

**Purpose（目的）**: GitHub APIへのアクセスが正常に動作し、実際にPR情報を取得できることを確認します。これにより、プラグインの主要機能が動作することを検証します。

**Expected Result（期待される結果）**: APIアクセスが成功した場合、golang/goリポジトリから取得したPR情報がJSON形式で表示されます。

**Example Output（サンプル出力）**:
```
6. Basic Functionality Test
Fetching 1 PR from golang/go...
[
  {
    "number": 12345,
    "title": "example: improve error handling",
    "author": {
      "login": "example-user"
    }
  }
]
```

**Troubleshooting（トラブルシューティング）**:
- **認証エラーの場合**: 
  - `gh auth status` で認証状態を確認
  - 認証が無効な場合は `gh auth login` を再実行
  - トークンの有効期限を確認: `gh auth status`
- **レート制限エラーの場合**: 
  - GitHub APIのレート制限に達している可能性があります
  - しばらく待ってから再試行
  - 認証済みユーザーはより高いレート制限があります: `gh auth login`
- **ネットワークエラーの場合**: 
  - インターネット接続を確認
  - プロキシ設定を確認: `gh api --hostname github.com`
  - GitHubのステータスを確認: [status.github.com](https://www.githubstatus.com/)
- **リポジトリアクセスエラーの場合**: 
  - golang/goリポジトリへのアクセス権限を確認（通常は公開リポジトリなので問題ありません）
  - リポジトリ名が正しいか確認

**Why it matters（重要性）**: このテストは、プラグインの主要機能（PR情報の取得）が実際に動作することを確認します。APIアクセスが失敗すると、プラグインの目的を達成できません。

---

## Steps

1. Check zellij session
2. Use existing right pane or create new one
3. Run plugin test suite
4. Return to main pane

## Execution

```bash
# 1. Check zellij session
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: Must run inside a zellij session"
  exit 1
fi

# 2. Use existing pane or create new one
zellij action new-pane --direction right --name "plugin-test" 2>/dev/null || zellij action move-focus right
zellij action write-chars "clear"
zellij action write 10

# 3. Run test suite
zellij action write-chars "echo '=== Upgo Plugin Test Suite ==='"
zellij action write 10
zellij action write-chars "echo ''"
zellij action write 10

# Test 1: File Structure
zellij action write-chars "echo '1. File Structure Validation'"
zellij action write 10
zellij action write-chars "test -f .claude-plugin/plugin.json && echo '✓ plugin.json' || echo '✗ plugin.json MISSING'"
zellij action write 10
zellij action write-chars "test -f skills/go-pr-fetcher/SKILL.md && echo '✓ go-pr-fetcher' || echo '✗ go-pr-fetcher MISSING'"
zellij action write 10
zellij action write-chars "test -f skills/go-pr-analyzer/SKILL.md && echo '✓ go-pr-analyzer' || echo '✗ go-pr-analyzer MISSING'"
zellij action write 10
zellij action write-chars "test -f commands/go-catchup.md && echo '✓ go-catchup' || echo '✗ go-catchup MISSING'"
zellij action write 10
zellij action write-chars "test -f README.md && echo '✓ README.md' || echo '✗ README.md MISSING'"
zellij action write 10
zellij action write-chars "echo ''"
zellij action write 10

# Test 2: Plugin Manifest
zellij action write-chars "echo '2. Plugin Manifest Validation'"
zellij action write 10
zellij action write-chars "cat .claude-plugin/plugin.json | python3 -m json.tool > /dev/null 2>&1 && echo '✓ Valid JSON' || echo '✗ INVALID JSON'"
zellij action write 10
zellij action write-chars "grep -q '\"name\"' .claude-plugin/plugin.json && echo '✓ name field' || echo '✗ name MISSING'"
zellij action write 10
zellij action write-chars "grep -q '\"version\"' .claude-plugin/plugin.json && echo '✓ version field' || echo '✗ version MISSING'"
zellij action write 10
zellij action write-chars "grep -q '\"description\"' .claude-plugin/plugin.json && echo '✓ description field' || echo '✗ description MISSING'"
zellij action write 10
zellij action write-chars "echo ''"
zellij action write 10

# Test 3: Environment Check
zellij action write-chars "echo '3. Environment Check'"
zellij action write 10
zellij action write-chars "which gh > /dev/null 2>&1 && echo '✓ gh command found' || echo '✗ gh NOT FOUND'"
zellij action write 10
zellij action write-chars "gh auth status > /dev/null 2>&1 && echo '✓ gh authenticated' || echo '✗ gh NOT authenticated'"
zellij action write 10
zellij action write-chars "echo ''"
zellij action write 10

# Test 4: Skill Definitions
zellij action write-chars "echo '4. Skill Definition Validation'"
zellij action write 10
zellij action write-chars "grep -q '^description:' skills/go-pr-fetcher/SKILL.md && echo '✓ go-pr-fetcher description' || echo '✗ Missing description'"
zellij action write 10
zellij action write-chars "grep -q '^allowed-tools:' skills/go-pr-fetcher/SKILL.md && echo '✓ go-pr-fetcher allowed-tools' || echo '✗ Missing allowed-tools'"
zellij action write 10
zellij action write-chars "grep -q '^description:' skills/go-pr-analyzer/SKILL.md && echo '✓ go-pr-analyzer description' || echo '✗ Missing description'"
zellij action write 10
zellij action write-chars "grep -q '^allowed-tools:' skills/go-pr-analyzer/SKILL.md && echo '✓ go-pr-analyzer allowed-tools' || echo '✗ Missing allowed-tools'"
zellij action write 10
zellij action write-chars "echo ''"
zellij action write 10

# Test 5: Command Definition
zellij action write-chars "echo '5. Command Definition Validation'"
zellij action write 10
zellij action write-chars "grep -q '^description:' commands/go-catchup.md && echo '✓ go-catchup description' || echo '✗ Missing description'"
zellij action write 10
zellij action write-chars "grep -q '^allowed-tools:' commands/go-catchup.md && echo '✓ go-catchup allowed-tools' || echo '✗ Missing allowed-tools'"
zellij action write 10
zellij action write-chars "echo ''"
zellij action write 10

# Test 6: Basic Functionality
zellij action write-chars "echo '6. Basic Functionality Test'"
zellij action write 10
zellij action write-chars "echo 'Fetching 1 PR from golang/go...'"
zellij action write 10
zellij action write-chars "gh pr list --repo golang/go --state merged --limit 1 --json number,title,author 2>&1 | python3 -m json.tool"
zellij action write 10
zellij action write-chars "echo ''"
zellij action write 10

# Summary
zellij action write-chars "echo '=== Test Complete ==='"
zellij action write 10
zellij action write-chars "echo 'Review results above. All ✓ means plugin is ready for distribution.'"
zellij action write 10

# 4. Return to main
zellij action move-focus left
```

## Completion

### 結果の確認

テストは右側のペインで実行されます。テストが完了したら、以下の手順で結果を確認してください：

1. **右側のペインを確認**: zellijの右側ペイン（`plugin-test`）にテスト結果が表示されます
2. **各テストセクションを確認**: 6つのテストカテゴリそれぞれの結果を確認します
3. **✓マーカーを確認**: すべての項目が `✓` で表示されているか確認します

### 成功の判断基準

**すべてのテストが成功した場合**:
- すべてのテスト項目の前に `✓` マーカーが表示されている
- Test 6でPR情報がJSON形式で正常に表示されている
- エラーメッセージ（`✗` マーカー）が一切表示されていない

**成功時の表示例**:
```
=== Upgo Plugin Test Suite ===

1. File Structure Validation
✓ plugin.json
✓ go-pr-fetcher
✓ go-pr-analyzer
✓ go-catchup
✓ README.md

2. Plugin Manifest Validation
✓ Valid JSON
✓ name field
✓ version field
✓ description field

3. Environment Check
✓ gh command found
✓ gh authenticated

4. Skill Definition Validation
✓ go-pr-fetcher description
✓ go-pr-fetcher allowed-tools
✓ go-pr-analyzer description
✓ go-pr-analyzer allowed-tools

5. Command Definition Validation
✓ go-catchup description
✓ go-catchup allowed-tools

6. Basic Functionality Test
Fetching 1 PR from golang/go...
[正常なJSON出力]

=== Test Complete ===
Review results above. All ✓ means plugin is ready for distribution.
```

### テスト失敗時の次のステップ

**一部のテストが失敗した場合**:

1. **失敗した項目を特定**: `✗` マーカーが表示されている項目を確認
2. **該当するトラブルシューティングセクションを参照**: 上記の「Test Details」セクションで、失敗したテストのトラブルシューティング手順を確認
3. **問題を修正**: トラブルシューティング手順に従って問題を解決
4. **再テスト**: 修正後、再度テストを実行して確認

**よくある失敗パターンと対処**:
- **ファイルが見つからない**: gitでファイルを復元するか、正しいパスに配置
- **JSONエラー**: `plugin.json` の構文エラーを修正
- **ghコマンドが見つからない**: GitHub CLIをインストール
- **認証エラー**: `gh auth login` を実行して再認証
- **APIエラー**: ネットワーク接続とGitHubのステータスを確認

**すべてのテストが成功したら**:
プラグインは配布可能な状態です。プラグインをパッケージ化して配布する準備が整いました。

### Report Message

Report: "Plugin tests are running in the right pane. Check results there. All ✓ markers indicate success. If any ✗ markers appear, refer to the troubleshooting section for each test."

## Test Coverage

- File structure validation
- Plugin manifest (plugin.json) validation
- Environment requirements (gh command, authentication)
- Skill definition format validation
- Command definition format validation
- Basic API functionality (fetch 1 PR from golang/go)
