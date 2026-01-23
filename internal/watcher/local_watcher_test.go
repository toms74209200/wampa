//go:build small

package watcher

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestCheckFiles tests the pure function that detects file changes
func TestCheckFiles(t *testing.T) {
	baseTime := time.Now()
	laterTime := baseTime.Add(time.Second)

	tests := []struct {
		name     string
		current  map[string]FileState
		previous map[string]FileState
		want     []FileChange
	}{
		{
			name: "no changes",
			current: map[string]FileState{
				"file1": {Path: "file1", ModTime: baseTime, Exists: true},
			},
			previous: map[string]FileState{
				"file1": {Path: "file1", ModTime: baseTime, Exists: true},
			},
			want: []FileChange{},
		},
		{
			name: "file modified",
			current: map[string]FileState{
				"file1": {Path: "file1", ModTime: laterTime, Exists: true},
			},
			previous: map[string]FileState{
				"file1": {Path: "file1", ModTime: baseTime, Exists: true},
			},
			want: []FileChange{
				{Path: "file1", IsNew: false},
			},
		},
		{
			name: "new file",
			current: map[string]FileState{
				"file1": {Path: "file1", ModTime: baseTime, Exists: true},
				"file2": {Path: "file2", ModTime: baseTime, Exists: true},
			},
			previous: map[string]FileState{
				"file1": {Path: "file1", ModTime: baseTime, Exists: true},
			},
			want: []FileChange{
				{Path: "file2", IsNew: true},
			},
		},
		{
			name: "file removed",
			current: map[string]FileState{
				"file1": {Path: "file1", ModTime: baseTime, Exists: true},
			},
			previous: map[string]FileState{
				"file1": {Path: "file1", ModTime: baseTime, Exists: true},
				"file2": {Path: "file2", ModTime: baseTime, Exists: true},
			},
			want: []FileChange{
				{Path: "file2", IsError: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckFiles(tt.current, tt.previous)
			if !compareFileChanges(got, tt.want) {
				t.Errorf("CheckFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCreateEvents tests the pure function that creates events from changes
func TestCreateEvents(t *testing.T) {
	tests := []struct {
		name     string
		changes  []FileChange
		isRemote bool
		want     []Event
	}{
		{
			name:     "no changes",
			changes:  []FileChange{},
			isRemote: false,
			want:     []Event{},
		},
		{
			name: "single change",
			changes: []FileChange{
				{Path: "file1", IsNew: false},
			},
			isRemote: false,
			want: []Event{
				{FilePath: "file1", IsRemote: false},
			},
		},
		{
			name: "multiple changes",
			changes: []FileChange{
				{Path: "file1", IsNew: false},
				{Path: "file2", IsNew: true},
			},
			isRemote: true,
			want: []Event{
				{FilePath: "file1", IsRemote: true},
				{FilePath: "file2", IsRemote: true},
			},
		},
		{
			name: "ignore error changes",
			changes: []FileChange{
				{Path: "file1", IsNew: false},
				{Path: "file2", IsError: true},
			},
			isRemote: false,
			want: []Event{
				{FilePath: "file1", IsRemote: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateEvents(tt.changes, tt.isRemote)
			if !compareEvents(got, tt.want) {
				t.Errorf("CreateEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}

// MockFileSystem implements FileSystem interface for testing
type MockFileSystem struct {
	mu     sync.Mutex
	states map[string]FileState
	errors map[string]error
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		states: make(map[string]FileState),
		errors: make(map[string]error),
	}
}

func (m *MockFileSystem) GetFileState(path string) (FileState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// First check for simulated errors
	if err, ok := m.errors[path]; ok && err != nil {
		return FileState{}, err
	}

	// Then check for existing state
	if state, ok := m.states[path]; ok {
		return state, nil
	}

	// If path doesn't exist, return non-existent state
	return FileState{Path: path, Exists: false}, nil
}

func (m *MockFileSystem) ResolvePath(path string) (string, error) {
	if err, ok := m.errors[path]; ok && err != nil {
		return "", err
	}
	return "/mock/" + path, nil
}

func (m *MockFileSystem) SetFileState(path string, state FileState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[path] = state
}

func (m *MockFileSystem) SetError(path string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[path] = err
}

// TestLocalWatcher tests the watcher implementation
func TestLocalWatcher(t *testing.T) {
	testCases := []struct {
		name        string
		setupFS     func(*MockFileSystem)
		actions     func(*MockFileSystem)
		wantEvents  int
		wantTimeout bool
		wantError   bool
	}{
		{
			name: "detect single file modification",
			setupFS: func(fs *MockFileSystem) {
				fs.SetFileState("/mock/test.txt", FileState{
					Path:    "/mock/test.txt",
					ModTime: time.Now(),
					Exists:  true,
				})
			},
			actions: func(fs *MockFileSystem) {
				fs.SetFileState("/mock/test.txt", FileState{
					Path:    "/mock/test.txt",
					ModTime: time.Now().Add(time.Second),
					Exists:  true,
				})
			},
			wantEvents:  1,
			wantTimeout: false,
			wantError:   false,
		},
		{
			name: "detect multiple file modifications",
			setupFS: func(fs *MockFileSystem) {
				now := time.Now()
				fs.SetFileState("/mock/test1.txt", FileState{
					Path:    "/mock/test1.txt",
					ModTime: now,
					Exists:  true,
				})
				fs.SetFileState("/mock/test2.txt", FileState{
					Path:    "/mock/test2.txt",
					ModTime: now,
					Exists:  true,
				})
			},
			actions: func(fs *MockFileSystem) {
				later := time.Now().Add(time.Second)
				fs.SetFileState("/mock/test1.txt", FileState{
					Path:    "/mock/test1.txt",
					ModTime: later,
					Exists:  true,
				})
				fs.SetFileState("/mock/test2.txt", FileState{
					Path:    "/mock/test2.txt",
					ModTime: later,
					Exists:  true,
				})
			},
			wantEvents:  2,
			wantTimeout: false,
			wantError:   false,
		},
		{
			name: "handle file resolve error",
			setupFS: func(fs *MockFileSystem) {
				fs.SetError("test.txt", errors.New("resolve error"))
			},
			actions:     func(fs *MockFileSystem) {},
			wantEvents:  0,
			wantTimeout: false,
			wantError:   true,
		},
		{
			name: "handle get state error",
			setupFS: func(fs *MockFileSystem) {
				fs.SetError("/mock/test.txt", errors.New("state error"))
			},
			actions:     func(fs *MockFileSystem) {},
			wantEvents:  0,
			wantTimeout: false,
			wantError:   true,
		},
		{
			name: "handle watching already started error",
			setupFS: func(fs *MockFileSystem) {
				fs.SetFileState("/mock/test.txt", FileState{
					Path:    "/mock/test.txt",
					ModTime: time.Now(),
					Exists:  true,
				})
			},
			actions:     func(fs *MockFileSystem) {},
			wantEvents:  0,
			wantTimeout: false,
			wantError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockFS := NewMockFileSystem()
			tc.setupFS(mockFS)

			w := &LocalWatcher{
				fs:         mockFS,
				states:     make(map[string]FileState),
				done:       make(chan struct{}),
				pollPeriod: 10 * time.Millisecond,
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			events := make(chan Event, tc.wantEvents+1) // +1 for potential extra events
			files := []string{"test.txt"}
			if tc.wantEvents > 1 {
				files = []string{"test1.txt", "test2.txt"}
			}

			// For "watching already started" test
			if tc.wantError && tc.name == "handle watching already started error" {
				w.watching = true
			}

			err := w.Watch(ctx, files, events)
			if tc.wantError {
				if err == nil {
					t.Error("Watch() expected error but got nil")
				}
				return
			} else if err != nil {
				t.Fatalf("Watch() error = %v", err)
			}

			// Wait for initial setup
			time.Sleep(20 * time.Millisecond)

			// Perform test actions
			tc.actions(mockFS)

			// Check for events
			receivedCount := 0
			timeout := time.After(100 * time.Millisecond)

			for receivedCount < tc.wantEvents {
				select {
				case <-events:
					receivedCount++
				case <-timeout:
					if !tc.wantTimeout {
						t.Errorf("Timeout waiting for events, got %d, want %d", receivedCount, tc.wantEvents)
					}
					return
				}
			}

			// Test double close
			if err := w.Close(); err != nil {
				t.Errorf("First Close() error = %v", err)
			}
			if err := w.Close(); err != nil {
				t.Errorf("Second Close() error = %v", err)
			}
		})
	}
}

// Helper functions for comparing test results
func compareFileChanges(a, b []FileChange) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func compareEvents(a, b []Event) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
