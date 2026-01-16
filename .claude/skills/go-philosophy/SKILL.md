---
name: go-philosophy
description: Go言語の設計思想と哲学を理解し、Goらしいコードを書くためのガイド。「なぜGoはこう設計されたのか」を理解したいとき、Goのイディオムを学びたいときに使用。
---

# Go の設計思想と哲学

## 核となる原則

### 1. シンプルさ (Simplicity)

Go は意図的に機能を制限しています。これは欠点ではなく、設計思想です。

```go
// Good: シンプルで明確
func Process(items []Item) error {
    for _, item := range items {
        if err := item.Validate(); err != nil {
            return err
        }
    }
    return nil
}

// Avoid: 過度な抽象化
type Processor interface {
    Process(Processable) error
}
type Processable interface {
    Validate() error
}
```

**Rob Pike**: "少ないものでより多くを達成する"

### 2. 明示性 (Explicitness)

Go は暗黙の動作を避け、コードの意図を明確にします。

```go
// Good: エラーを明示的に処理
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}

// Avoid: エラーを無視
result, _ := doSomething()  // なぜ無視するのか不明
```

### 3. 合成 (Composition over Inheritance)

Go には継承がありません。代わりに合成を使います。

```go
// Good: 埋め込みによる合成
type Logger struct {
    prefix string
}

func (l *Logger) Log(msg string) {
    fmt.Printf("[%s] %s\n", l.prefix, msg)
}

type Server struct {
    Logger  // 埋め込み
    addr string
}

// Server は Logger のメソッドを持つ
server := &Server{Logger: Logger{prefix: "HTTP"}, addr: ":8080"}
server.Log("Starting...")  // Logger.Log を呼び出し
```

### 4. インターフェースは小さく

```go
// Good: 単一メソッドのインターフェース
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 必要に応じて合成
type ReadWriter interface {
    Reader
    Writer
}

// Avoid: 大きなインターフェース
type Everything interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
    Close() error
    Seek(offset int64, whence int) (int64, error)
    // ...多すぎる
}
```

## Go Proverbs (Go の格言)

Rob Pike による Go の格言を理解しましょう。

### "Don't communicate by sharing memory, share memory by communicating"

```go
// Avoid: 共有メモリ + mutex
var (
    counter int
    mu      sync.Mutex
)

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// Good: チャネルで通信
func counter(ch chan int) {
    count := 0
    for delta := range ch {
        count += delta
    }
}
```

### "The bigger the interface, the weaker the abstraction"

```go
// Weak: 大きなインターフェース
type UserService interface {
    Create(user User) error
    Update(user User) error
    Delete(id string) error
    Get(id string) (User, error)
    List() ([]User, error)
    // 使う側は全メソッドに依存
}

// Strong: 必要最小限
type UserCreator interface {
    Create(user User) error
}

func RegisterUser(creator UserCreator, user User) error {
    return creator.Create(user)
}
```

### "Make the zero value useful"

```go
// Good: ゼロ値が有用
type Buffer struct {
    buf []byte
}

func (b *Buffer) Write(p []byte) (int, error) {
    b.buf = append(b.buf, p...)  // nil スライスへの append は OK
    return len(p), nil
}

var buf Buffer  // 初期化なしで使える
buf.Write([]byte("hello"))

// sync.Mutex もゼロ値で使える
var mu sync.Mutex
mu.Lock()  // 初期化不要
```

### "Clear is better than clever"

```go
// Clever but unclear
func f(n int) int { return n & (n - 1) }

// Clear
func clearLowestBit(n int) int {
    return n & (n - 1)  // 最下位ビットをクリア
}
```

## エラー処理の哲学

### エラーは値である

```go
// エラーを値として扱う
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

### エラーを装飾する

```go
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        // コンテキストを追加
        return nil, fmt.Errorf("read config %s: %w", path, err)
    }
    // ...
}
```

## 並行処理の哲学

### Goroutine は軽量

```go
// goroutine を躊躇なく使う
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

### Context でキャンセルを伝播

```go
func fetchData(ctx context.Context, url string) ([]byte, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    // ctx がキャンセルされるとリクエストも中断
    resp, err := http.DefaultClient.Do(req)
    // ...
}
```

## パッケージ設計の哲学

### 循環依存を避ける

```
// Good: 一方向の依存
main → handler → service → repository

// Avoid: 循環依存
service → repository → service  // コンパイルエラー
```

### パッケージ名は簡潔に

```go
// Good
import "encoding/json"
json.Marshal(data)

// Avoid
import "encoding/jsonencoder"
jsonencoder.Marshal(data)
```

## 参考資料

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Proverbs](https://go-proverbs.github.io/)
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- [Practical Go](https://dave.cheney.net/practical-go)
