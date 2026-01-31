---
name: go-testing
description: Go testing strategy and TDD practice guide. Use when asked about unit tests, table-driven tests, mocks, or benchmarks.
---

# Go Testing Strategy

## TDD Cycle

### Red → Green → Refactor

```go
// 1. Red: Write a failing test
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", result)
    }
}

// 2. Green: Minimal implementation
func Add(a, b int) int {
    return a + b
}

// 3. Refactor: Refactor as needed
```

## Table-Driven Tests

Table-driven tests define multiple test cases in table format and execute them in a loop.
This is a standard Go testing approach that makes it easy to add test cases and avoids code duplication.

**Benefits:**

- Easy to add test cases
- Avoids duplication of test logic
- Easy to compare test cases
- Can be executed individually with `t.Run`

**When to use:**

- When testing the same function with different inputs multiple times
- When testing error cases and success cases with the same structure
- When there are 5 or more test cases (fewer cases can use individual functions)

### Basic Pattern

```go
func TestCalculate(t *testing.T) {
    // Define test cases in table format
    tests := []struct {
        name     string  // Test case name (used with t.Run)
        input    int     // Input value
        expected int     // Expected value
    }{
        {"positive", 5, 10},
        {"zero", 0, 0},
        {"negative", -3, -6},
    }

    // Execute each test case in a loop
    for _, tt := range tests {
        // Execute as subtest with t.Run (can be executed and debugged individually)
        t.Run(tt.name, func(t *testing.T) {
            result := Calculate(tt.input)
            if result != tt.expected {
                t.Errorf("Calculate(%d) = %d, want %d",
                    tt.input, result, tt.expected)
            }
        })
    }
}
```

### Including Error Cases

When testing functions that return errors, use the `wantErr` flag to explicitly indicate whether an error is expected.
Managing error cases and success cases in the same table improves test coverage.

**Common pitfalls:**

- Forgetting `return` after error check causes subsequent assertions to execute
- Easy to forget return value validation in `wantErr: true` cases (validate as needed)

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config      // Expected value for success case
        wantErr bool         // Whether error is expected
    }{
        {
            name:  "valid config",
            input: `{"port": 8080}`,
            want:  &Config{Port: 8080},
            // wantErr: false is default
        },
        {
            name:    "invalid json",
            input:   `{invalid}`,
            wantErr: true,  // Error expected
        },
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseConfig(tt.input)

            // Validate error case
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return  // Early return for error case
            }

            // Validate success case
            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }

            // Validate return value
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %+v, want %+v", got, tt.want)
            }
        })
    }
}
```

## Test Helpers

Test helper functions reduce duplication in test code and improve readability.
Calling `t.Helper()` makes error messages point to the caller's line instead of inside the helper function.

**Why `t.Helper()` is needed:**

- Error messages point to the actual test code line
- Makes it easier to identify problem locations during debugging
- Test framework skips helper functions when generating stack traces

**Common pitfalls:**

- Forgetting to call `t.Helper()` makes error messages point inside helper function
- When using `t.Fatal` in helper functions, always call `t.Helper()` first

### t.Helper()

```go
// Helper functions must call t.Helper() first
func assertEqual(t *testing.T, got, want int) {
    t.Helper()  // Set error location to caller
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    result := Calculate(5)
    // Error message points to this line (due to t.Helper())
    assertEqual(t, result, 10)
}
```

### Test Setup

Manage resource setup and cleanup with helper functions.
Using `t.Cleanup()` ensures resources are released even if tests fail.

**Best practices:**

- Setup functions call `t.Helper()` first
- Register resource cleanup with `t.Cleanup()`
- Use `t.Fatalf` to immediately terminate tests on error

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()  // Mark as helper function

    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open db: %v", err)
    }

    // Ensures cleanup when test ends
    t.Cleanup(func() {
        if err := db.Close(); err != nil {
            t.Logf("failed to close db: %v", err)
        }
    })

    return db
}

func TestUserRepository(t *testing.T) {
    db := setupTestDB(t)  // Delegate setup to helper
    repo := NewUserRepository(db)
    // db.Close() is automatically called when test ends
}
```

## Interfaces and Mocks

Using interfaces allows you to replace external dependencies (databases, APIs, file systems, etc.) with mocks.
This makes tests fast, stable, and independent of production environments.

**Principles of testable design:**

- Abstract dependencies with interfaces
- Inject dependencies through constructors (dependency injection)
- Keep interfaces small with specific responsibilities

**Mocks vs Integration Tests:**

- Mocks: Fast, stable, suitable for unit tests
- Integration tests: Test actual dependencies, slow but highly reliable
- It's important to use both appropriately

### Testable Design

