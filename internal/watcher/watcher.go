// Package watcher provides file monitoring functionality
package watcher

import (
	"context"
	"time"
)

// Event represents a file event
type Event struct {
	FilePath string
	IsRemote bool
}

// FileState represents the state of a file at a point in time
type FileState struct {
	Path    string
	ModTime time.Time
	Exists  bool
}

// Equal compares two FileStates for equality
func (fs FileState) Equal(other FileState) bool {
	return fs.Path == other.Path &&
		fs.ModTime.Equal(other.ModTime) &&
		fs.Exists == other.Exists
}

// FileSystem defines the port for file system operations
type FileSystem interface {
	// GetFileState returns the current state of a file
	GetFileState(path string) (FileState, error)
	// ResolvePath converts a path to its canonical form
	ResolvePath(path string) (string, error)
}

// Watcher defines the interface for file monitoring
type Watcher interface {
	// Watch starts watching the specified files and sends events when they change
	// The ctx parameter can be used to cancel watching
	// The files parameter is a slice of file paths to watch
	// The events channel receives events when files change
	Watch(ctx context.Context, files []string, events chan<- Event) error
	// Close stops watching and cleans up resources
	Close() error
}

// FileChange represents a detected change in a file
type FileChange struct {
	Path    string
	IsNew   bool
	IsError bool
}

// CheckFiles compares current and previous file states to detect changes
// This is a pure function that can be easily tested
func CheckFiles(current, previous map[string]FileState) []FileChange {
	changes := make([]FileChange, 0)

	// Check for modified or new files
	for path, currentState := range current {
		if prevState, exists := previous[path]; exists {
			if !currentState.Equal(prevState) {
				changes = append(changes, FileChange{Path: path, IsNew: false})
			}
		} else {
			changes = append(changes, FileChange{Path: path, IsNew: true})
		}
	}

	// Check for error states (files that no longer exist)
	for path := range previous {
		if _, exists := current[path]; !exists {
			changes = append(changes, FileChange{Path: path, IsError: true})
		}
	}

	return changes
}

// CreateEvents converts file changes to events
// This is a pure function that can be easily tested
func CreateEvents(changes []FileChange, isRemote bool) []Event {
	events := make([]Event, 0, len(changes))
	for _, change := range changes {
		if !change.IsError { // Only create events for valid changes
			events = append(events, Event{
				FilePath: change.Path,
				IsRemote: isRemote,
			})
		}
	}
	return events
}

// GetFileStates retrieves the current state of multiple files
// This function isolates the impure file system operations
func GetFileStates(fs FileSystem, paths []string) (map[string]FileState, error) {
	states := make(map[string]FileState)

	for _, path := range paths {
		resolvedPath, err := fs.ResolvePath(path)
		if err != nil {
			return nil, err
		}

		state, err := fs.GetFileState(resolvedPath)
		if err != nil {
			return nil, err
		}

		states[resolvedPath] = state
	}

	return states, nil
}
