package models

import "time"

type Change struct {
	ID           int        `json:"id"`
	ChangeID     string     `json:"change_id"`
	ChangeNumber int        `json:"change_number"`
	Project      string     `json:"project"`
	Branch       string     `json:"branch"`
	Status       string     `json:"status"`
	Subject      string     `json:"subject"`
	Message      string     `json:"message"`
	OwnerName    string     `json:"owner_name"`
	OwnerEmail   string     `json:"owner_email"`
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	Submitted    *time.Time `json:"submitted,omitempty"`
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`
}

type Revision struct {
	ID            int       `json:"id"`
	ChangeDBID    int       `json:"change_db_id"`
	RevisionID    string    `json:"revision_id"`
	PatchsetNum   int       `json:"patchset_num"`
	UploaderName  string    `json:"uploader_name"`
	UploaderEmail string    `json:"uploader_email"`
	Created       time.Time `json:"created"`
	CommitMessage string    `json:"commit_message"`
}

type File struct {
	ID            int    `json:"id"`
	RevisionDBID  int    `json:"revision_db_id"`
	FilePath      string `json:"file_path"`
	Status        string `json:"status"`
	LinesInserted int    `json:"lines_inserted"`
	LinesDeleted  int    `json:"lines_deleted"`
	SizeDelta     int    `json:"size_delta"`
}

type Diff struct {
	ID           int    `json:"id"`
	FileDBID     int    `json:"file_db_id"`
	DiffContent  string `json:"diff_content"`
	IsBinary     bool   `json:"is_binary"`
	SizeExceeded bool   `json:"size_exceeded"`
}

type ChangeComment struct {
	ID           int        `json:"id"`
	ChangeDBID   int        `json:"change_db_id"`
	CommentID    string     `json:"comment_id"`
	RevisionDBID *int       `json:"revision_db_id,omitempty"`
	FilePath     *string    `json:"file_path,omitempty"`
	Line         *int       `json:"line,omitempty"`
	AuthorName   string     `json:"author_name"`
	AuthorEmail  string     `json:"author_email"`
	Message      string     `json:"message"`
	Created      time.Time  `json:"created"`
	Updated      time.Time  `json:"updated"`
	InReplyTo    *string    `json:"in_reply_to,omitempty"`
	Unresolved   bool       `json:"unresolved"`
}

type Label struct {
	ID           int       `json:"id"`
	ChangeDBID   int       `json:"change_db_id"`
	LabelName    string    `json:"label_name"`
	Value        int       `json:"value"`
	AccountName  string    `json:"account_name"`
	AccountEmail string    `json:"account_email"`
	GrantedOn    time.Time `json:"granted_on"`
}

type ChangeMessage struct {
	ID             int       `json:"id"`
	ChangeDBID     int       `json:"change_db_id"`
	MessageID      string    `json:"message_id"`
	AuthorName     string    `json:"author_name"`
	AuthorEmail    string    `json:"author_email"`
	Message        string    `json:"message"`
	Date           time.Time `json:"date"`
	RevisionNumber *int      `json:"revision_number,omitempty"`
}

type Commit struct {
	ID           int       `json:"id"`
	Hash         string    `json:"hash"`
	AuthorName   string    `json:"author_name"`
	AuthorEmail  string    `json:"author_email"`
	CommitDate   time.Time `json:"commit_date"`
	Message      string    `json:"message"`
	ChangeID     *string   `json:"change_id,omitempty"`
	ReviewedOn   *string   `json:"reviewed_on,omitempty"`
	ChangeDBID   *int      `json:"change_db_id,omitempty"`
}

type ChangeWithDetails struct {
	Change    Change          `json:"change"`
	Revisions []Revision      `json:"revisions"`
	Files     []File          `json:"files"`
	Diffs     []Diff          `json:"diffs"`
	Comments  []ChangeComment `json:"comments"`
	Labels    []Label         `json:"labels"`
	Messages  []ChangeMessage `json:"messages"`
}
