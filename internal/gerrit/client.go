package gerrit

import (
	"bytes"
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

// Client wraps the Gerrit REST API client
type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
	logger   *zap.Logger
}

// BaseURL returns the base URL of the Gerrit instance
func (c *Client) BaseURL() string {
	return c.baseURL
}

// NewClient creates a new Gerrit client
// If username/password are empty, uses anonymous access
func NewClient(baseURL, username, password string, logger *zap.Logger) *Client {
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		username: username,
		password: password,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Gerrit JSON responses start with ")]}'\n" to prevent XSSI attacks
func stripGerritPrefix(data []byte) []byte {
	prefix := []byte(")]}'\n")
	if len(data) >= len(prefix) && bytes.Equal(data[:len(prefix)], prefix) {
		return data[len(prefix):]
	}
	return data
}

// makeRequest makes an HTTP request to Gerrit REST API
func (c *Client) makeRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	u := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}

	// Set authentication if provided
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
		// Use /a/ prefix for authenticated requests
		if !strings.Contains(path, "/a/") {
			u = strings.Replace(u, c.baseURL, c.baseURL+"/a", 1)
			req.URL, _ = url.Parse(u)
		}
	} else {
		// Anonymous access - ensure /a/ is not in path
		u = strings.Replace(u, "/a/", "/", 1)
		req.URL, _ = url.Parse(u)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

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
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	data = stripGerritPrefix(data)
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("JSONの解析に失敗しました: %w", err)
	}

	return nil
}

// ChangeInfo represents a Gerrit Change
type ChangeInfo struct {
	ID                   string                 `json:"id"`
	Project              string                 `json:"project"`
	Branch               string                 `json:"branch"`
	ChangeID             string                 `json:"change_id"`
	Subject              string                 `json:"subject"`
	Status               string                 `json:"status"` // NEW/MERGED/ABANDONED
	Created              string                 `json:"created"`
	Updated              string                 `json:"updated"`
	Submitted            string                 `json:"submitted,omitempty"`
	Number               int                    `json:"_number"`
	Owner                AccountInfo            `json:"owner"`
	CurrentRevision      string                 `json:"current_revision,omitempty"`
	Revisions            map[string]RevisionInfo `json:"revisions,omitempty"`
	Labels               map[string]LabelInfo   `json:"labels,omitempty"`
	Reviewers            map[string][]AccountInfo `json:"reviewers,omitempty"`
	Messages             []ChangeMessageInfo    `json:"messages,omitempty"`
	Insertions           int                    `json:"insertions"`
	Deletions            int                    `json:"deletions"`
	Mergeable            bool                   `json:"mergeable,omitempty"`
	WorkInProgress       bool                   `json:"work_in_progress,omitempty"`
}

// Message returns the commit message (subject + body)
// For Gerrit, the message is typically in the commit message of the current revision
func (c *ChangeInfo) Message() string {
	if c.CurrentRevision != "" && c.Revisions != nil {
		if rev, ok := c.Revisions[c.CurrentRevision]; ok {
			return rev.Commit.Message
		}
	}
	// Fallback: return subject if no commit message available
	return c.Subject
}

// AccountInfo represents a Gerrit account
type AccountInfo struct {
	AccountID int    `json:"_account_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

// RevisionInfo represents a Gerrit revision (patchset)
type RevisionInfo struct {
	Number      int                    `json:"_number"`
	Ref         string                 `json:"ref"`
	Created     string                 `json:"created"`
	Uploader    AccountInfo            `json:"uploader"`
	Kind        string                 `json:"kind"`
	Commit      CommitInfo             `json:"commit,omitempty"`
	Files       map[string]FileInfo    `json:"files,omitempty"`
	Fetch       map[string]FetchInfo   `json:"fetch,omitempty"`
}

// CommitInfo represents commit information
type CommitInfo struct {
	Commit    string       `json:"commit"`
	Parents   []ParentInfo `json:"parents"`
	Author    PersonInfo   `json:"author"`
	Committer PersonInfo   `json:"committer"`
	Subject   string       `json:"subject"`
	Message   string       `json:"message"`
}

// ParentInfo represents a parent commit
type ParentInfo struct {
	Commit  string `json:"commit"`
	Subject string `json:"subject"`
}

// PersonInfo represents author/committer information
type PersonInfo struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
	TZ    int    `json:"tz"`
}

// FileInfo represents file change information
type FileInfo struct {
	Status        string `json:"status"` // A/D/M/R/C/W
	OldPath       string `json:"old_path,omitempty"`
	LinesInserted int    `json:"lines_inserted,omitempty"`
	LinesDeleted  int    `json:"lines_deleted,omitempty"`
	SizeDelta     int64  `json:"size_delta,omitempty"`
	Size          int64  `json:"size,omitempty"`
	Binary        bool   `json:"binary,omitempty"`
}

// FetchInfo represents fetch information
type FetchInfo struct {
	URL      string            `json:"url"`
	Ref      string           `json:"ref"`
	Commands map[string]string `json:"commands,omitempty"`
}

// LabelInfo represents label information
type LabelInfo struct {
	All     []ApprovalInfo          `json:"all,omitempty"`
	Values  map[string]string       `json:"values,omitempty"`
	Approved *AccountInfo           `json:"approved,omitempty"`
	Rejected *AccountInfo           `json:"rejected,omitempty"`
	Recommended *AccountInfo       `json:"recommended,omitempty"`
	Disliked *AccountInfo          `json:"disliked,omitempty"`
	Blocking bool                  `json:"blocking,omitempty"`
}

// ApprovalInfo represents an approval
type ApprovalInfo struct {
	AccountInfo
	Value int    `json:"value"`
	Date  string `json:"date,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

// ChangeMessageInfo represents a change message
type ChangeMessageInfo struct {
	ID             string      `json:"id"`
	Author         AccountInfo `json:"author"`
	Date           string      `json:"date"`
	Message        string      `json:"message"`
	RevisionNumber int         `json:"_revision_number,omitempty"`
}

// CommentInfo represents a comment
type CommentInfo struct {
	ID         string      `json:"id"`
	Path       string      `json:"path,omitempty"`
	Line       int         `json:"line,omitempty"`
	PatchSet   int         `json:"patch_set,omitempty"`
	Message    string      `json:"message"`
	Author     AccountInfo `json:"author,omitempty"`
	Updated    string      `json:"updated"`
	InReplyTo  string      `json:"in_reply_to,omitempty"`
	Unresolved bool        `json:"unresolved,omitempty"`
}

// DiffInfo represents diff information
type DiffInfo struct {
	MetaA       DiffFileMetaInfo `json:"meta_a,omitempty"`
	MetaB       DiffFileMetaInfo `json:"meta_b,omitempty"`
	ChangeType  string           `json:"change_type"`
	DiffHeader  []string         `json:"diff_header,omitempty"`
	Content     []DiffContent    `json:"content,omitempty"`
	Binary      bool             `json:"binary,omitempty"`
}

// DiffFileMetaInfo represents file metadata in diff
type DiffFileMetaInfo struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Lines       int    `json:"lines"`
}

