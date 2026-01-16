---
name: go-concurrency
description: Goの並行処理パターンとベストプラクティス。goroutine、channel、sync パッケージ、context について質問されたときに使用。
---

# Go 並行処理パターン

## 基本原則

### "Don't communicate by sharing memory, share memory by communicating"

```go
// Avoid: 共有メモリ
var (
    data   map[string]int
    mu     sync.Mutex
)

func update(key string, value int) {
    mu.Lock()
    data[key] = value
    mu.Unlock()
}

// Good: チャネルで通信
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

## Goroutine パターン

### 基本的な起動

```go
// 引数を渡す
for _, item := range items {
    item := item  // ループ変数をキャプチャ
    go func() {
        process(item)
    }()
}

// または引数として渡す
for _, item := range items {
    go func(item Item) {
        process(item)
    }(item)
}
```

### WaitGroup で待機

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

    wg.Wait()  // 全完了を待機
}
```

## Channel パターン

### 基本操作

```go
// unbuffered channel（同期）
ch := make(chan int)

// buffered channel（非同期）
ch := make(chan int, 10)

// 送信
ch <- 42

// 受信
value := <-ch

// クローズ
close(ch)

// クローズ確認
value, ok := <-ch
if !ok {
    // channel はクローズされている
}
```

### Generator パターン

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

// 使用
for n := range generate(1, 2, 3, 4, 5) {
    fmt.Println(n)
}
```

### Fan-out / Fan-in

```go
// Fan-out: 複数のワーカーに分散
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

// Fan-in: 複数のチャネルを1つに集約
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

## Select パターン

### 複数チャネルの待機

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

### タイムアウト

```go
select {
case result := <-ch:
    return result, nil
case <-time.After(5 * time.Second):
    return nil, errors.New("timeout")
}
```

### キャンセル

```go
func worker(ctx context.Context, jobs <-chan Job) {
    for {
        select {
        case <-ctx.Done():
            return  // キャンセルされた
        case job := <-jobs:
            process(job)
        }
    }
}
```

## Context パターン

### キャンセル伝播

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go worker(ctx)

    // シグナルでキャンセル
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

### タイムアウト付きコンテキスト

```go
func fetchData(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

### 値の伝播

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

## sync パッケージ

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

## Semaphore パターン

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

// 使用例
sem := NewSemaphore(10)  // 最大10並列

for _, item := range items {
    sem.Acquire()
    go func(item Item) {
        defer sem.Release()
        process(item)
    }(item)
}
```

## アンチパターン

### 1. Goroutine リーク

```go
// Avoid: チャネルが受信されないと goroutine がリーク
func leak() <-chan int {
    ch := make(chan int)
    go func() {
        ch <- expensiveComputation()  // 永遠にブロック
    }()
    return ch
}

// Good: context でキャンセル可能に
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

### 2. データ競合

```go
// Avoid: データ競合
var counter int
go func() { counter++ }()
go func() { counter++ }()

// Good: 適切な同期
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
