package tui

import (
	"fmt"
)

// DocFileType represents the classification of a file.
type DocFileType int

const (
	TypeDoc DocFileType = iota
	TypeCode
	TypeIgnored // Files that are explicitly ignored by the user or default rules
	TypeOther   // Files that don't fit into doc or code (e.g., config files, images)
)

// FileInfo holds information about a discovered file.
type FileInfo struct {
	Path       string
	Type       DocFileType
	Size       int64
	IsSelected bool   // For interactive selection
	IsIgnored  bool   // For interactive ignore
	LLMSummary string // LLM generated description
}

// FilterValue implements list.Item.
func (f FileInfo) FilterValue() string { return f.Path }

// Title implements list.Item.
func (f FileInfo) Title() string {
	if f.IsIgnored {
		return "[ ] " + f.Path
	}
	return "[x] " + f.Path
}

// Description implements list.Item.
func (f FileInfo) Description() string {
	return fmt.Sprintf("Size: %d bytes", f.Size)
}

// ProviderItem adapts provider strings to list.Item
type ProviderItem struct {
	Name string
	Desc string
}

func (p ProviderItem) Title() string       { return p.Name }
func (p ProviderItem) Description() string { return p.Desc }
func (p ProviderItem) FilterValue() string { return p.Name }
