---
description: Run plugin tests in a separate zellij pane
allowed-tools: Bash
---

Run comprehensive plugin tests in a separate pane.

## Test Details

### Test 1: File Structure Validation

**Purpose**: Validates that the plugin's basic structure exists correctly. For the plugin to function properly, all required files and directories must be present.

**Expected Result**: When all required files exist, a `✓` marker is displayed before each filename.

**Example Output**:

```text
1. File Structure Validation
✓ plugin.json
✓ go-pr-fetcher
✓ go-pr-analyzer
✓ go-catchup
✓ README.md
```

**Troubleshooting**:

- **If files are not found**:
  - If files were deleted, restore from git: `git checkout -- <file-path>`
  - If files were moved, restore to the correct path or check with git: `git status`
  - If new files need to be created, reference existing files as examples

**Why it matters**: Missing these files prevents the plugin from being recognized correctly by Claude Code and causes functionality to fail.

---

### Test 2: Plugin Manifest Validation

**Purpose**: Validates that `.claude-plugin/plugin.json` is valid JSON and contains required fields (name, version, description) necessary for plugin operation.

**Expected Result**: When JSON is valid and all required fields exist, a `✓` marker is displayed before each item.

**Example Output**:

```text
2. Plugin Manifest Validation
✓ Valid JSON
✓ name field
✓ version field
✓ description field
```

**Troubleshooting**:

- **If JSON is invalid**:
  - Check JSON syntax errors: `cat .claude-plugin/plugin.json | python3 -m json.tool`
  - Fix missing commas, mismatched quotes, missing closing brackets, etc.
  - Validate with an online JSON validator (e.g., jsonlint.com)
- **If required fields are missing**:
  - `name`: Add plugin name (e.g., `"name": "go-pr-insights"`)
  - `version`: Add version number (e.g., `"version": "1.0.0"`)
  - `description`: Add plugin description

**Why it matters**: Invalid manifest files prevent plugin installation and recognition. Claude Code reads this file to obtain plugin metadata.

---

### Test 3: Environment Check

**Purpose**: Validates that the required environment (GitHub CLI `gh` command, authentication status, and Python 3) is properly set up for the plugin to function.

**Expected Result**: When the `gh` command is installed, authenticated, and Python 3 is available, a `✓` marker is displayed before all items.

**Example Output**:

```text
3. Environment Check
✓ gh command found
✓ gh authenticated
✓ Python 3 available
```

**Troubleshooting**:

