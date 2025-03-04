package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// RealFileSystem implements FileSystem interface using actual OS operations
type RealFileSystem struct{}

func (fs *RealFileSystem) GetFileState(path string) (FileState, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return FileState{Path: path, Exists: false}, nil
		}
		return FileState{}, err
	}
	return FileState{
		Path:    path,
		ModTime: info.ModTime(),
		Exists:  true,
	}, nil
}

func (fs *RealFileSystem) ResolvePath(path string) (string, error) {
	return filepath.Abs(path)
}

// LocalWatcher implements Watcher for local files
type LocalWatcher struct {
	mu         sync.Mutex
	fs         FileSystem
	states     map[string]FileState
	watching   bool
	done       chan struct{}
	pollPeriod time.Duration
}

// NewLocalWatcher creates a new LocalWatcher instance
func NewLocalWatcher() (*LocalWatcher, error) {
	return &LocalWatcher{
		fs:         &RealFileSystem{},
		states:     make(map[string]FileState),
		done:       make(chan struct{}),
		pollPeriod: 100 * time.Millisecond,
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

	// Get initial states
	initialStates, err := GetFileStates(w.fs, files)
	if err != nil {
		w.mu.Lock()
		w.watching = false
		w.mu.Unlock()
		return fmt.Errorf("failed to get initial file states: %w", err)
	}

	w.mu.Lock()
	w.states = initialStates
	w.mu.Unlock()

	// Start polling goroutine
	go w.poll(ctx, events)

	return nil
}

// poll periodically checks for file changes
func (w *LocalWatcher) poll(ctx context.Context, events chan<- Event) {
	ticker := time.NewTicker(w.pollPeriod)
	defer ticker.Stop()
	defer func() {
		w.mu.Lock()
		w.watching = false
		w.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		case <-ticker.C:
			if err := w.checkChanges(events); err != nil {
				fmt.Printf("Error checking changes: %v\n", err)
			}
		}
	}
}

// checkChanges checks for file changes and sends events
func (w *LocalWatcher) checkChanges(events chan<- Event) error {
	w.mu.Lock()
	paths := make([]string, 0, len(w.states))
	previousStates := make(map[string]FileState, len(w.states))
	for path, state := range w.states {
		paths = append(paths, path)
		previousStates[path] = state
	}
	w.mu.Unlock()

	// Get current states
	currentStates, err := GetFileStates(w.fs, paths)
	if err != nil {
		return fmt.Errorf("failed to get file states: %w", err)
	}

	// Check for changes
	changes := CheckFiles(currentStates, previousStates)
	if len(changes) > 0 {
		// Create and send events
		fileEvents := CreateEvents(changes, false)

		// Update states before sending events to prevent race conditions
		w.mu.Lock()
		w.states = currentStates
		w.mu.Unlock()

		// Send events without holding the lock
		for _, event := range fileEvents {
			select {
			case events <- event:
				// Event sent successfully
			default:
				return fmt.Errorf("event channel is blocked")
			}
		}
	}

	return nil
}

// Close stops watching and cleans up resources
func (w *LocalWatcher) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.watching {
		return nil
	}

	close(w.done)
	w.watching = false
	return nil
}
