package gitiles

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

// Client wraps the Gitiles API client
type Client struct {
	baseURL string
	client  *http.Client
	logger  *zap.Logger
}

// NewClient creates a new Gitiles client
func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// makeRequest makes an HTTP request to Gitiles API
func (c *Client) makeRequest(ctx context.Context, method, path string) (*http.Response, error) {
	u := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエストの実行に失敗しました: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		resp.Body.Close()
		return nil, fmt.Errorf("APIエラー: %d %s", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// getJSON performs a GET request and unmarshals JSON response
func (c *Client) getJSON(ctx context.Context, path string, result interface{}) error {
	resp, err := c.makeRequest(ctx, "GET", path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("JSONの解析に失敗しました: %w", err)
	}

	return nil
}

// RefInfo represents a ref (branch/tag)
type RefInfo struct {
	Ref    string `json:"ref"`
	Object string `json:"object"` // SHA-1
}

// ListRefs lists all refs (branches and tags)
func (c *Client) ListRefs(ctx context.Context, project string) (map[string]RefInfo, error) {
	path := fmt.Sprintf("/%s/+refs?format=JSON", url.QueryEscape(project))

	resp, err := c.makeRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	// Gitiles JSON responses start with ")]}'\n"
	prefix := []byte(")]}'\n")
	if len(data) >= len(prefix) && string(data[:len(prefix)]) == string(prefix) {
		data = data[len(prefix):]
	}

	var refs map[string]RefInfo
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, fmt.Errorf("JSONの解析に失敗しました: %w", err)
	}

	return refs, nil
}

// LogInfo represents commit log information
type LogInfo struct {
	Log  []CommitInfo `json:"log"`
	Next string       `json:"next,omitempty"`
}

// CommitInfo represents commit information from Gitiles
type CommitInfo struct {
	Commit    string       `json:"commit"`
	Tree      string       `json:"tree"`
	Parents   []string     `json:"parents"`
	Author    PersonInfo   `json:"author"`
	Committer PersonInfo   `json:"committer"`
	Message   string       `json:"message"`
}

// PersonInfo represents author/committer information
type PersonInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Time  string `json:"time"`
}

// GetLog retrieves commit log for a ref
func (c *Client) GetLog(ctx context.Context, project, ref string, limit int, startCommit string) (*LogInfo, error) {
	path := fmt.Sprintf("/%s/+log/%s?format=JSON&n=%d", url.QueryEscape(project), url.QueryEscape(ref), limit)
	if startCommit != "" {
		path += "&s=" + url.QueryEscape(startCommit)
	}

	resp, err := c.makeRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	// Gitiles JSON responses start with ")]}'\n"
	prefix := []byte(")]}'\n")
	if len(data) >= len(prefix) && string(data[:len(prefix)]) == string(prefix) {
		data = data[len(prefix):]
	}

	var logInfo LogInfo
	if err := json.Unmarshal(data, &logInfo); err != nil {
		return nil, fmt.Errorf("JSONの解析に失敗しました: %w", err)
	}

	return &logInfo, nil
}

// GetCommit retrieves a single commit
func (c *Client) GetCommit(ctx context.Context, project, commitSHA string) (*CommitInfo, error) {
	path := fmt.Sprintf("/%s/+show/%s?format=JSON", url.QueryEscape(project), url.QueryEscape(commitSHA))

	resp, err := c.makeRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	// Gitiles JSON responses start with ")]}'\n"
	prefix := []byte(")]}'\n")
	if len(data) >= len(prefix) && string(data[:len(prefix)]) == string(prefix) {
		data = data[len(prefix):]
	}

	var commit CommitInfo
	if err := json.Unmarshal(data, &commit); err != nil {
		return nil, fmt.Errorf("JSONの解析に失敗しました: %w", err)
	}

	return &commit, nil
}

// GetFileContent retrieves file content at a specific commit
func (c *Client) GetFileContent(ctx context.Context, project, ref, filePath string) ([]byte, error) {
	path := fmt.Sprintf("/%s/+show/%s/%s?format=TEXT", url.QueryEscape(project), url.QueryEscape(ref), url.QueryEscape(filePath))

	resp, err := c.makeRequest(ctx, "GET", path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	// Gitiles returns base64-encoded content when format=TEXT
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("base64デコードに失敗しました: %w", err)
	}

	return decoded, nil
}

// GetDiff retrieves diff between two commits
func (c *Client) GetDiff(ctx context.Context, project, fromCommit, toCommit string) (string, error) {
	path := fmt.Sprintf("/%s/+diff/%s..%s?format=TEXT", url.QueryEscape(project), url.QueryEscape(fromCommit), url.QueryEscape(toCommit))

	resp, err := c.makeRequest(ctx, "GET", path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	// Gitiles returns base64-encoded content when format=TEXT
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", fmt.Errorf("base64デコードに失敗しました: %w", err)
	}

	return string(decoded), nil
}
