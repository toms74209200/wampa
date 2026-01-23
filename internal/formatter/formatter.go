// Package formatter provides functionality for combining file contents
package formatter

import (
	"path/filepath"
)

// Formatter defines the interface for combining file contents
type Formatter interface {
	// Format combines multiple file contents into a single output
	// files is a slice of paths, and contents is a map of paths to their contents
	// Returns the formatted combined content
	Format(files []string, contents map[string]string) (string, error)
}

// DefaultFormatter implements the standard formatting logic
type DefaultFormatter struct{}

// NewDefaultFormatter creates a new DefaultFormatter
func NewDefaultFormatter() *DefaultFormatter {
	return &DefaultFormatter{}
}

// Format combines multiple file contents with proper section separators
func (f *DefaultFormatter) Format(files []string, contents map[string]string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	// 各ファイルの内容を結合（指定された順序を維持）
	var parts []string
	for _, file := range files {
		content, ok := contents[file]
		if !ok {
			continue
		}
		// 相対パスに変換
		relPath := filepath.Base(file)
		parts = append(parts,
			`[//]: # "filepath: `+relPath+`"`+"\n"+
				content)
	}

	return joinParts(parts), nil
}

// joinParts joins parts with double newlines
func joinParts(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += "\n\n" + parts[i]
	}
	return result
}
