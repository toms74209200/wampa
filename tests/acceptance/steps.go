package acceptance

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/toms74209200/wampa/pkg/formatter"
	"github.com/toms74209200/wampa/pkg/wampa"
	"github.com/toms74209200/wampa/pkg/watcher"

	"github.com/cucumber/godog"
)

type testContext struct {
	dir           string
	outputPath    string
	watcher       *watcher.LocalWatcher
	formatter     *formatter.DefaultFormatter
	events        chan watcher.Event
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	cmdDone       chan struct{} // コマンド実行完了通知用のチャネル
	cmdError      error         // コマンド実行のエラー結果
	stdoutCapture *bytes.Buffer // 標準出力をキャプチャするバッファ
	stderrCapture *bytes.Buffer // 標準エラー出力をキャプチャするバッファ
	origStdout    *os.File      // 元の標準出力
	origStderr    *os.File      // 元の標準エラー出力
	stdoutReader  *os.File      // 標準出力のリーダー
	stdoutWriter  *os.File      // 標準出力のライター
	stderrReader  *os.File      // 標準エラー出力のリーダー
	stderrWriter  *os.File      // 標準エラー出力のライター
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
		dir:           dir,
		watcher:       w,
		formatter:     formatter.NewDefaultFormatter(),
		events:        make(chan watcher.Event, 10),
		ctx:           ctx,
		cancel:        cancel,
		cmdDone:       make(chan struct{}), // チャネルを初期化
		stdoutCapture: &bytes.Buffer{},
		stderrCapture: &bytes.Buffer{},
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
	// 標準出力と標準エラー出力を元に戻す
	tc.restoreStdoutAndStderr()
	if tc.watcher != nil {
		tc.watcher.Close()
	}
	if tc.dir != "" {
		os.RemoveAll(tc.dir)
	}
}

// 標準出力と標準エラー出力をキャプチャするための準備
func (tc *testContext) captureStdoutAndStderr() {
	// 標準出力のキャプチャ
	tc.origStdout = os.Stdout
	tc.stdoutReader, tc.stdoutWriter, _ = os.Pipe()
	os.Stdout = tc.stdoutWriter
	// 標準エラー出力のキャプチャ
	tc.origStderr = os.Stderr
	tc.stderrReader, tc.stderrWriter, _ = os.Pipe()
	os.Stderr = tc.stderrWriter
	// キャプチャ開始
	tc.stdoutCapture.Reset()
	tc.stderrCapture.Reset()
	// 別goroutineでパイプから読み取り
	tc.wg.Add(2)
	go func() {
		defer tc.wg.Done()
		io.Copy(tc.stdoutCapture, tc.stdoutReader)
	}()
	go func() {
		defer tc.wg.Done()
		io.Copy(tc.stderrCapture, tc.stderrReader)
	}()
}

// 標準出力と標準エラー出力を元に戻す
func (tc *testContext) restoreStdoutAndStderr() {
	if tc.stdoutWriter != nil {
		tc.stdoutWriter.Close()
		tc.stdoutWriter = nil
	}
	if tc.stderrWriter != nil {
		tc.stderrWriter.Close()
		tc.stderrWriter = nil
	}
	if tc.origStdout != nil {
		os.Stdout = tc.origStdout
		tc.origStdout = nil
	}
	if tc.origStderr != nil {
		os.Stderr = tc.origStderr
		tc.origStderr = nil
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
	// 標準出力と標準エラー出力をキャプチャ
	tc.captureStdoutAndStderr()
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
		tc.cmdError = wampa.Run(tc.ctx, cmdArgs)
		// パイプを閉じて、キャプチャ用goroutineが終了できるようにする
		tc.stdoutWriter.Close()
		tc.stderrWriter.Close()
	}()
	// コマンド開始待機
	time.Sleep(500 * time.Millisecond)
	return nil
}

// カレントディレクトリにwampa.tomlが存在しない状態でコマンドを実行
func (tc *testContext) executeWampaCommandWithoutConfig(command string) error {
	// wampa.tomlが存在しないことを確認
	configPath := filepath.Join(tc.dir, "wampa.toml")
	if _, err := os.Stat(configPath); err == nil {
		os.Remove(configPath)
	}
	return tc.executeWampaCommand(command)
}

