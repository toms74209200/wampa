// Package watcher provides file monitoring functionality
package watcher

import (
	"context"
)

// Event represents a file event
type Event struct {
	FilePath string
	IsRemote bool
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