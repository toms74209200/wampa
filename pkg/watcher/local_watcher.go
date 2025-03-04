package watcher

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// LocalWatcher implements Watcher for local files
type LocalWatcher struct {
	fsWatcher *fsnotify.Watcher
	mu        sync.Mutex
	watching  bool
}

// NewLocalWatcher creates a new LocalWatcher instance
func NewLocalWatcher() (*LocalWatcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create fsnotify watcher: %w", err)
	}
	
	return &LocalWatcher{
		fsWatcher: fsWatcher,
		watching:  false,
	}, nil
}

// Watch starts watching the specified local files
func (w *LocalWatcher) Watch(ctx context.Context, files []string, events chan<- Event) error {
	w.mu.Lock()
	if w.watching {
		w.mu.Unlock()
		return fmt.Errorf("watcher is already watching")
	}
	w.watching = true
	w.mu.Unlock()
	
	// Add all files to watch
	for _, file := range files {
		// Get the absolute path
		absPath, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", file, err)
		}
		
		// Add the file to the watcher
		if err := w.fsWatcher.Add(absPath); err != nil {
			return fmt.Errorf("failed to watch file %s: %w", absPath, err)
		}
	}
	
	// Start the watching goroutine
	go func() {
		defer func() {
			w.mu.Lock()
			w.watching = false
			w.mu.Unlock()
		}()
		
		for {
			select {
			case <-ctx.Done():
				// Context was cancelled
				return
				
			case event, ok := <-w.fsWatcher.Events:
				if !ok {
					// Channel was closed
					return
				}
				
				// Check if this is a modification event
				if event.Op&fsnotify.Write == fsnotify.Write {
					// Send an event to the channel
					events <- Event{
						FilePath: event.Name,
						IsRemote: false,
					}
				}
				
			case err, ok := <-w.fsWatcher.Errors:
				if !ok {
					// Channel was closed
					return
				}
				// Log the error but continue watching
				fmt.Printf("Error watching file: %v\n", err)
			}
		}
	}()
	
	return nil
}

// Close stops watching and cleans up resources
func (w *LocalWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if !w.watching {
		return nil
	}
	
	w.watching = false
	return w.fsWatcher.Close()
}