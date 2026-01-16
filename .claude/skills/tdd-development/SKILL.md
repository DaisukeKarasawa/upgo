---
name: tdd-development
description: Guide TDD (Test-Driven Development) workflow following Kent Beck's methodology. Use when writing new features, fixing bugs, or when the user mentions "TDD", "test first", or asks about testing approach.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

# TDD Development Skill

This skill guides development using Kent Beck's TDD methodology.

## TDD Cycle

### 1. Red Phase
- Write a failing test that defines expected behavior
- Test should be the simplest possible case
- Run test to confirm it fails for the right reason

### 2. Green Phase
- Write the minimum code to make the test pass
- Do not over-engineer or add extra functionality
- Focus only on making the current test pass

### 3. Refactor Phase
- Improve code structure while keeping tests green
- Remove duplication
- Improve naming and readability
- Run tests after each refactoring step

## Best Practices

### Test Naming
Use descriptive names that explain behavior:
- Go: `TestCalculateTax_WithValidInput_ReturnsCorrectAmount`
- TypeScript: `it('should calculate tax correctly for valid input')`

### Test Structure
Follow Arrange-Act-Assert pattern:
```go
func TestFunction(t *testing.T) {
    // Arrange
    input := setupTestData()

    // Act
    result := Function(input)

    // Assert
    if result != expected {
        t.Errorf("expected %v, got %v", expected, result)
    }
}
```

### Table-Driven Tests (Go)
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive number", 5, 10},
        {"zero", 0, 0},
        {"negative number", -5, -10},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Function(tt.input)
            if result != tt.expected {
                t.Errorf("expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

## Tidy First Principle

Separate commits into:
1. **Structural changes**: Refactoring without behavior changes
2. **Behavioral changes**: Adding/modifying functionality

Never mix these in the same commit.

## Commands

- Run Go tests: `make test`
- Run specific test: `go test -run TestName ./...`
- Run with verbose: `go test -v ./...`
- Run frontend tests: `cd web && npm test`