```go
// Define interface (depend on interface, not implementation)
type UserRepository interface {
    Get(id string) (*User, error)
    Save(user *User) error
}

// Production implementation (uses PostgreSQL)
type PostgresUserRepository struct {
    db *sql.DB
}

// Test mock (flexible behavior definition with function fields)
type MockUserRepository struct {
    GetFunc  func(id string) (*User, error)
    SaveFunc func(user *User) error
}

func (m *MockUserRepository) Get(id string) (*User, error) {
    if m.GetFunc != nil {
        return m.GetFunc(id)
    }
    return nil, errors.New("GetFunc not set")
}

func (m *MockUserRepository) Save(user *User) error {
    if m.SaveFunc != nil {
        return m.SaveFunc(user)
    }
    return errors.New("SaveFunc not set")
}
```

### Using Mocks

Mocks allow testing business logic without depending on databases or external APIs.
Define necessary behavior for each test case and keep tests independent.

**How to use mocks:**

- Define necessary behavior for each test case
- Test error cases easily
- Test multiple scenarios with subtests

```go
func TestUserService_GetUser(t *testing.T) {
    // Define mock repository (behavior can be customized per test case)
    mockRepo := &MockUserRepository{
        GetFunc: func(id string) (*User, error) {
            if id == "123" {
                return &User{ID: "123", Name: "Test"}, nil
            }
            return nil, ErrNotFound
        },
    }

    service := NewUserService(mockRepo)

    t.Run("existing user", func(t *testing.T) {
        user, err := service.GetUser("123")
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if user.Name != "Test" {
            t.Errorf("got name %q, want %q", user.Name, "Test")
        }
    })

    t.Run("not found", func(t *testing.T) {
        _, err := service.GetUser("999")
        // Check error type with errors.Is
        if !errors.Is(err, ErrNotFound) {
            t.Errorf("got error %v, want ErrNotFound", err)
        }
    })
}
```

## Subtests

Using `t.Run` allows logical grouping of tests and individual execution/debugging.
Combined with table-driven tests, each test case can be executed independently.

**Benefits:**

- Logically group test cases
- Execute specific subtests only (`go test -run TestMath/Addition`)
- Organized test output makes it clear which cases failed
- Can run in parallel (combined with `t.Parallel()`)

**When to use:**

- When grouping multiple related test cases
- When executing each case individually in table-driven tests
- When you want to clarify test structure

### t.Run

```go
func TestMath(t *testing.T) {
    // Group tests with subtests
    t.Run("Addition", func(t *testing.T) {
        t.Run("positive numbers", func(t *testing.T) {
            if Add(2, 3) != 5 {
                t.Error("failed")
            }
        })
        t.Run("negative numbers", func(t *testing.T) {
            if Add(-2, -3) != -5 {
                t.Error("failed")
            }
        })
    })

    t.Run("Multiplication", func(t *testing.T) {
        // Tests for another group
    })
}

// Execution examples:
// go test -run TestMath                    // Run all
// go test -run TestMath/Addition          // Addition group only
// go test -run TestMath/Addition/positive // Specific subtest only
```

### Parallel Tests

`t.Parallel()` allows running multiple tests in parallel.
This shortens test execution time, but care is needed when accessing shared resources.

**Notes:**

- Always copy loop variables to local variables (`tt := tt`)
- Avoid accessing shared resources (global variables, files, etc.)
- Use independent instances for external resources like databases per test

**Common pitfalls:**

- Forgetting to capture loop variables causes all tests to use the same value
- Accessing shared resources can cause race conditions

```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name  string
        input int
    }{
        {"test1", 1},
        {"test2", 2},
        {"test3", 3},
    }

    for _, tt := range tests {
        tt := tt  // Important: Copy loop variable to local variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Mark for parallel execution
            result := SlowOperation(tt.input)
            // assertions...
        })
    }
}

// Execution example:
// go test -parallel 4  // Run up to 4 tests in parallel
```

## Benchmarks

Benchmark tests measure code performance.
Run with `go test -bench` to compare execution time and memory usage.

**How to use benchmarks:**

- Detect performance regressions
- Measure optimization effectiveness
- Compare different implementations

**Execution methods:**

```bash
go test -bench=.           # Run all benchmarks
go test -bench=Fibonacci   # Specific benchmark only
go test -bench=. -benchmem # Also show memory allocations
```

### Basic

```go
func BenchmarkFibonacci(b *testing.B) {
    // b.N is automatically adjusted by benchmark framework
    // Executed repeatedly until sufficient statistical precision is achieved
    for i := 0; i < b.N; i++ {
        Fibonacci(20)
    }
}
```

### Memory Allocation

`b.ReportAllocs()` measures memory allocation count per iteration.
Useful for tracking memory efficiency improvements.

