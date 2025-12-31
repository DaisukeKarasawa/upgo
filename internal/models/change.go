package models

import "time"

// Change represents a Gerrit Change (CL)
type Change struct {
	ID            int        `json:"id"`
	RepositoryID  int        `json:"repository_id"`
	ChangeNumber  int        `json:"change_number"`  // Gerrit change number
	ChangeID      string     `json:"change_id"`      // Gerrit Change-Id (I...)
	Project       string     `json:"project"`        // 固定: "go"
	Branch        string     `json:"branch"`
	Subject       string     `json:"subject"`        // タイトル
	Message       string     `json:"message"`        // 説明
	Status        string     `json:"status"`         // NEW/MERGED/ABANDONED
	PreviousStatus string    `json:"previous_status"`
	Owner         string     `json:"owner"`          // オーナーのアカウント名
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	SubmittedAt   *time.Time `json:"submitted_at"`   // merged時
	URL           string     `json:"url"`            // Gerrit URL
	LastSyncedAt  *time.Time `json:"last_synced_at"`
}

// Revision represents a Gerrit patchset (revision)
type Revision struct {
	ID              int       `json:"id"`
	ChangeID        int       `json:"change_id"`        // Change.id
	PatchSetNumber  int       `json:"patch_set_number"` // パッチセット番号
	RevisionSHA     string    `json:"revision_sha"`    // コミットSHA
	Uploader        string    `json:"uploader"`        // アップロードした人のアカウント名
	CreatedAt       time.Time `json:"created_at"`
	Kind            string    `json:"kind"`            // REWORK/TRIVIAL_REBASE/MERGE_FIRST_PARENT_UPDATE/NO_CODE_CHANGE/NO_CHANGE
	CommitMessage   string    `json:"commit_message"`
	AuthorName      string    `json:"author_name"`
	AuthorEmail     string    `json:"author_email"`
	CommitterName   string    `json:"committer_name"`
	CommitterEmail  string    `json:"committer_email"`
}

// ChangeFile represents a file changed in a revision
type ChangeFile struct {
	ID              int    `json:"id"`
	RevisionID      int    `json:"revision_id"`
	FilePath        string `json:"file_path"`
	Status          string `json:"status"`           // A(Added)/D(Deleted)/M(Modified)/R(Renamed)/C(Copied)/W(Rewritten)
	OldPath         string `json:"old_path"`         // Renamed/Copiedの場合
	LinesInserted   int    `json:"lines_inserted"`
	LinesDeleted    int    `json:"lines_deleted"`
	SizeDelta       int64  `json:"size_delta"`       // バイト単位
	Size            int64  `json:"size"`              // バイト単位
	Binary          bool   `json:"binary"`
}

// ChangeDiff represents a diff for a file in a revision
type ChangeDiff struct {
	ID              int    `json:"id"`
	RevisionID      int    `json:"revision_id"`
	FilePath        string `json:"file_path"`
	DiffText        string `json:"diff_text"`        // パッチ全文（サイズ上限あり）
	DiffSize        int    `json:"diff_size"`        // バイト単位
	CreatedAt       time.Time `json:"created_at"`
}

// ChangeComment represents a comment on a change (change-level or inline)
type ChangeComment struct {
	ID              int       `json:"id"`
	ChangeID        int       `json:"change_id"`      // Change.id
	RevisionID      *int      `json:"revision_id"`    // Revision.id (nilの場合はchange-level)
	CommentID       string    `json:"comment_id"`     // Gerrit comment UUID
	FilePath        string    `json:"file_path"`      // inlineコメントの場合、空文字列の場合はchange-level
	Line            *int      `json:"line"`           // inlineコメントの行番号
	PatchSetNumber  int       `json:"patch_set_number"`
	Message         string    `json:"message"`
	Author          string    `json:"author"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	InReplyTo       string    `json:"in_reply_to"`    // 返信先コメントID
	Unresolved      bool      `json:"unresolved"`     // 未解決フラグ
}

// ChangeLabel represents a label vote on a change
type ChangeLabel struct {
	ID              int       `json:"id"`
	ChangeID        int       `json:"change_id"`      // Change.id
	LabelName       string    `json:"label_name"`     // Code-Review, Verified等
	Account         string    `json:"account"`       // 投票した人のアカウント名
	Value           int       `json:"value"`          // -2/-1/0/+1/+2
	Date            time.Time `json:"date"`
}

// ChangeMessage represents a change message (review message)
type ChangeMessage struct {
	ID              int       `json:"id"`
	ChangeID        int       `json:"change_id"`      // Change.id
	MessageID       string    `json:"message_id"`     // Gerrit message ID
	Author          string    `json:"author"`
	Message         string    `json:"message"`
	Date            time.Time `json:"date"`
	RevisionNumber  *int      `json:"revision_number"` // パッチセット番号
}
