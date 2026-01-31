---
name: go-code-review
description: Checklist and best practices for Go code review. Use when reviewing code, PRs, or improving code quality.
---

# Go Code Review Guide

## Review Priority

1. **Correctness** - Does it work correctly?
2. **Security** - Are there vulnerabilities?
3. **Performance** - Is it efficient?
4. **Readability** - Is it easy to understand?
5. **Maintainability** - Is it easy to modify?

## Required Check Items

### 1. Error Handling

```go
// Check: Are errors not ignored?
result, err := doSomething()
if err != nil {
    return err  // or handle appropriately
}

// Check: Do error messages have context?
return fmt.Errorf("fetch user %s: %w", id, err)

// Check: Are sentinel errors compared with errors.Is?
if errors.Is(err, ErrNotFound) {
    // ...
}
```

### 2. Resource Management

```go
// Check: Is Close() called reliably with defer?
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()

// Check: Is HTTP Response Body closed?
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()

// Check: Are database connections pooled?
// (Not calling Open() every time)
```

### 3. Concurrency

```go
// Check: Is there a possibility of goroutine leaks?
// Check: Are there data races? (go test -race)
// Check: Does it handle cancellation via context?

func worker(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // work
    }
}

// Check: Are WaitGroup Add/Done pairs correct?
wg.Add(1)
go func() {
    defer wg.Done()
    // ...
}()
```

### 4. Input Validation

```go
// Check: Is user input validated?
func CreateUser(name string, age int) error {
    if name == "" {
        return errors.New("name is required")
    }
    if age < 0 || age > 150 {
        return errors.New("invalid age")
    }
    // ...
}

// Check: SQL injection prevention
// Avoid
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)

// Good
query := "SELECT * FROM users WHERE id = ?"
db.Query(query, id)
```

### 5. API Design

```go
// Check: Is the function signature appropriate?
// - context.Context is the first argument
// - Error is the last return value
func GetUser(ctx context.Context, id string) (*User, error)

// Check: Is the Option pattern used appropriately?
type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.Timeout = d
    }
}
```

## Code Style

### Naming

```go
// Check: Are variable names clear?
// Avoid
var d int  // What?
var data int  // Still unclear

// Good
var daysSinceLastLogin int

// Check: Are abbreviations common?
// OK: id, url, http, ctx, err, req, resp
// Avoid: Custom abbreviations

// Check: Does it conflict with package name?
// Avoid
package user
type User struct{}  // user.User

// Good
package user
type Info struct{}  // user.Info
```

### Structure

```go
// Check: Are functions appropriately sized? (Guideline: under 40 lines)
// Check: Is nesting too deep? (Guideline: under 3 levels)

// Check: Is early return used?
// Avoid
func process(x int) int {
    if x > 0 {
        // long process
        return result
    } else {
        return 0
    }
}

// Good
func process(x int) int {
    if x <= 0 {
        return 0
    }
    // long process
    return result
}
```

### Comments

```go
// Check: Do public APIs have documentation?
// Package user provides user management functionality.
package user

// User represents a registered user in the system.
type User struct {
    ID   string
    Name string
}

// GetByID retrieves a user by their unique identifier.
// It returns ErrNotFound if the user does not exist.
func GetByID(id string) (*User, error)

// Check: Do comments explain "why"?
// Avoid: Explaining what it does
// Increment i
i++

// Good: Explaining why
// Wait 1 second to avoid GitHub API rate limits
time.Sleep(1 * time.Second)
```

## Performance

### Memory Allocation

```go
// Check: Is slice capacity pre-allocated?
// Avoid
var items []Item
for _, raw := range rawItems {
    items = append(items, parse(raw))
}

// Good
items := make([]Item, 0, len(rawItems))
for _, raw := range rawItems {
    items = append(items, parse(raw))
}

// Check: Is strings.Builder used?
// Avoid
var s string
for _, item := range items {
    s += item.String()
}

// Good
var sb strings.Builder
for _, item := range items {
    sb.WriteString(item.String())
}
s := sb.String()
```

### Unnecessary Copies

```go
// Check: Are large structs passed by pointer?
type LargeStruct struct {
    Data [1024]byte
}

// Avoid: Pass by value (copy occurs)
func process(s LargeStruct)

// Good: Pass by pointer
func process(s *LargeStruct)

// Check: Are copies in range loops considered?
for i := range items {
    process(&items[i])  // Avoid copy
}
```

## Security

### Checklist

```go
// [ ] SQL injection prevention (use placeholders)
// [ ] XSS prevention (use html/template)
// [ ] Passwords not stored in plain text (use bcrypt, etc.)
// [ ] Sensitive information not logged
// [ ] HTTPS enforced
// [ ] CORS settings appropriate?
// [ ] Authentication and authorization correctly implemented?
```

```go
// Check: Logging sensitive information
// Avoid
log.Printf("user login: %s, password: %s", user, password)

// Good
log.Printf("user login: %s", user)

// Check: Timing attack prevention
// Avoid
if password == storedPassword {

// Good
if subtle.ConstantTimeCompare([]byte(password), []byte(storedPassword)) == 1 {
```

## Testing

```go
// Check: Are there sufficient tests?
// Check: Are edge cases tested?
// - nil input
// - empty string
// - boundary values
// - error cases

// Check: Are tests independent? (not order-dependent)
// Check: Are tests deterministic? (not dependent on time.Now())

// Check: Do test helpers call t.Helper()?
func assertEqual(t *testing.T, got, want int) {
    t.Helper()
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}
```

## How to Write Review Comments

### Examples of Good Comments

```
// Suggestion
Consider using `errors.Is()` here to handle wrapped errors correctly.

// Question
Could you explain why we need this sleep? Is there a way to use
channels or WaitGroup instead?

// Required
This SQL query is vulnerable to injection. Please use parameterized queries.

// Praise
Great use of table-driven tests! This makes it easy to add new cases.
```

### Comments to Avoid

```
// Vague
This looks wrong.

// Personal attack
Who wrote this terrible code?

// Overly nitpicky
Change `i` to `index` (when it's not a meaningful improvement)
```

## Review Efficiency

```bash
# Use static analysis
go vet ./...
golangci-lint run

# Detect data races
go test -race ./...

# Check test coverage
go test -cover ./...
```