**Interpretation points:**

- `allocs/op`: Allocation count per operation
- `B/op`: Memory usage per operation (bytes)
- 0 allocs/op is ideal, but consider practical trade-offs

```go
func BenchmarkConcat(b *testing.B) {
    b.ReportAllocs()  // Report memory allocation
    for i := 0; i < b.N; i++ {
        Concat("hello", "world")
    }
}

// Output example:
// BenchmarkConcat-8    1000000    1200 ns/op    32 B/op    2 allocs/op
```

### Sub-benchmarks

Define sub-benchmarks with `b.Run` to compare performance under different conditions.
Effective for measuring scalability by varying data sizes or parameters.

**Notes:**

- Exclude setup time from measurement with `b.ResetTimer()`
- Place setup like data generation before `b.ResetTimer()`
- Each sub-benchmark runs independently

```go
func BenchmarkSort(b *testing.B) {
    sizes := []int{100, 1000, 10000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            // Setup (excluded from measurement)
            data := generateData(size)
            b.ResetTimer()  // Start measurement from here
            for i := 0; i < b.N; i++ {
                Sort(data)
            }
        })
    }
}

// Execution example:
// go test -bench=BenchmarkSort/size=100
```

## Test Coverage

```bash
# Measure coverage
go test -cover ./...

# Generate HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Tags

### Separating Integration Tests

```go
//go:build integration

package mypackage

func TestIntegration(t *testing.T) {
    // Integration test
}
```

```bash
# Run integration tests
go test -tags=integration ./...

# Exclude integration tests
go test ./...
```

## Golden Files

Golden file tests compare output with expected values stored in files.
Effective when output is complex and changes frequently, such as HTML rendering, code generation, or configuration file formatting.

**When to use:**

- When output is large and manual assertions are difficult
- When detecting output format changes
- When functioning as regression tests

**Notes:**

- Golden files must be included in version control
- Update with `-update` flag when making intentional changes
- Always check file I/O errors

```go
var update = flag.Bool("update", false, "update golden files")

func TestRender(t *testing.T) {
    result := Render(input)

    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {
        // Check error when updating golden file
        if err := os.WriteFile(golden, []byte(result), 0644); err != nil {
            t.Fatalf("failed to write golden file: %v", err)
        }
        return
    }

    // Always check read errors for golden file
    expected, err := os.ReadFile(golden)
    if err != nil {
        t.Fatalf("failed to read golden file: %v", err)
    }

    if result != string(expected) {
        t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", result, expected)
    }
}
```

## Anti-patterns

### 1. Sleep in Tests

```go
// Avoid
func TestAsync(t *testing.T) {
    go doSomething()
    time.Sleep(1 * time.Second)  // Unstable
    // assert...
}

// Good: Synchronize with channels or WaitGroup
func TestAsync(t *testing.T) {
    done := make(chan struct{})
    go func() {
        doSomething()
        close(done)
    }()
    <-done
    // assert...
}
```

### 2. Test Dependencies

```go
// Avoid: Dependent on test execution order
var globalState int

func TestA(t *testing.T) {
    globalState = 1
}

func TestB(t *testing.T) {
    if globalState != 1 {  // Fails unless TestA runs first
        t.Error("failed")
    }
}

// Good: Independent setup for each test
func TestB(t *testing.T) {
    state := 1  // Manage independent state within test
    if state != 1 {
        t.Error("failed")
    }
}
```

### 3. Ignoring Errors

```go
// Avoid: Ignoring errors
func TestReadFile(t *testing.T) {
    data, _ := os.ReadFile("test.txt")  // Ignore error
    // ...
}

// Good: Handle errors appropriately
func TestReadFile(t *testing.T) {
    data, err := os.ReadFile("test.txt")
    if err != nil {
        t.Fatalf("failed to read file: %v", err)
    }
    // ...
}
```

### 4. Tests Too Slow

```go
// Avoid: Tests depending on production environment
func TestAPI(t *testing.T) {
    resp, err := http.Get("https://api.example.com/data")
    // Network delay makes tests slow
}

// Good: Use mocks or test servers
func TestAPI(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"data": "test"}`))
    }))
    defer server.Close()

    resp, err := http.Get(server.URL)
    // Fast and stable test
}
```

## Best Practices

### Test Organization

**File structure:**

- Test files follow `*_test.go` naming convention
- Test files are placed in the same package (`package mypackage`)
- Integration tests are separated with `*_integration_test.go` or build tags

**Naming conventions:**

- Test functions start with `Test` (e.g., `TestCalculate`)
- Benchmark functions start with `Benchmark` (e.g., `BenchmarkSort`)
- Test case names should be descriptive (e.g., `"adds two positive numbers"` instead of `"positive numbers"`)

