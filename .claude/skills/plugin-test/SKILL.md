---
description: Test upgo plugin functionality and structure
allowed-tools: Bash, Read, Glob, Grep, WebFetch
---

# Plugin Test Suite

Validates the upgo plugin structure, configuration, and basic functionality.

## Test Categories

### 1. File Structure Validation

**Required Files:**
- `.claude-plugin/plugin.json`
- `skills/go-pr-fetcher/SKILL.md`
- `skills/go-pr-analyzer/SKILL.md`
- `commands/go-catchup.md`
- `README.md`

**Validation:**
```bash
# Check all required files exist
test -f .claude-plugin/plugin.json && echo "✓ plugin.json" || echo "✗ plugin.json MISSING"
test -f skills/go-pr-fetcher/SKILL.md && echo "✓ go-pr-fetcher" || echo "✗ go-pr-fetcher MISSING"
test -f skills/go-pr-analyzer/SKILL.md && echo "✓ go-pr-analyzer" || echo "✗ go-pr-analyzer MISSING"
test -f commands/go-catchup.md && echo "✓ go-catchup" || echo "✗ go-catchup MISSING"
test -f README.md && echo "✓ README.md" || echo "✗ README.md MISSING"
```

### 2. Plugin Manifest Validation

**Required Fields in plugin.json:**
- `name`
- `version`
- `description`
- `author` (with `name` and `url`)
- `repository`
- `license`

**Validation:**
```bash
# Read and validate plugin.json
cat .claude-plugin/plugin.json | python3 -m json.tool > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ plugin.json is valid JSON"
else
    echo "✗ plugin.json is INVALID JSON"
fi

# Check required fields
for field in name version description author repository license; do
    grep -q "\"$field\"" .claude-plugin/plugin.json && echo "✓ $field" || echo "✗ $field MISSING"
done
```

### 3. Environment Check

**Requirements:**
- GitHub CLI (`gh`) installed
- `gh` authenticated

**Validation:**
```bash
# Check gh command
which gh > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ gh command found"
    gh --version
else
    echo "✗ gh command NOT FOUND"
    exit 1
fi

# Check authentication
gh auth status > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ gh authenticated"
else
    echo "✗ gh NOT authenticated"
    echo "Run: gh auth login"
    exit 1
fi
```

### 4. Basic Functionality Test

**Test: Fetch 1 PR from golang/go**

```bash
echo "Testing PR fetch..."
PR_DATA=$(gh pr list --repo golang/go --state merged --limit 1 --json number,title,author 2>&1)

if [ $? -eq 0 ]; then
    echo "✓ PR fetch successful"
    echo "$PR_DATA" | python3 -m json.tool
else
    echo "✗ PR fetch FAILED"
    echo "$PR_DATA"
    exit 1
fi
```

### 5. Skill Definition Validation

**Check SKILL.md format:**

```bash
# Check go-pr-fetcher skill
if grep -q "^description:" skills/go-pr-fetcher/SKILL.md; then
    echo "✓ go-pr-fetcher has description"
else
    echo "✗ go-pr-fetcher missing description"
fi

if grep -q "^allowed-tools:" skills/go-pr-fetcher/SKILL.md; then
    echo "✓ go-pr-fetcher has allowed-tools"
else
    echo "✗ go-pr-fetcher missing allowed-tools"
fi

# Check go-pr-analyzer skill
if grep -q "^description:" skills/go-pr-analyzer/SKILL.md; then
    echo "✓ go-pr-analyzer has description"
else
    echo "✗ go-pr-analyzer missing description"
fi

if grep -q "^allowed-tools:" skills/go-pr-analyzer/SKILL.md; then
    echo "✓ go-pr-analyzer has allowed-tools"
else
    echo "✗ go-pr-analyzer missing allowed-tools"
fi
```

### 6. Command Definition Validation

**Check command format:**

```bash
# Check go-catchup command
if grep -q "^description:" commands/go-catchup.md; then
    echo "✓ go-catchup has description"
else
    echo "✗ go-catchup missing description"
fi

if grep -q "^allowed-tools:" commands/go-catchup.md; then
    echo "✓ go-catchup has allowed-tools"
else
    echo "✗ go-catchup missing allowed-tools"
fi
```

## Test Execution Flow

Run all tests in sequence:

1. File structure validation
2. Plugin manifest validation
3. Environment check
4. Skill definition validation
5. Command definition validation
6. Basic functionality test

Generate test report showing:
- ✓ Passed tests
- ✗ Failed tests
- Summary statistics

## Exit Codes

- `0`: All tests passed
- `1`: One or more tests failed
- `2`: Critical error (missing gh, not authenticated)

## Usage

This skill is invoked by the `/test-plugin` command, which runs all tests in a separate zellij pane.
