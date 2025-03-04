// Package formatter provides functionality for combining file contents
package formatter

// Formatter defines the interface for combining file contents
type Formatter interface {
	// Format combines multiple file contents into a single output
	// files is a map of file paths to their contents
	// Returns the formatted combined content
	Format(files map[string]string) (string, error)
}

// DefaultFormatter implements the standard formatting logic
type DefaultFormatter struct{}

// NewDefaultFormatter creates a new DefaultFormatter
func NewDefaultFormatter() *DefaultFormatter {
	return &DefaultFormatter{}
}

// Format combines multiple file contents with proper section separators
func (f *DefaultFormatter) Format(files map[string]string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	var result string
	first := true

	// Sort files by name to ensure consistent output
	fileNames := make([]string, 0, len(files))
	for fileName := range files {
		fileNames = append(fileNames, fileName)
	}
	sortStrings(fileNames)

	// Combine files with section headers
	for _, fileName := range fileNames {
		content := files[fileName]

		// Add a newline between files
		if !first {
			result += "\n\n"
		}
		first = false

		// Add file path as a section separator using Markdown comment syntax
		result += "[//]: # \"filepath: " + fileName + "\"\n"
		result += content
	}

	return result, nil
}

// sortStrings sorts a slice of strings in place
func sortStrings(s []string) {
	// Simple insertion sort for deterministic ordering
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
