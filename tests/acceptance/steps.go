package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"wampa/pkg/formatter"
	"wampa/pkg/watcher"

	"github.com/cucumber/godog"
)

type testContext struct {
	dir        string
	outputPath string
	watcher    *watcher.LocalWatcher
	formatter  *formatter.DefaultFormatter
	events     chan watcher.Event
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

func newTestContext() *testContext {
	dir, err := os.MkdirTemp("", "wampa-test-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to create test directory: %v", err))
	}
	w, err := watcher.NewLocalWatcher()
	if err != nil {
		panic(fmt.Sprintf("Failed to create watcher: %v", err))
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &testContext{
		dir:       dir,
		watcher:   w,
		formatter: formatter.NewDefaultFormatter(),
		events:    make(chan watcher.Event, 10),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (tc *testContext) cleanup() {
	if tc.cancel != nil {
		tc.cancel()
	}
	tc.wg.Wait() // goroutineの終了を待つ
	if tc.watcher != nil {
		tc.watcher.Close()
	}
	if tc.dir != "" {
		os.RemoveAll(tc.dir)
	}
}

func (tc *testContext) thereIsFileWithContent(filename, content string) error {
	path := filepath.Join(tc.dir, filename)
	return os.WriteFile(path, []byte(content), 0644)
}

func (tc *testContext) checkFileContent(path, want string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", path, err)
	}
	if string(content) != want {
		return fmt.Errorf("unexpected content.\nWant:\n%s\nGot:\n%s", want, string(content))
	}
	return nil
}

func (tc *testContext) waitForContent(want string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := tc.checkFileContent(tc.outputPath, want); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for content: %v", lastErr)
}

func (tc *testContext) executeWampaCommand(command string) error {
	// コマンドを分割して引数を取得
	args := strings.Fields(command)
	if len(args) < 1 || args[0] != "wampa" {
		return fmt.Errorf("invalid command: %s", command)
	}

	// -i と -o オプションを解析
	var inputFiles []string
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-i":
			// -i の後の引数をすべて入力ファイルとして扱う
			i++
			for ; i < len(args) && !strings.HasPrefix(args[i], "-"); i++ {
				inputFiles = append(inputFiles, filepath.Join(tc.dir, args[i]))
			}
			i-- // for文のi++で次のイテレーションに進むため、ここで戻す
		case "-o":
			if i+1 >= len(args) {
				return fmt.Errorf("missing output file path")
			}
			tc.outputPath = filepath.Join(tc.dir, args[i+1])
			i++
		}
	}

	if len(inputFiles) == 0 {
		return fmt.Errorf("no input files specified")
	}
	if tc.outputPath == "" {
		return fmt.Errorf("no output file specified")
	}

	// 初期ファイルの生成
	contents := make(map[string]string)
	for _, file := range inputFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read input file %s: %v", file, err)
		}
		contents[file] = string(content)
	}

	// 初期出力ファイルの生成
	formatted, err := tc.formatter.Format(inputFiles, contents)
	if err != nil {
		return fmt.Errorf("failed to format files: %v", err)
	}
	if err := os.WriteFile(tc.outputPath, []byte(formatted), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %v", err)
	}

	// ファイル監視を開始
	tc.wg.Add(2) // 2つのgoroutineを追加
	go func() {
		defer tc.wg.Done()
		if err := tc.watcher.Watch(tc.ctx, inputFiles, tc.events); err != nil {
			fmt.Printf("Failed to watch files: %v\n", err)
			return
		}
	}()

	// イベントを処理して出力ファイルを更新
	go func() {
		defer tc.wg.Done()
		for {
			select {
			case <-tc.ctx.Done():
				return
			case <-tc.events:
				// ファイルの内容を読み込み
				contents := make(map[string]string)
				for _, file := range inputFiles {
					content, err := os.ReadFile(file)
					if err != nil {
						fmt.Printf("Failed to read file %s: %v\n", file, err)
						continue
					}
					contents[file] = string(content)
				}
				// フォーマットして保存
				formatted, err := tc.formatter.Format(inputFiles, contents)
				if err != nil {
					fmt.Printf("Failed to format files: %v\n", err)
					continue
				}
				if err := os.WriteFile(tc.outputPath, []byte(formatted), 0644); err != nil {
					fmt.Printf("Failed to write output file: %v\n", err)
				}
			}
		}
	}()

	// 初期ファイルが生成されるまで少し待機
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (tc *testContext) outputFileContains(content string) error {
	return tc.checkFileContent(tc.outputPath, content)
}

func (tc *testContext) outputFileUpdatedWithin5Seconds(content string) error {
	return tc.waitForContent(content, 5*time.Second)
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	testCtx := newTestContext()
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		testCtx.cleanup()
		return ctx, nil
	})

	ctx.Step(`^以下の内容の([^"]*)が存在する:$`, testCtx.thereIsFileWithContent)
	ctx.Step(`^wampaを以下のコマンドで実行:$`, testCtx.executeWampaCommand)
	ctx.Step(`^output\.mdは以下の内容を含む:$`, testCtx.outputFileContains)
	ctx.Step(`^([^"]*)を以下の内容に変更:$`, testCtx.thereIsFileWithContent)
	ctx.Step(`^5秒以内にoutput\.mdは以下の内容に更新される:$`, testCtx.outputFileUpdatedWithin5Seconds)
}
