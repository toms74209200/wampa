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
	"wampa/pkg/wampa"
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
	cmdDone    chan struct{} // コマンド実行完了通知用のチャネル
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
		cmdDone:   make(chan struct{}), // チャネルを初期化
	}
}

func (tc *testContext) cleanup() {
	if tc.cancel != nil {
		tc.cancel()
	}
	// コマンド完了を待機
	select {
	case <-tc.cmdDone:
		// 正常に終了した場合
	case <-time.After(2 * time.Second):
		// タイムアウトした場合
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

	// wampaコマンドの部分を除外
	cmdArgs := []string{}

	// 引数の変換: 相対パスをテストディレクトリの絶対パスに変換
	i := 1 // wampa の後から処理開始
	for i < len(args) {
		switch args[i] {
		case "-i":
			// 入力ファイル指定
			cmdArgs = append(cmdArgs, "-i")
			i++
			// 複数のファイルが指定されている可能性があるため、次の"-"で始まるオプションが来るまで処理
			for i < len(args) && !strings.HasPrefix(args[i], "-") {
				// 相対パスを絶対パスに変換
				absPath := filepath.Join(tc.dir, args[i])
				cmdArgs = append(cmdArgs, absPath)
				i++
			}
		case "-o":
			// 出力ファイル指定
			cmdArgs = append(cmdArgs, "-o")
			i++
			if i < len(args) {
				// 出力ファイルのパスを絶対パスに変換
				tc.outputPath = filepath.Join(tc.dir, args[i])
				cmdArgs = append(cmdArgs, tc.outputPath)
				i++
			}
		default:
			// その他のオプションはそのままコピー
			cmdArgs = append(cmdArgs, args[i])
			i++
		}
	}

	// デバッグ用
	fmt.Printf("処理後の引数: %v\n", cmdArgs)
	fmt.Printf("出力ファイル: %s\n", tc.outputPath)

	// 別goroutineでRun関数を実行
	tc.wg.Add(1)
	go func() {
		defer tc.wg.Done()
		defer close(tc.cmdDone) // 実行完了を通知

		// run関数を実行してwampaコマンドを実行
		err := wampa.Run(tc.ctx, cmdArgs)
		if err != nil {
			fmt.Printf("Failed to execute wampa command: %v\n", err)
		}
	}()

	// コマンド開始待機
	time.Sleep(500 * time.Millisecond)
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
