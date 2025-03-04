//go:build small

package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// mockFsnotify は単体テスト用のfsnotifyモック
type mockFsnotify struct {
	events chan fsnotifyEvent
	errors chan error
	files  []string
	closed bool
	mu     sync.Mutex
}

type fsnotifyEvent struct {
	name string
	op   uint32 // fsnotify.Op と互換性を持たせる
}

// NewWatcher は新しいモックウォッチャーを作成
func (m *mockFsnotify) NewWatcher() (*mockFsnotify, error) {
	return &mockFsnotify{
		events: make(chan fsnotifyEvent),
		errors: make(chan error),
		files:  []string{},
		closed: false,
	}, nil
}

// Add はファイルをウォッチリストに追加
func (m *mockFsnotify) Add(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files = append(m.files, name)
	return nil
}

// Close はウォッチャーを閉じる
func (m *mockFsnotify) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		m.closed = true
		close(m.events)
		close(m.errors)
	}
	return nil
}

// SimulateEvent は指定したイベントをシミュレート
func (m *mockFsnotify) SimulateEvent(name string, op uint32) {
	m.events <- fsnotifyEvent{name: name, op: op}
}

// TestLocalWatcher_Watch は小さな単位でローカルファイル監視をテスト
func TestLocalWatcher_Watch(t *testing.T) {
	// テストケース
	tests := []struct {
		name     string
		files    []string
		simulate func(*testing.T, *mockFsnotify)
		want     int // 受信を期待するイベント数
		wantErr  bool
	}{
		{
			name:  "シングルファイル監視",
			files: []string{"test.md"},
			simulate: func(t *testing.T, m *mockFsnotify) {
				m.SimulateEvent("test.md", 2) // Write操作を模倣
			},
			want:    1,
			wantErr: false,
		},
		{
			name:  "複数ファイル監視",
			files: []string{"test1.md", "test2.md"},
			simulate: func(t *testing.T, m *mockFsnotify) {
				m.SimulateEvent("test1.md", 2) // Write操作を模倣
				m.SimulateEvent("test2.md", 2) // Write操作を模倣
			},
			want:    2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックの設定
			mockFs := &mockFsnotify{}
			mockWatcher, _ := mockFs.NewWatcher()

			// コンテキストとチャネルの設定
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			events := make(chan Event)
			done := make(chan struct{})

			// イベント受信用ゴルーチン
			receivedEvents := 0
			go func() {
				for range events {
					receivedEvents++
					if receivedEvents >= tt.want {
						close(done)
						return
					}
				}
			}()

			// テスト対象の関数へのインターフェース呼び出し
			// 実際の実装は使わずにモックを使う
			err := performWatch(ctx, mockWatcher, tt.files, events)
			if (err != nil) != tt.wantErr {
				t.Errorf("Watch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// イベントシミュレーション
			if tt.simulate != nil {
				tt.simulate(t, mockWatcher)
			}

			// 期待するイベント数が受信されるまで待機、またはタイムアウト
			select {
			case <-done:
				// 正常に完了
			case <-time.After(100 * time.Millisecond):
				t.Errorf("期待されるイベント数 %d に達しなかった, 実際: %d", tt.want, receivedEvents)
			}

			// クリーンアップ
			cancel()
			mockWatcher.Close()
		})
	}
}

// performWatch はテスト用のヘルパー関数
func performWatch(ctx context.Context, watcher *mockFsnotify, files []string, events chan<- Event) error {
	for _, file := range files {
		if err := watcher.Add(file); err != nil {
			return err
		}
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.events:
				if !ok {
					return
				}
				if event.op == 2 { // Write操作を想定
					events <- Event{
						FilePath: event.name,
						IsRemote: false,
					}
				}
			case _, ok := <-watcher.errors:
				if !ok {
					return
				}
			}
		}
	}()

	return nil
}

// 統合テスト用の実際のファイルを使ったテスト
// Small Testではないためビルドタグで除外する
func TestLocalWatcher_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// テスト用の一時ファイルを作成
	tempFile, err := os.CreateTemp("", "watcher_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// テスト対象の初期化
	watcher, err := NewLocalWatcher()
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// コンテキストとチャネルの設定
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan Event)

	// 監視開始
	if err := watcher.Watch(ctx, []string{tempFile.Name()}, events); err != nil {
		t.Fatalf("Failed to start watching: %v", err)
	}

	// イベント受信用ゴルーチン
	var receivedPath string
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		select {
		case event := <-events:
			receivedPath = event.FilePath
		case <-time.After(2 * time.Second):
			t.Errorf("Timeout waiting for file change event")
		}
	}()

	// ファイル変更イベントをトリガー
	time.Sleep(100 * time.Millisecond) // 監視が開始されるまで少し待つ
	if _, err := tempFile.WriteString("test content"); err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}
	if err := tempFile.Sync(); err != nil {
		t.Fatalf("Failed to sync file: %v", err)
	}

	// 結果を確認
	wg.Wait()

	// 絶対パスに変換して比較
	absPath, _ := filepath.Abs(tempFile.Name())
	if receivedPath != absPath && receivedPath != tempFile.Name() {
		t.Errorf("Unexpected file path: got %v, want %v or %v", receivedPath, absPath, tempFile.Name())
	}
}
