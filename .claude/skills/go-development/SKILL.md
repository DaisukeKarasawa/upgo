---
name: go-development
description: Guide Go development best practices for this project. Use when writing Go code, working with Gin framework, SQLite database operations, or GitHub API integration.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

# Go Development Skill

This skill provides guidance for Go development in the Upgo project.

## Project Structure

```
internal/
├── api/          # HTTP handlers (Gin)
├── config/       # Configuration (Viper)
├── database/     # SQLite operations
├── github/       # GitHub API client
├── llm/          # Ollama integration
├── models/       # Data models
├── scheduler/    # Background tasks
└── service/      # Business logic
```

## Gin Framework Patterns

### Handler Structure
```go
func (h *Handler) GetPR(c *gin.Context) {
    id := c.Param("id")

    pr, err := h.service.GetPR(id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, pr)
}
```

### Middleware
```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Authentication logic
        c.Next()
    }
}
```

## Database Patterns (SQLite)

### Query with mattn/go-sqlite3
```go
func (r *Repository) GetByID(id int64) (*Model, error) {
    var m Model
    err := r.db.QueryRow(
        "SELECT id, name, created_at FROM models WHERE id = ?",
        id,
    ).Scan(&m.ID, &m.Name, &m.CreatedAt)

    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &m, err
}
```

### Transaction Pattern
```go
func (r *Repository) CreateWithTx(m *Model) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Operations...

    return tx.Commit()
}
```

## Error Handling

Use wrapped errors for context:
```go
import "fmt"

func DoSomething() error {
    if err := operation(); err != nil {
        return fmt.Errorf("failed to do something: %w", err)
    }
    return nil
}
```

## Logging with Zap

```go
import "go.uber.org/zap"

logger.Info("operation completed",
    zap.String("id", id),
    zap.Int("count", count),
)

logger.Error("operation failed",
    zap.Error(err),
)
```

## GitHub API (go-github)

```go
import "github.com/google/go-github/v60/github"

func (c *Client) GetPR(owner, repo string, number int) (*github.PullRequest, error) {
    pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
    return pr, err
}
```

## Testing

### Unit Test
```go
func TestService_GetPR(t *testing.T) {
    // Setup mock repository
    repo := &MockRepository{}
    service := NewService(repo)

    // Test
    result, err := service.GetPR(1)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

### HTTP Handler Test
```go
func TestHandler_GetPR(t *testing.T) {
    router := gin.New()
    handler := NewHandler(mockService)
    router.GET("/prs/:id", handler.GetPR)

    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/prs/1", nil)
    router.ServeHTTP(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }
}
```

## Commands

- Build: `go build -o bin/server cmd/server/main.go`
- Run: `go run cmd/server/main.go`
- Test: `go test ./...`
- Lint: `go vet ./...`
- Format: `go fmt ./...`