**Test grouping:**

```go
// Group related tests in the same file
// calculator_test.go
func TestAdd(t *testing.T) { /* ... */ }
func TestSubtract(t *testing.T) { /* ... */ }
func TestMultiply(t *testing.T) { /* ... */ }

// Tests for different features go in separate files
// parser_test.go
func TestParse(t *testing.T) { /* ... */ }
```

### Test Pattern Selection Guide

**Table-driven tests vs Individual test functions:**

| Situation                            | Recommended Pattern  | Reason                         |
| ------------------------------------ | -------------------- | ------------------------------ |
| 3 or fewer test cases                | Individual functions | Simple and readable            |
| 4 or more test cases                 | Table-driven         | Avoid duplication, easy to add |
| Many error cases                     | Table-driven         | Easy to cover error cases      |
| Setup differs significantly per case | Individual functions | Table becomes too complex      |

**Mocks vs Integration tests:**

| Situation                | Recommended Approach      | Reason                           |
| ------------------------ | ------------------------- | -------------------------------- |
| Business logic testing   | Mocks                     | Fast and stable                  |
| API contract testing     | Integration tests         | Verify actual behavior           |
| Database schema testing  | Integration tests         | Detect schema changes            |
| External service testing | Mocks + Integration tests | Test different aspects with both |

**Golden files vs Assertions:**

| Situation                   | Recommended Pattern | Reason                         |
| --------------------------- | ------------------- | ------------------------------ |
| Small, structured output    | Assertions          | Clear and readable             |
| Large, complex output       | Golden files        | Manual assertions difficult    |
| HTML/XML/JSON generation    | Golden files        | Detect format changes          |
| Numeric calculation results | Assertions          | Numeric comparison appropriate |

### Error Handling Best Practices

**Error handling in tests:**

1. **Always check I/O operation errors:**

```go
// Good
data, err := os.ReadFile("test.txt")
if err != nil {
    t.Fatalf("failed to read file: %v", err)
}

// Avoid
data, _ := os.ReadFile("test.txt")
```

2. **Fail immediately on setup errors with `t.Fatalf`:**

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open db: %v", err)  // Terminate immediately on setup failure
    }
    return db
}
```

3. **Validate errors under test appropriately:**

```go
// When error is expected
if err == nil {
    t.Error("expected error, got nil")
}

// When error is not expected
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}

// Validate specific error
if !errors.Is(err, ErrNotFound) {
    t.Errorf("got error %v, want ErrNotFound", err)
}
```

4. **Clear error messages:**

```go
// Good: Clearly state what was expected and what was got
if got != want {
    t.Errorf("Calculate(%d) = %d, want %d", input, got, want)
}

// Avoid: Vague messages
if got != want {
    t.Error("failed")
}
```

### Test Maintainability

**Ways to maintain test code quality:**

1. **Keep tests independent:**

   - Don't share state between tests
   - Perform necessary setup in each test
   - Avoid dependencies on global variables

2. **Make tests readable:**

   - Use descriptive test names
   - Make test case names clear about what's being tested
   - Extract complex logic into helper functions

3. **Keep tests concise:**

   - Test one thing per test
   - Consider splitting if tests are too long
   - Remove unnecessary assertions

4. **Keep tests stable:**

   - Avoid time-dependent tests (use channels instead of `time.Sleep`)
   - Fix seed when using random values
   - Minimize dependencies on external resources

5. **Keep tests fast:**
   - Avoid external dependencies with mocks
   - Use `t.Parallel()` for parallelizable tests
   - Minimize heavy setup

**Test refactoring:**

```go
// Before: Too much duplication
func TestAddPositive(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Error("failed")
    }
}

func TestAddNegative(t *testing.T) {
    result := Add(-2, -3)
    if result != -5 {
        t.Error("failed")
    }
}

// After: Refactored with table-driven test
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a        int
        b        int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### 3. Ignoring Errors

```go
// Avoid: Ignoring errors
func TestReadFile(t *testing.T) {
    data, _ := os.ReadFile("test.txt")  // Ignore error
    // ...
}

// Good: Handle errors appropriately
func TestReadFile(t *testing.T) {
    data, err := os.ReadFile("test.txt")
    if err != nil {
        t.Fatalf("failed to read file: %v", err)
    }
    // ...
}
```

### 4. Tests Too Slow

```go
// Avoid: Tests depending on production environment
func TestAPI(t *testing.T) {
    resp, err := http.Get("https://api.example.com/data")
    // Network delay makes tests slow
}

// Good: Use mocks or test servers
func TestAPI(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"data": "test"}`))
    }))
    defer server.Close()

    resp, err := http.Get(server.URL)
    // Fast and stable test
}
```
