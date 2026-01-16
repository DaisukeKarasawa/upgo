---
name: go-testing
description: Goのテスト戦略とTDD実践ガイド。ユニットテスト、テーブル駆動テスト、モック、ベンチマークについて質問されたときに使用。
---

# Go テスト戦略

## TDD サイクル

### Red → Green → Refactor

```go
// 1. Red: 失敗するテストを書く
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", result)
    }
}

// 2. Green: 最小限の実装
func Add(a, b int) int {
    return a + b
}

// 3. Refactor: 必要に応じてリファクタリング
```

## テーブル駆動テスト

### 基本パターン

```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive", 5, 10},
        {"zero", 0, 0},
        {"negative", -3, -6},
    }

    for _, tt := range tests {
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

### エラーケースを含む

```go
func TestParseConfig(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Config
        wantErr bool
    }{
        {
            name:  "valid config",
            input: `{"port": 8080}`,
            want:  &Config{Port: 8080},
        },
        {
            name:    "invalid json",
            input:   `{invalid}`,
            wantErr: true,
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

            if tt.wantErr {
                if err == nil {
                    t.Error("expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Errorf("unexpected error: %v", err)
                return
            }

            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %+v, want %+v", got, tt.want)
            }
        })
    }
}
```

## テストヘルパー

### t.Helper()

```go
func assertEqual(t *testing.T, got, want int) {
    t.Helper()  // エラー発生位置を呼び出し元に
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}

func TestSomething(t *testing.T) {
    result := Calculate(5)
    assertEqual(t, result, 10)  // エラーはこの行を指す
}
```

### テスト用セットアップ

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()

    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("failed to open db: %v", err)
    }

    t.Cleanup(func() {
        db.Close()
    })

    return db
}

func TestUserRepository(t *testing.T) {
    db := setupTestDB(t)
    repo := NewUserRepository(db)
    // テスト...
}
```

## インターフェースとモック

### テスト可能な設計

```go
// インターフェースを定義
type UserRepository interface {
    Get(id string) (*User, error)
    Save(user *User) error
}

// 本番実装
type PostgresUserRepository struct {
    db *sql.DB
}

// テスト用モック
type MockUserRepository struct {
    GetFunc  func(id string) (*User, error)
    SaveFunc func(user *User) error
}

func (m *MockUserRepository) Get(id string) (*User, error) {
    return m.GetFunc(id)
}

func (m *MockUserRepository) Save(user *User) error {
    return m.SaveFunc(user)
}
```

### モックの使用

```go
func TestUserService_GetUser(t *testing.T) {
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
        if !errors.Is(err, ErrNotFound) {
            t.Errorf("got error %v, want ErrNotFound", err)
        }
    })
}
```

## サブテスト

### t.Run

```go
func TestMath(t *testing.T) {
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
        // ...
    })
}
```

### 並列テスト

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
        tt := tt  // ループ変数をキャプチャ
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // 並列実行
            result := SlowOperation(tt.input)
            // assertions...
        })
    }
}
```

## ベンチマーク

### 基本

```go
func BenchmarkFibonacci(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Fibonacci(20)
    }
}
```

### メモリアロケーション

```go
func BenchmarkConcat(b *testing.B) {
    b.ReportAllocs()  // メモリ割り当てを報告
    for i := 0; i < b.N; i++ {
        Concat("hello", "world")
    }
}
```

### サブベンチマーク

```go
func BenchmarkSort(b *testing.B) {
    sizes := []int{100, 1000, 10000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
            data := generateData(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                Sort(data)
            }
        })
    }
}
```

## テストカバレッジ

```bash
# カバレッジ計測
go test -cover ./...

# HTML レポート生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## テスト用タグ

### 統合テストの分離

```go
//go:build integration

package mypackage

func TestIntegration(t *testing.T) {
    // 統合テスト
}
```

```bash
# 統合テストを実行
go test -tags=integration ./...

# 統合テストを除外
go test ./...
```

## ゴールデンファイル

```go
func TestRender(t *testing.T) {
    result := Render(input)

    golden := filepath.Join("testdata", t.Name()+".golden")

    if *update {
        os.WriteFile(golden, []byte(result), 0644)
    }

    expected, _ := os.ReadFile(golden)
    if result != string(expected) {
        t.Errorf("output mismatch\ngot:\n%s\nwant:\n%s", result, expected)
    }
}
```

## アンチパターン

### 1. テストでの sleep

```go
// Avoid
func TestAsync(t *testing.T) {
    go doSomething()
    time.Sleep(1 * time.Second)  // 不安定
    // assert...
}

// Good: チャネルや WaitGroup で同期
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

### 2. テスト間の依存

```go
// Avoid: テストの実行順序に依存
var globalState int

func TestA(t *testing.T) {
    globalState = 1
}

func TestB(t *testing.T) {
    if globalState != 1 {  // TestA の後でないと失敗
        t.Error("failed")
    }
}

// Good: 各テストで独立したセットアップ
```
