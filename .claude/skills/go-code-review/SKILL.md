---
name: go-code-review
description: Goコードレビューのチェックリストとベストプラクティス。コードレビュー時、PRレビュー時、コード品質改善時に使用。
---

# Go コードレビューガイド

## レビューの優先順位

1. **正確性** - 正しく動作するか
2. **セキュリティ** - 脆弱性はないか
3. **パフォーマンス** - 効率的か
4. **可読性** - 理解しやすいか
5. **保守性** - 変更しやすいか

## 必須チェック項目

### 1. エラーハンドリング

```go
// Check: エラーは無視されていないか
result, err := doSomething()
if err != nil {
    return err  // または適切に処理
}

// Check: エラーメッセージにコンテキストがあるか
return fmt.Errorf("fetch user %s: %w", id, err)

// Check: sentinel errors は errors.Is で比較しているか
if errors.Is(err, ErrNotFound) {
    // ...
}
```

### 2. リソース管理

```go
// Check: Close() は defer で確実に呼ばれているか
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()

// Check: HTTP Response Body はクローズされているか
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()

// Check: database 接続はプールされているか
// (毎回 Open() していないか)
```

### 3. 並行処理

```go
// Check: goroutine リークの可能性はないか
// Check: データ競合はないか (go test -race)
// Check: context によるキャンセルに対応しているか

func worker(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // work
    }
}

// Check: WaitGroup の Add/Done の対応は正しいか
wg.Add(1)
go func() {
    defer wg.Done()
    // ...
}()
```

### 4. 入力検証

```go
// Check: ユーザー入力は検証されているか
func CreateUser(name string, age int) error {
    if name == "" {
        return errors.New("name is required")
    }
    if age < 0 || age > 150 {
        return errors.New("invalid age")
    }
    // ...
}

// Check: SQL インジェクション対策
// Avoid
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)

// Good
query := "SELECT * FROM users WHERE id = ?"
db.Query(query, id)
```

### 5. API 設計

```go
// Check: 関数シグネチャは適切か
// - context.Context は第一引数
// - エラーは最後の戻り値
func GetUser(ctx context.Context, id string) (*User, error)

// Check: Option パターンは適切に使われているか
type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.Timeout = d
    }
}
```

## コードスタイル

### 命名

```go
// Check: 変数名は明確か
// Avoid
var d int  // 何？
var data int  // まだ不明確

// Good
var daysSinceLastLogin int

// Check: 省略形は一般的か
// OK: id, url, http, ctx, err, req, resp
// Avoid: 独自の省略形

// Check: パッケージ名と重複していないか
// Avoid
package user
type User struct{}  // user.User

// Good
package user
type Info struct{}  // user.Info
```

### 構造

```go
// Check: 関数は適切な長さか (目安: 40行以下)
// Check: ネストは深すぎないか (目安: 3レベル以下)

// Check: 早期リターンを使っているか
// Avoid
func process(x int) int {
    if x > 0 {
        // 長い処理
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
    // 長い処理
    return result
}
```

### コメント

```go
// Check: 公開 API にはドキュメントがあるか
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

// Check: コメントは「なぜ」を説明しているか
// Avoid: 何をしているかの説明
// i をインクリメント
i++

// Good: なぜそうするかの説明
// GitHub API のレート制限を回避するため 1 秒待機
time.Sleep(1 * time.Second)
```

## パフォーマンス

### メモリ割り当て

```go
// Check: スライスの容量は事前確保されているか
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

// Check: strings.Builder を使っているか
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

### 不要なコピー

```go
// Check: 大きな構造体はポインタで渡しているか
type LargeStruct struct {
    Data [1024]byte
}

// Avoid: 値渡し（コピー発生）
func process(s LargeStruct)

// Good: ポインタ渡し
func process(s *LargeStruct)

// Check: range でのコピーを意識しているか
for i := range items {
    process(&items[i])  // コピーを避ける
}
```

## セキュリティ

### チェックリスト

```go
// [ ] SQL インジェクション対策（プレースホルダ使用）
// [ ] XSS 対策（html/template 使用）
// [ ] パスワードは平文で保存していない（bcrypt 等）
// [ ] 機密情報はログに出力していない
// [ ] HTTPS を強制している
// [ ] CORS 設定は適切か
// [ ] 認証・認可は正しく実装されているか
```

```go
// Check: 機密情報のログ出力
// Avoid
log.Printf("user login: %s, password: %s", user, password)

// Good
log.Printf("user login: %s", user)

// Check: タイミング攻撃対策
// Avoid
if password == storedPassword {

// Good
if subtle.ConstantTimeCompare([]byte(password), []byte(storedPassword)) == 1 {
```

## テスト

```go
// Check: テストは十分にあるか
// Check: エッジケースはテストされているか
// - nil 入力
// - 空文字列
// - 境界値
// - エラーケース

// Check: テストは独立しているか（順序依存していないか）
// Check: テストは決定的か（time.Now() に依存していないか）

// Check: テストヘルパーは t.Helper() を呼んでいるか
func assertEqual(t *testing.T, got, want int) {
    t.Helper()
    if got != want {
        t.Errorf("got %d, want %d", got, want)
    }
}
```

## レビューコメントの書き方

### 良いコメントの例

```
// 提案
Consider using `errors.Is()` here to handle wrapped errors correctly.

// 質問
Could you explain why we need this sleep? Is there a way to use
channels or WaitGroup instead?

// 必須
This SQL query is vulnerable to injection. Please use parameterized queries.

// 称賛
Great use of table-driven tests! This makes it easy to add new cases.
```

### 避けるべきコメント

```
// 曖昧
This looks wrong.

// 人格攻撃
Who wrote this terrible code?

// 過度に細かい
Change `i` to `index` (意味のある改善でない場合)
```

## レビュー効率化

```bash
# 静的解析を活用
go vet ./...
golangci-lint run

# データ競合検出
go test -race ./...

# テストカバレッジ確認
go test -cover ./...
```
