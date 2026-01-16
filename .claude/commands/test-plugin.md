---
description: Run plugin tests in a separate zellij pane
allowed-tools: Bash
---

Run comprehensive plugin tests in a separate pane.

## Steps

1. Check zellij session
2. Create new pane on the right
3. Run plugin test suite
4. Return to main pane

## Execution

```bash
# 1. Check zellij session
if [ -z "$ZELLIJ" ]; then
  echo "ERROR: Must run inside a zellij session"
  exit 1
fi

# 2. Create test pane
zellij action new-pane --direction right --name "plugin-test"

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

Report: "Plugin tests are running in the right pane. Check results there."

## Test Coverage

- File structure validation
- Plugin manifest (plugin.json) validation
- Environment requirements (gh command, authentication)
- Skill definition format validation
- Command definition format validation
- Basic API functionality (fetch 1 PR from golang/go)