// DiffContent represents diff content
type DiffContent struct {
	AB []string `json:"ab,omitempty"` // Common lines
	A  []string `json:"a,omitempty"`  // Deleted lines
	B  []string `json:"b,omitempty"`  // Added lines
}

// QueryChanges queries changes with the given query string
func (c *Client) QueryChanges(ctx context.Context, query string, limit int, start int, options []string) ([]ChangeInfo, error) {
	path := fmt.Sprintf("/changes/?q=%s&n=%d", url.QueryEscape(query), limit)
	if start > 0 {
		// Gerrit uses "S" as the start offset parameter.
		path += fmt.Sprintf("&S=%d", start)
	}
	if len(options) > 0 {
		for _, opt := range options {
			path += "&o=" + url.QueryEscape(opt)
		}
	}

	var changes []ChangeInfo
	if err := c.getJSON(ctx, path, &changes); err != nil {
		return nil, err
	}

	return changes, nil
}

// GetChange retrieves a change by change ID
func (c *Client) GetChange(ctx context.Context, changeID string, options []string) (*ChangeInfo, error) {
	path := fmt.Sprintf("/changes/%s", url.QueryEscape(changeID))
	if len(options) > 0 {
		path += "?"
		for i, opt := range options {
			if i > 0 {
				path += "&"
			}
			path += "o=" + url.QueryEscape(opt)
		}
	}

	var change ChangeInfo
	if err := c.getJSON(ctx, path, &change); err != nil {
		return nil, err
	}

	return &change, nil
}

// GetChangeComments retrieves all comments for a change
func (c *Client) GetChangeComments(ctx context.Context, changeID string) (map[string][]CommentInfo, error) {
	path := fmt.Sprintf("/changes/%s/comments", url.QueryEscape(changeID))

	var comments map[string][]CommentInfo
	if err := c.getJSON(ctx, path, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// GetRevisionComments retrieves comments for a specific revision
func (c *Client) GetRevisionComments(ctx context.Context, changeID, revisionID string) (map[string][]CommentInfo, error) {
	path := fmt.Sprintf("/changes/%s/revisions/%s/comments", url.QueryEscape(changeID), url.QueryEscape(revisionID))

	var comments map[string][]CommentInfo
	if err := c.getJSON(ctx, path, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// GetFileDiff retrieves diff for a specific file in a revision
func (c *Client) GetFileDiff(ctx context.Context, changeID, revisionID, filePath string) (*DiffInfo, error) {
	path := fmt.Sprintf("/changes/%s/revisions/%s/files/%s/diff",
		url.QueryEscape(changeID),
		url.QueryEscape(revisionID),
		url.QueryEscape(filePath))

	var diff DiffInfo
	if err := c.getJSON(ctx, path, &diff); err != nil {
		return nil, err
	}

	return &diff, nil
}

// GetFileContent retrieves file content from a revision
func (c *Client) GetFileContent(ctx context.Context, changeID, revisionID, filePath string) ([]byte, error) {
	path := fmt.Sprintf("/changes/%s/revisions/%s/files/%s/content",
		url.QueryEscape(changeID),
		url.QueryEscape(revisionID),
		url.QueryEscape(filePath))

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み込みに失敗しました: %w", err)
	}

	// Gerrit returns base64-encoded content
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("base64デコードに失敗しました: %w", err)
	}

	return decoded, nil
}
