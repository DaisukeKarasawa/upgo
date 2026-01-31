---
name: go-concurrency
description: Go concurrency patterns and best practices. Use when asked about goroutines, channels, sync package, or context.
---

# Go Concurrency Patterns

## Basic Principles

### "Don't communicate by sharing memory, share memory by communicating"

```go
// Avoid: Shared memory
var (
    data   map[string]int
    mu     sync.Mutex
)

func update(key string, value int) {
    mu.Lock()
    data[key] = value
    mu.Unlock()
}

// Good: Communicate via channels
type update struct {
    key   string
    value int
}

func manager(updates <-chan update) {
    data := make(map[string]int)
    for u := range updates {
        data[u.key] = u.value
    }
}
```

## Goroutine Patterns

### Basic Launch

```go
// Pass arguments
for _, item := range items {
    item := item  // Capture loop variable
    go func() {
        process(item)
    }()
}

// Or pass as argument
for _, item := range items {
    go func(item Item) {
        process(item)
    }(item)
}
```

### Wait with WaitGroup

```go
func processAll(items []Item) {
    var wg sync.WaitGroup

    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            process(item)
        }(item)
    }

    wg.Wait()  // Wait for all to complete
}
```

## Channel Patterns

### Basic Operations

```go
// unbuffered channel (synchronous)
ch := make(chan int)

// buffered channel (asynchronous)
ch := make(chan int, 10)

// Send
ch <- 42

// Receive
value := <-ch

// Close
close(ch)

// Check if closed
value, ok := <-ch
if !ok {
    // channel is closed
}
```

### Generator Pattern

```go
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for _, n := range nums {
            out <- n
        }
    }()
    return out
}

// Usage
for n := range generate(1, 2, 3, 4, 5) {
    fmt.Println(n)
}
```

### Fan-out / Fan-in

```go
// Fan-out: Distribute to multiple workers
func fanOut(in <-chan int, workers int) []<-chan int {
    outs := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        outs[i] = worker(in)
    }
    return outs
}

func worker(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- process(n)
        }
    }()
    return out
}

// Fan-in: Aggregate multiple channels into one
func fanIn(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)

    output := func(ch <-chan int) {
        defer wg.Done()
        for n := range ch {
            out <- n
        }
    }

    wg.Add(len(channels))
    for _, ch := range channels {
        go output(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

### Pipeline

```go
func pipeline() {
    // Stage 1: Generate
    nums := generate(1, 2, 3, 4, 5)

    // Stage 2: Square
    squared := square(nums)

    // Stage 3: Print
    for n := range squared {
        fmt.Println(n)
    }
}

func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        defer close(out)
        for n := range in {
            out <- n * n
        }
    }()
    return out
}
```

## Select Patterns

### Waiting on Multiple Channels

```go
select {
case msg := <-ch1:
    fmt.Println("ch1:", msg)
case msg := <-ch2:
    fmt.Println("ch2:", msg)
case ch3 <- "hello":
    fmt.Println("sent to ch3")
default:
    fmt.Println("no activity")
}
```

### Timeout

```go
select {
case result := <-ch:
    return result, nil
case <-time.After(5 * time.Second):
    return nil, errors.New("timeout")
}
```

### Cancellation

```go
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return  // Cancelled
        case job := <-jobs:
            process(job)
        }
    }
}
```

## Context Patterns

### Cancellation Propagation

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go worker(ctx)

    // Cancel on signal
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT)
    <-sigCh
    cancel()
}

func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("worker stopped:", ctx.Err())
            return
        default:
            doWork()
        }
    }
}
```

### Context with Timeout

```go
func fetchData(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

### Value Propagation

```go
type ctxKey string

const requestIDKey ctxKey = "requestID"

func withRequestID(ctx context.Context, id string) context.Context {
    return context.WithValue(ctx, requestIDKey, id)
}

func getRequestID(ctx context.Context) string {
    if v := ctx.Value(requestIDKey); v != nil {
        return v.(string)
    }
    return ""
}
```

## sync Package

### Mutex

```go
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}
```

### RWMutex

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    return item, ok
}

func (c *Cache) Set(key string, item Item) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items[key] = item
}
```

### Once

```go
var (
    instance *Singleton
    once     sync.Once
)

func GetInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}
```

### Pool

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func process(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer func() {
        buf.Reset()
        bufferPool.Put(buf)
    }()

    buf.Write(data)
    // ...
}
```

## Worker Pool

```go
func workerPool(jobs <-chan Job, results chan<- Result, workers int) {
    var wg sync.WaitGroup

    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- process(job)
            }
        }()
    }

    wg.Wait()
    close(results)
}
```

## Semaphore Pattern

```go
type Semaphore chan struct{}

func NewSemaphore(max int) Semaphore {
    return make(chan struct{}, max)
}

func (s Semaphore) Acquire() {
    s <- struct{}{}
}

func (s Semaphore) Release() {
    <-s
}

// Usage example
sem := NewSemaphore(10)  // Max 10 concurrent

for _, item := range items {
    sem.Acquire()
    go func(item Item) {
        defer sem.Release()
        process(item)
    }(item)
}
```

## Anti-patterns

### 1. Goroutine Leak

```go
// Avoid: Goroutine leaks if channel is never received
func leak() <-chan int {
    ch := make(chan int)
    go func() {
        ch <- expensiveComputation()  // Blocks forever
    }()
    return ch
}

// Good: Cancellable with context
func noLeak(ctx context.Context) <-chan int {
    ch := make(chan int)
    go func() {
        select {
        case ch <- expensiveComputation():
        case <-ctx.Done():
        }
    }()
    return ch
}
```

### 2. Data Race

```go
// Avoid: Data race
var counter int
go func() { counter++ }()
go func() { counter++ }()

// Good: Proper synchronization
var (
    counter int
    mu      sync.Mutex
)
go func() {
    mu.Lock()
    counter++
    mu.Unlock()
}()
```
