---
name: go-error-handling
description: Goのエラーハンドリングパターンとベストプラクティス。エラー処理、カスタムエラー型、エラーラッピング、sentinel errors について質問されたときに使用。
---

# Go エラーハンドリング

## 基本原則

### 1. エラーは常にチェックする

```go
// Good: エラーをチェック
result, err := doSomething()
if err != nil {
    return err
}

// Avoid: エラーを無視
result, _ := doSomething()  // 危険
```

### 2. エラーは一度だけ処理する

```go
// Good: 一度だけ処理
func process() error {
    if err := step1(); err != nil {
        return fmt.Errorf("step1 failed: %w", err)
    }
    return nil
}

// Avoid: ログして返す（二重処理）
func process() error {
    if err := step1(); err != nil {
        log.Printf("step1 failed: %v", err)  // ログ
        return err                            // さらに返す
    }
    return nil
}
```

## エラーラッピング

### fmt.Errorf with %w

```go
func ReadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        // %w でラップ（errors.Is/As で検査可能）
        return nil, fmt.Errorf("read config file %s: %w", path, err)
    }

    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, fmt.Errorf("parse config: %w", err)
    }

    return &cfg, nil
}
```

### errors.Is と errors.As

```go
// errors.Is: 特定のエラーかチェック
if errors.Is(err, os.ErrNotExist) {
    // ファイルが存在しない
}

// errors.As: エラーを特定の型に変換
var pathErr *os.PathError
if errors.As(err, &pathErr) {
    fmt.Printf("operation: %s, path: %s\n", pathErr.Op, pathErr.Path)
}
```

## カスタムエラー型

### シンプルなカスタムエラー

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

### Unwrap を実装したカスタムエラー

```go
type AppError struct {
    Code    int
    Message string
    Err     error  // 元のエラー
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

// 使用例
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

### 定義

```go
// パッケージレベルで定義
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrInvalidInput = errors.New("invalid input")
)
```

### 使用

```go
func GetUser(id string) (*User, error) {
    user, ok := users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return user, nil
}

// 呼び出し側
user, err := GetUser(id)
if errors.Is(err, ErrNotFound) {
    // 404 を返す
}
```

## エラーハンドリングパターン

### 早期リターン

```go
// Good: 早期リターン（フラットな構造）
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

// Avoid: ネストが深い
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

### defer でクリーンアップ

```go
func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()  // 確実にクローズ

    // 処理...
    return nil
}
```

### エラーの集約

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

## HTTP エラーハンドリング

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

## panic と recover

### panic は本当に回復不能な場合のみ

```go
// OK: 初期化時の致命的エラー
func MustCompile(pattern string) *regexp.Regexp {
    re, err := regexp.Compile(pattern)
    if err != nil {
        panic(err)  // プログラマのミス
    }
    return re
}

var emailRegex = MustCompile(`^[a-z]+@[a-z]+\.[a-z]+$`)
```

### recover はミドルウェアで

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

## アンチパターン

### 1. エラーメッセージの重複

```go
// Avoid
return fmt.Errorf("failed to read file: %w", err)
// err が "open /path: no such file" なら
// "failed to read file: open /path: no such file" になる

// Good: コンテキストを追加
return fmt.Errorf("loading config: %w", err)
```

### 2. エラーの握りつぶし

```go
// Avoid
result, _ := riskyOperation()

// Good
result, err := riskyOperation()
if err != nil {
    // 少なくともログに残す
    log.Printf("riskyOperation failed (ignored): %v", err)
}
```

### 3. 過度なエラーラッピング

```go
// Avoid: 各層でラップしすぎ
// "handler: service: repository: sql: connection refused"

// Good: 意味のある境界でのみラップ
// "get user 123: connection refused"
```
