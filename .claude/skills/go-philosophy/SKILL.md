---
name: go-philosophy
description: Guide to understanding Go's design philosophy and writing idiomatic Go code. Use when you want to understand "why Go is designed this way" or learn Go idioms.
---

# Go Design Philosophy

## Core Principles

### 1. Simplicity

Go intentionally limits features. This is not a weakness, but a design philosophy.

```go
// Good: Simple and clear
func Process(items []Item) error {
    for _, item := range items {
        if err := item.Validate(); err != nil {
            return err
        }
    }
    return nil
}

// Avoid: Excessive abstraction
type Processor interface {
    Process(Processable) error
}
type Processable interface {
    Validate() error
}
```

**Rob Pike**: "Do more with less"

### 2. Explicitness

Go avoids implicit behavior and makes code intent clear.

```go
// Good: Explicitly handle errors
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}

// Avoid: Ignore errors
result, _ := doSomething()  // Unclear why ignored
```

### 3. Composition over Inheritance

Go has no inheritance. Instead, use composition.

```go
// Good: Composition via embedding
type Logger struct {
    prefix string
}

func (l *Logger) Log(msg string) {
    fmt.Printf("[%s] %s\n", l.prefix, msg)
}

type Server struct {
    Logger  // Embedding
    addr string
}

// Server has Logger's methods
server := &Server{Logger: Logger{prefix: "HTTP"}, addr: ":8080"}
server.Log("Starting...")  // Calls Logger.Log
```

### 4. Small Interfaces

```go
// Good: Single-method interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Compose as needed
type ReadWriter interface {
    Reader
    Writer
}

// Avoid: Large interfaces
type Everything interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    Seek(offset int64, whence int) (int64, error)
    // ...too many
}
```

## Go Proverbs

Understand Go proverbs by Rob Pike.

### "Don't communicate by sharing memory, share memory by communicating"

```go
// Avoid: Shared memory + mutex
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// Good: Communicate via channels
func counter(ch chan int) {
    count := 0
    for delta := range ch {
        count += delta
    }
}
```

### "The bigger the interface, the weaker the abstraction"

```go
// Weak: Large interface
type UserService interface {
    Create(user User) error
    Update(user User) error
    Delete(id string) error
    Get(id string) (User, error)
    List() ([]User, error)
    // Users depend on all methods
}

// Strong: Minimal necessary
type UserCreator interface {
    Create(user User) error
}

func RegisterUser(creator UserCreator, user User) error {
    return creator.Create(user)
}
```

### "Make the zero value useful"

```go
// Good: Zero value is useful
type Buffer struct {
    buf []byte
}

func (b *Buffer) Write(p []byte) (int, error) {
    b.buf = append(b.buf, p...)  // Append to nil slice is OK
    return len(p), nil
}

var buf Buffer  // Can use without initialization
buf.Write([]byte("hello"))

// sync.Mutex is also usable with zero value
var mu sync.Mutex
mu.Lock()  // No initialization needed
```

### "Clear is better than clever"

```go
// Clever but unclear
func f(n int) int { return n & (n - 1) }

// Clear
func clearLowestBit(n int) int {
    return n & (n - 1)  // Clear lowest bit
}
```

## Error Handling Philosophy

### Errors are Values

```go
// Treat errors as values
type PathError struct {
    Op   string
    Path string
    Err  error
}

func (e *PathError) Error() string {
    return e.Op + " " + e.Path + ": " + e.Err.Error()
}

func (e *PathError) Unwrap() error {
    return e.Err
}
```

### Decorate Errors

```go
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        // Add context
        return nil, fmt.Errorf("read config %s: %w", path, err)
    }
    // ...
}
```

## Concurrency Philosophy

### Goroutines are Lightweight

```go
// Don't hesitate to use goroutines
func processItems(items []Item) {
    var wg sync.WaitGroup
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            process(item)
        }(item)
    }
    wg.Wait()
}
```

### Propagate Cancellation with Context

```go
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    // Request is cancelled when ctx is cancelled
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

## Package Design Philosophy

### Avoid Circular Dependencies

```
// Good: One-way dependency
main → handler → service → repository

// Avoid: Circular dependency
service → repository → service  // Compile error
```

### Keep Package Names Concise

```go
// Good
import "encoding/json"
json.Marshal(data)

// Avoid
import "encoding/jsonencoder"
jsonencoder.Marshal(data)
```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Practical Go](https://dave.cheney.net/practical-go)
