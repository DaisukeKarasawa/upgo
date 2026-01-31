---
name: go-error-handling
description: Go error handling patterns and best practices. Use when asked about error handling, custom error types, error wrapping, or sentinel errors.
---

# Go Error Handling

## Basic Principles

### 1. Always Check Errors

```go
// Good: Check error
result, err := doSomething()
if err != nil {
    return err
}

// Avoid: Ignore error
result, _ := doSomething()  // Dangerous
```

### 2. Handle Errors Only Once

```go
// Good: Handle once
func process() error {
    if err := step1(); err != nil {
        return fmt.Errorf("step1 failed: %w", err)
    }
    return nil
}

// Avoid: Log and return (double handling)
func process() error {
    if err := step1(); err != nil {
        log.Printf("step1 failed: %v", err)  // Log
        return err                            // Also return
    }
    return nil
}
```

## Error Wrapping

### fmt.Errorf with %w

```go
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        // Wrap with %w (can be inspected with errors.Is/As)
        return nil, fmt.Errorf("read config file %s: %w", path, err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    return &cfg, nil
}
```

### errors.Is and errors.As

```go
// errors.Is: Check for specific error
if errors.Is(err, os.ErrNotExist) {
    // File does not exist
}

// errors.As: Convert error to specific type
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Printf("operation: %s, path: %s\n", pathErr.Op, pathErr.Path)
}
```

## Custom Error Types

### Simple Custom Error

```go
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

func Validate(user User) error {
    if user.Name == "" {
        return &ValidationError{Field: "name", Message: "required"}
    }
    return nil
}
```

### Custom Error with Unwrap

```go
type AppError struct {
    Code    int
    Message string
    Err     error  // Original error
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Usage example
func fetchUser(id string) (*User, error) {
    user, err := db.GetUser(id)
    if err != nil {
        return nil, &AppError{
            Code:    500,
            Message: "failed to fetch user",
            Err:     err,
        }
    }
    return user, nil
}
```

## Sentinel Errors

### Definition

```go
// Define at package level
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)
```

### Usage

```go
func GetUser(id string) (*User, error) {
    user, ok := users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return user, nil
}

// Caller side
user, err := GetUser(id)
if errors.Is(err, ErrNotFound) {
    // Return 404
}
```

## Error Handling Patterns

### Early Return

```go
// Good: Early return (flat structure)
func process(data []byte) error {
    if len(data) == 0 {
        return errors.New("empty data")
    }

    result, err := parse(data)
    if err != nil {
        return fmt.Errorf("parse: %w", err)
    }

    if err := validate(result); err != nil {
        return fmt.Errorf("validate: %w", err)
    }

    return save(result)
}

// Avoid: Deep nesting
func process(data []byte) error {
    if len(data) > 0 {
        result, err := parse(data)
        if err == nil {
            if err := validate(result); err == nil {
                return save(result)
            } else {
                return err
            }
        } else {
            return err
        }
    }
    return errors.New("empty data")
}
```

### Cleanup with defer

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // Ensure close

    // Process...
    return nil
}
```

### Error Aggregation

```go
func validateUser(u User) error {
    var errs []error

    if u.Name == "" {
        errs = append(errs, errors.New("name is required"))
    }
    if u.Email == "" {
        errs = append(errs, errors.New("email is required"))
    }
    if u.Age < 0 {
        errs = append(errs, errors.New("age must be positive"))
    }

    return errors.Join(errs...)  // Go 1.20+
}
```

## HTTP Error Handling

```go
func handleGetUser(w http.ResponseWriter, r *http.Request) {
    id := r.URL.Query().Get("id")

    user, err := userService.Get(id)
    if err != nil {
        switch {
        case errors.Is(err, ErrNotFound):
            http.Error(w, "User not found", http.StatusNotFound)
        case errors.Is(err, ErrUnauthorized):
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
        default:
            log.Printf("unexpected error: %v", err)
            http.Error(w, "Internal server error", http.StatusInternalServerError)
        }
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

## panic and recover

### panic Only for Truly Unrecoverable Cases

```go
// OK: Fatal error during initialization
func MustCompile(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        panic(err)  // Programmer error
    }
    return re
}

var emailRegex = MustCompile(`^[a-z]+@[a-z]+\.[a-z]+$`)
```

### recover in Middleware

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v", err)
                http.Error(w, "Internal Server Error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

## Anti-patterns

### 1. Duplicate Error Messages

```go
// Avoid
return fmt.Errorf("failed to read file: %w", err)
// If err is "open /path: no such file"
// Result: "failed to read file: open /path: no such file"

// Good: Add context
return fmt.Errorf("loading config: %w", err)
```

### 2. Swallowing Errors

```go
// Avoid
result, _ := riskyOperation()

// Good
result, err := riskyOperation()
if err != nil {
    // At least log it
    log.Printf("riskyOperation failed (ignored): %v", err)
}
```

### 3. Excessive Error Wrapping

```go
// Avoid: Wrapping at every layer
// "handler: service: repository: sql: connection refused"

// Good: Wrap only at meaningful boundaries
// "get user 123: connection refused"
```