- **If `gh` command is not found**:
  - macOS: `brew install gh`
  - Linux: Install via package manager (e.g., `sudo apt install gh` or `sudo dnf install gh`)
  - Windows: Download installer from [GitHub CLI official site](https://cli.github.com/)
  - After installation, restart shell and verify with `which gh`
- **If not authenticated**:
  - Run `gh auth login`
  - Complete authentication in browser or enter token
  - Check authentication status: `gh auth status`
  - If authentication expired, re-authenticate: `gh auth refresh`
- **If Python 3 is not found**:
  - macOS: `brew install python3` or install from [python.org](https://www.python.org/downloads/)
  - Linux: Install via package manager (e.g., `sudo apt install python3` or `sudo dnf install python3`)
  - Windows: Download installer from [python.org](https://www.python.org/downloads/)
  - After installation, verify with `python3 --version` (Python 3.6+ required)
  - **Note**: Test scripts use `python3 -m json.tool` for JSON validation and formatting (used in Test 2 and Test 6)

**Why it matters**: The `gh` command and authentication are required to fetch PR information from the GitHub API. Python 3 is needed for JSON validation and formatting (`python3 -m json.tool`). Without these, the plugin's core functionality will not work.

---

### Test 4: Skill Definition Validation

**Purpose**: Validates that Skill files (`skills/*/SKILL.md`) have the correct format and contain required fields (`description`, `allowed-tools`).

**Expected Result**: When required fields exist in each Skill file, a `✓` marker is displayed before each item.

**Example Output**:

```text
4. Skill Definition Validation
✓ go-pr-fetcher description
✓ go-pr-fetcher allowed-tools
✓ go-pr-analyzer description
✓ go-pr-analyzer allowed-tools
```

**Troubleshooting**:

- **If fields are missing**:
  - Add frontmatter section at the beginning of the Skill file:
    ```yaml
    ---
    description: |
      Describe the skill here
    allowed-tools:
      - Bash
      - WebFetch
    ---
    ```
  - Verify YAML format is correct (indentation, space after colon)
  - Reference existing Skill files for format

**Why it matters**: If Skill file format is incorrect, Claude Code cannot recognize the skill and users cannot use it.

---

### Test 5: Command Definition Validation

**Purpose**: Validates that Command files (`commands/*.md`) have the correct format and contain required fields (`description`, `allowed-tools`).

**Expected Result**: When required fields exist in the Command file, a `✓` marker is displayed before each item.

**Example Output**:

```text
5. Command Definition Validation
✓ go-catchup description
✓ go-catchup allowed-tools
```

**Troubleshooting**:

- **If fields are missing**:
  - Add frontmatter section at the beginning of the Command file:
    ```yaml
    ---
    description: Describe the command here
    allowed-tools: Bash
    ---
    ```
  - Verify YAML format is correct
  - Reference existing Command files for format

**Why it matters**: If Command file format is incorrect, Claude Code cannot recognize the command and users cannot execute it.

---

### Test 6: Basic Functionality Test

**Purpose**: Validates that GitHub API access works correctly and can actually fetch PR information. This verifies that the plugin's core functionality works.

**Expected Result**: When API access succeeds, PR information fetched from the golang/go repository is displayed in JSON format.

**Example Output**:

```text
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

**Troubleshooting**:

- **If authentication error occurs**:
  - Check authentication status: `gh auth status`
  - If authentication is invalid, re-run `gh auth login`
  - Check token expiration: `gh auth status`
- **If rate limit error occurs**:
  - GitHub API rate limit may have been reached
  - Wait a while and retry
  - Authenticated users have higher rate limits: `gh auth login`
- **If network error occurs**:
  - Check internet connection
  - Check proxy settings: `gh api --hostname github.com`
  - Check GitHub status: [www.githubstatus.com](https://www.githubstatus.com/)
- **If repository access error occurs**:
  - Verify access permissions to golang/go repository (usually not an issue as it's a public repository)
  - Verify repository name is correct

**Why it matters**: This test verifies that the plugin's core functionality (fetching PR information) actually works. If API access fails, the plugin cannot achieve its purpose.

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
zellij action write-chars "python3 --version > /dev/null 2>&1 && echo '✓ Python 3 available' || echo '✗ Python 3 NOT FOUND'"
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

# 4. Return to main pane
zellij action move-focus left
```

## Completion

### Result Verification

Tests run in the right pane. After tests complete, follow these steps to verify results:

1. **Check the right pane**: Test results are displayed in zellij's right pane (`plugin-test`)
2. **Check each test section**: Verify results for each of the 6 test categories
3. **Check ✓ markers**: Verify all items are displayed with `✓` markers

### Success Criteria

**When all tests succeed**:

- All test items display `✓` markers
- Test 6 displays PR information in JSON format correctly
- No error messages (`✗` markers) are displayed

**Example output on success**:

```text
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
✓ Python 3 available

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
[Valid JSON output]

=== Test Complete ===
Review results above. All ✓ means plugin is ready for distribution.
```

### Next Steps on Test Failure

**If some tests fail**:

1. **Identify failed items**: Check items displaying `✗` markers
2. **Refer to relevant troubleshooting section**: Check troubleshooting steps for failed tests in the "Test Details" section above
3. **Fix the issue**: Resolve problems following troubleshooting steps
4. **Re-test**: After fixes, run tests again to verify

**Common failure patterns and solutions**:

- **Files not found**: Restore files from git or place in correct path
- **JSON error**: Fix syntax errors in `plugin.json`
- **gh command not found**: Install GitHub CLI
- **Authentication error**: Run `gh auth login` to re-authenticate
- **Python 3 not found**: Install Python 3 (verify with `python3 --version`)
- **API error**: Check network connection and GitHub status

**When all tests succeed**:
The plugin is ready for distribution. The plugin is ready to be packaged and distributed.

### Report Message

Report: "Plugin tests are running in the right pane. Check results there. All ✓ markers indicate success. If any ✗ markers appear, refer to the troubleshooting section for each test."

## Test Coverage

- File structure validation
- Plugin manifest (plugin.json) validation
- Environment requirements (gh command, authentication, Python 3)
- Skill definition format validation
- Command definition format validation
- Basic API functionality (fetch 1 PR from golang/go)
