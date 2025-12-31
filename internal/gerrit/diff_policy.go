package gerrit

import (
	"fmt"
	"strings"

	"upgo/internal/models"
)

// DiffPolicy handles diff size limits and filtering
type DiffPolicy struct {
	sizeLimit int // Maximum size in bytes (default: 500KB)
}

// NewDiffPolicy creates a new diff policy
func NewDiffPolicy(sizeLimit int) *DiffPolicy {
	return &DiffPolicy{
		sizeLimit: sizeLimit,
	}
}

// ShouldStoreDiff determines if diff should be stored based on size
func (p *DiffPolicy) ShouldStoreDiff(diffSize int) bool {
	return diffSize <= p.sizeLimit
}

// ProcessDiff processes diff and returns whether to store it
// Returns: (shouldStore, diffText, statsOnly)
func (p *DiffPolicy) ProcessDiff(diffInfo *DiffInfo, filePath string) (bool, string, bool) {
	// Convert diff content to text format
	diffText := p.diffInfoToText(diffInfo)
	diffSize := len([]byte(diffText))

	if p.ShouldStoreDiff(diffSize) {
		return true, diffText, false
	}

	// Size exceeds limit - store stats only
	return false, "", true
}

// diffInfoToText converts DiffInfo to unified diff text format
func (p *DiffPolicy) diffInfoToText(diff *DiffInfo) string {
	var buf strings.Builder

	// Write diff header
	for _, line := range diff.DiffHeader {
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	// Write diff content
	for _, content := range diff.Content {
		// Common lines (ab)
		for _, line := range content.AB {
			buf.WriteString(" ")
			buf.WriteString(line)
			buf.WriteString("\n")
		}

		// Deleted lines (a)
		for _, line := range content.A {
			buf.WriteString("-")
			buf.WriteString(line)
			buf.WriteString("\n")
		}

		// Added lines (b)
		for _, line := range content.B {
			buf.WriteString("+")
			buf.WriteString(line)
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// GetFileDiffStats extracts statistics from FileInfo
func (p *DiffPolicy) GetFileDiffStats(fileInfo *FileInfo) models.ChangeFile {
	return models.ChangeFile{
		FilePath:      "", // Will be set by caller
		Status:        fileInfo.Status,
		OldPath:       fileInfo.OldPath,
		LinesInserted: fileInfo.LinesInserted,
		LinesDeleted:  fileInfo.LinesDeleted,
		SizeDelta:     fileInfo.SizeDelta,
		Size:          fileInfo.Size,
		Binary:        fileInfo.Binary,
	}
}

// FormatDiffForDisplay formats diff text for display
func (p *DiffPolicy) FormatDiffForDisplay(diffText string, maxLines int) string {
	lines := strings.Split(diffText, "\n")
	if len(lines) <= maxLines {
		return diffText
	}

	// Truncate and add indicator
	truncated := strings.Join(lines[:maxLines], "\n")
	return fmt.Sprintf("%s\n... (truncated, %d more lines)", truncated, len(lines)-maxLines)
}
