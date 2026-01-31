---
description: Run plugin tests in a separate zellij pane
allowed-tools: Bash
---

Run comprehensive plugin tests in a separate pane.

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
zellij action write-chars "which curl > /dev/null 2>&1 && echo '✓ curl command found' || echo '✗ curl NOT FOUND'"
zellij action write 10
zellij action write-chars "[ -n \"\$GERRIT_USER\" ] && echo '✓ GERRIT_USER set' || echo '✗ GERRIT_USER NOT SET'"
zellij action write 10
zellij action write-chars "[ -n \"\$GERRIT_HTTP_PASSWORD\" ] && echo '✓ GERRIT_HTTP_PASSWORD set' || echo '✗ GERRIT_HTTP_PASSWORD NOT SET'"
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
zellij action write-chars "echo 'Fetching 1 Change from golang/go...'"
zellij action write 10
zellij action write-chars "gerrit_api() { local e=\"\$1\"; local b=\"\${GERRIT_BASE_URL:-https://go-review.googlesource.com}\"; curl -sf -u \"\${GERRIT_USER}:\${GERRIT_HTTP_PASSWORD}\" \"\${b}/a\${e}\" | sed '1s/^)]\\}\\x27//'; }"
zellij action write 10
zellij action write-chars "gerrit_api '/changes/?q=project:go+status:merged&n=1&o=DETAILED_ACCOUNTS' 2>&1 | python3 -m json.tool | head -20"
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
- Environment requirements (curl command, Gerrit credentials)
- Skill definition format validation
- Command definition format validation
- Basic API functionality (fetch 1 Change from golang/go via Gerrit)