func (tc *testContext) outputFileContains(content string) error {
	return tc.checkFileContent(tc.outputPath, content)
}

func (tc *testContext) outputFileUpdatedWithin5Seconds(content string) error {
	return tc.waitForContent(content, 5*time.Second)
}

func (tc *testContext) outputFileDoesNotExist(filename string) error {
	path := filepath.Join(tc.dir, filename)
	_, err := os.Stat(path)
	if err == nil {
		return fmt.Errorf("file %s exists but should not", filename)
	}
	if !os.IsNotExist(err) {
		return fmt.Errorf("unexpected error checking for file existence: %v", err)
	}
	return nil
}

// 以下のヘルプメッセージが表示されることを確認
func (tc *testContext) helpMessageIsDisplayed(expected string) error {
	// キャプチャした標準出力を取得
	capturedOutput := tc.stdoutCapture.String()
	// メッセージが標準出力に含まれているか確認
	if !strings.Contains(capturedOutput, expected) {
		return fmt.Errorf("help message not found in output.\nExpected to contain:\n%s\nGot:\n%s",
			expected, capturedOutput)
	}
	return nil
}

// 以下のエラーメッセージが表示されることを確認
func (tc *testContext) errorMessageIsDisplayed(expected string) error {
	// キャプチャした標準エラー出力を取得
	capturedError := tc.stderrCapture.String()
	// メッセージが標準エラー出力に含まれているか確認
	if !strings.Contains(capturedError, expected) {
		return fmt.Errorf("error message not found in stderr.\nExpected to contain:\n%s\nGot:\n%s",
			expected, capturedError)
	}
	return nil
}

// プロセスがゼロの終了コードで終了することを確認
func (tc *testContext) processExitsWithZeroCode() error {
	// コマンドの実行が正常に終了したか確認
	if tc.cmdError != nil {
		return fmt.Errorf("process exited with error: %v", tc.cmdError)
	}
	return nil
}

// プロセスが非ゼロの終了コードで終了することを確認
func (tc *testContext) processExitsWithNonZeroCode() error {
	// コマンドの実行がエラーで終了したか確認
	if tc.cmdError == nil {
		return fmt.Errorf("process exited with zero code but expected non-zero")
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	testCtx := newTestContext()
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		testCtx.cleanup()
		return ctx, nil
	})
	ctx.Step(`^以下の内容の([^"]*)が存在する:$`, testCtx.thereIsFileWithContent)
	ctx.Step(`^wampaを以下のコマンドで実行:$`, testCtx.executeWampaCommand)
	ctx.Step(`^カレントディレクトリにwampa\.tomlが存在しない状態でwampaをパラメータなしで実行:$`, testCtx.executeWampaCommandWithoutConfig)
	ctx.Step(`^カレントディレクトリにwampa\.tomlが存在しない状態でwampaを以下のコマンドで実行:$`, testCtx.executeWampaCommandWithoutConfig)
	ctx.Step(`^output\.mdは以下の内容を含む:$`, testCtx.outputFileContains)
	ctx.Step(`^([^"]*)は以下の内容を含む:$`, testCtx.outputFileContains)
	ctx.Step(`^([^"]*)は作成されない$`, testCtx.outputFileDoesNotExist)
	ctx.Step(`^([^"]*)を以下の内容に変更:$`, testCtx.thereIsFileWithContent)
	ctx.Step(`^5秒以内にoutput\.mdは以下の内容に更新される:$`, testCtx.outputFileUpdatedWithin5Seconds)
	ctx.Step(`^以下のヘルプメッセージが表示される:$`, testCtx.helpMessageIsDisplayed)
	ctx.Step(`^以下のエラーメッセージが表示される:$`, testCtx.errorMessageIsDisplayed)
	ctx.Step(`^プロセスはゼロの終了コードで終了する$`, testCtx.processExitsWithZeroCode)
	ctx.Step(`^プロセスは非ゼロの終了コードで終了する$`, testCtx.processExitsWithNonZeroCode)
}
