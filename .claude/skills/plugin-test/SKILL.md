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
if jq -e . .claude-plugin/plugin.json > /dev/null 2>&1; then
    echo "✓ plugin.json is valid JSON"
else
    echo "✗ plugin.json is INVALID JSON"
    exit 1
fi

# Check required fields
MISSING_FIELDS=0
for field in name version description repository license; do
    if jq -e "has(\"$field\")" .claude-plugin/plugin.json > /dev/null 2>&1; then
        echo "✓ $field"
    else
        echo "✗ $field MISSING"
        MISSING_FIELDS=1
    fi
done

# Check author sub-fields
if jq -e '.author | type=="object" and has("name") and has("url")' .claude-plugin/plugin.json > /dev/null 2>&1; then
    echo "✓ author (with name and url)"
else
    echo "✗ author (with name and url) MISSING or incomplete"
    MISSING_FIELDS=1
fi

# Exit if any required fields are missing
if [ $MISSING_FIELDS -eq 1 ]; then
    exit 1
fi
```

### 3. Environment Check

**Requirements:**

- `curl` command installed
- `jq` command installed
- Gerrit environment variables set (`GERRIT_USER`, `GERRIT_HTTP_PASSWORD`)

**Validation:**

```bash
# Check curl command
which curl > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ curl command found"
    curl --version | head -1
else
    echo "✗ curl command NOT FOUND"
    exit 1
fi

# Check jq command
which jq > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✓ jq command found"
    jq --version
else
    echo "✗ jq command NOT FOUND"
    echo "Install jq for JSON processing: https://jqlang.github.io/jq/download/"
    exit 1
fi

# Check Gerrit environment variables
if [ -z "$GERRIT_USER" ]; then
    echo "✗ GERRIT_USER NOT SET"
    echo "Set GERRIT_USER environment variable"
    exit 1
fi

if [ -z "$GERRIT_HTTP_PASSWORD" ]; then
    echo "✗ GERRIT_HTTP_PASSWORD NOT SET"
    echo "Get HTTP password from: https://go-review.googlesource.com/settings/#HTTPCredentials"
    exit 1
fi

echo "✓ GERRIT_USER set"
echo "✓ GERRIT_HTTP_PASSWORD set"
```

### 4. Basic Functionality Test

#### Test: Fetch 1 Change from golang/go

```bash
# Helper function to fetch Gerrit API and strip XSSI prefix
gerrit_api() {
  local endpoint="$1"
  local base_url="${GERRIT_BASE_URL:-https://go-review.googlesource.com}"
  curl -sf -u "${GERRIT_USER}:${GERRIT_HTTP_PASSWORD}" \
    "${base_url}/a${endpoint}" | sed "1s/^)]}'//"
}

echo "Testing Change fetch..."
CHANGE_DATA=$(gerrit_api "/changes/?q=project:go+status:merged&n=1&o=DETAILED_ACCOUNTS" 2>&1)

if [ $? -eq 0 ] && echo "$CHANGE_DATA" | jq -e . > /dev/null 2>&1; then
    echo "✓ Change fetch successful"
    echo "$CHANGE_DATA" | jq . | head -20
else
    echo "✗ Change fetch FAILED"
    echo "$CHANGE_DATA"
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
- `2`: Critical error (missing curl or jq, Gerrit credentials not set)

## Usage

This skill is invoked by the `/test-plugin` command, which runs all tests in a separate zellij pane.
