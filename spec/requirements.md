# Wampa Technical Requirements

## 依存関係
- 外部依存を持たず、Go標準ライブラリのみを使用して実装する
  - ファイル監視: `fsnotify`パッケージまたはOSのsystem callを使用
  - ファイル操作: `os`、`io/fs`パッケージ
  - HTTP/HTTPS: `net/http`パッケージ
  - 設定ファイル解析: 
    - 最終目標: `encoding/toml`パッケージ（Go 2.0で導入予定）
    - 暫定対応: `encoding/json`パッケージを使用し、後にTOMLに移行
  - コマンドライン引数: `flag`パッケージ

## アーキテクチャ設計
- 責務ごとに明確に分離されたモジュール構造
  - ファイル監視モジュール
  - ファイル読み込みモジュール (ローカル/リモート)
  - ファイル結合モジュール
  - 設定管理モジュール
- 参照透過性を持った純粋関数として実装
  - 副作用は明示的な境界で管理
  - テスト可能なインターフェース設計

## テスト戦略

### テストの分類と実行
テストは以下のビルドタグで分類し、適切な環境で実行する：

1. テストファイル先頭のビルドタグ指定
   ```go
   //go:build small
   ```

2. テスト実行方法
   ```bash
   # Small Tests
   go test -tags=small ./...

   # Acceptance Tests
   # Feature filesに基づいて実行
   ```

### 単体テスト (Small Tests)
- 各関数の入出力をテストするテーブル駆動テスト
- 参照透過性を持つ純粋関数の振る舞いを検証
- モックを最小限にとどめたテスト設計
- `testing`パッケージを使用
- テストにはビルドタグ `small` を付与
- Small Testの制約（[Google Testing Blog](https://testing.googleblog.com/2010/12/test-sizes.html)参照）：
  - 単一のプロセスで実行される
  - ネットワークアクセスを行わない
  - ファイルシステムにアクセスしない
  - システムコールを行わない
  - 外部プロセスの起動を行わない
  - 並行処理を行わない
  - スリープやタイムアウトを使用しない
  - テストの実行時間は数ミリ秒以内

以下の機能に対して小さな単位でテストを行う：
- ファイル内容の読み込みと解析処理
- ファイル結合とフォーマット処理
- 設定ファイル読み込みロジック
- コマンドライン引数の解析
- ファイル監視イベントの処理

### 受入テスト (Acceptance Tests)
- Gherkin記法によるfeatureファイルに基づいたテスト
- ユーザーの視点からの振る舞い全体を検証
- 以下の機能ごとに定義されたシナリオのテスト
  - ローカルファイルの監視と結合
  - リモートファイルの取得と結合
  - 設定ファイルの読み込みと適用
- 別プロセスとしての実行と結果検証
- CI/CD環境での自動実行が可能な構成

## 完了の定義
- 全ての単体テスト (Small Tests) が成功すること
- 全ての受入テスト（Gherkinシナリオ）が成功すること
- テストカバレッジが80%以上であること
- lintエラーがないこと
- 標準ライブラリ以外の依存関係がないこと

## 実装ガイドライン
- ファイル監視はポーリングではなくOSのイベント通知を使用
- ファイル変更検知時は差分ではなく全体を再構築
- エラーは適切なコンテキストとともにログ出力
- 並行処理はチャネルを使用して明示的に制御
- 全ての公開関数・型にはドキュメントコメントを付与
- コードは明示的にエラー処理を行い、エラーを無視しない

## 開発プロセスとツール

### ドキュメント管理
- 要件仕様書（requirements.md）は常に最新の状態を維持する
  - 新しい要件や指示が見つかった場合は、即時に反映する
  - 変更履歴を明確にするため、更新はコミットメッセージに詳細を記載する
- AIツール（GitHub Copilot）の設定は .github/copilot-instructions.md で管理
  - プロジェクトの重要なファイルへの参照を明示
  - 開発プロセスと要件の遵守を指示

### バージョン管理
- Gitを使用してバージョン管理を行う
- コミットメッセージは英語で記述し、[gitmoji](https://gitmoji.dev/)を使用して視覚的に分類する
  - ✨ (`:sparkles:`) - New features
  - 🐛 (`:bug:`) - Bug fixes
  - ♻️ (`:recycle:`) - Code refactoring
  - 📝 (`:memo:`) - Documentation updates
  - ✅ (`:white_check_mark:`) - Adding or updating tests
- コミットメッセージは `emoji Short description in English` の形式で記述する
  例: `📝 Add technical requirements documentation`

### コードフォーマットとリンター
- フォーマッター: `gofmt` および `goimports` を使用
  - 保存時に自動フォーマットを適用
  - タブインデント、行末の改行なしを標準スタイルとする
- リンター: `golangci-lint` を使用
  - 基本設定: `golangci-lint run --enable=gofmt,goimports,gosimple,govet,staticcheck`
  - プロジェクトルート直下に `.golangci.yml` で設定を管理
  - 開発環境では以下のコマンドでインストールする:
    ```bash
    # 最新バージョンをインストール
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    # または特定バージョン（CI環境と同じ）をインストール
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2
    ```
  - 開発中は以下のいずれかの方法でlintを実行する:
    ```bash
    # 直接実行
    golangci-lint run
    
    # または、Makefileを使用して実行
    make lint
    ```
  - コミット前に必ずlintエラーがないことを確認する

### 開発用Makefileの使用
- プロジェクトルートのMakefileに共通タスクが定義されている
- 主なターゲット:
  - `make build`: プロジェクトをビルド
  - `make test`: small testsを実行
  - `make cover`: テストカバレッジを計測して表示
  - `make lint`: golangci-lintを実行
  - `make fmt`: コードをフォーマット
  - `make all`: lint、test、buildを順番に実行
- コミット前に `make all` を実行して問題がないか確認する

### プロジェクト状況の追跡
- プロジェクトルートにTODO.mdファイルを設置し、開発状況をリアルタイムに反映する
- **操作ごと**に詳細情報を記録し、AIコーディングエージェントが迅速にコンテキストを把握できるようにする
- 以下の形式で詳細な開発状況を記録する:

```markdown
# Wampa 開発状況

## 現在の作業コンテキスト
- **作業中のファイル**: `pkg/watcher/file_watcher.go`
- **実装中の機能**: ファイル変更検知イベントハンドラー
- **関連するテスト**: `pkg/watcher/file_watcher_test.go`
- **現在の課題**: fsnotifyのイベントが重複して発生する問題を解決中

## エラーとバグ追跡
- **コンパイルエラー**: `pkg/formatter/formatter.go:45` - インターフェース実装が不完全
  ```
  formatter.go:45:6: not enough methods for Formatter interface, missing Format method
  ```
- **テストエラー**: `TestFileWatcher_Watch`
  ```
  === RUN   TestFileWatcher_Watch
     watcher_test.go:32: expected event count 1, got 2
  --- FAIL: TestFileWatcher_Watch (0.15s)
  ```
- **リントエラー**: `pkg/config/config.go:23` - エラー処理が不十分
  ```
  config.go:23: error return value not checked (errcheck)
  ```

## 実装ステータス
### 完了した機能
- [x] 基本的な設計とディレクトリ構造 (2023-06-15)
- [x] 設定ファイル読み込み機能 (2023-06-16)
  - [x] TOML解析
  - [x] デフォルト値の設定

### 進行中の機能
- [ ] ファイル監視機能
  - [x] ローカルファイルの監視
  - [ ] リモートファイルの定期チェック
  - [ ] 複数ファイルの同時監視

### 保留中の機能
- [ ] ファイル結合機能
  - [ ] セクションヘッダー追加
  - [ ] 出力ファイル書き込み

## メモと参考情報
- fsnotifyのイベント重複問題: https://github.com/fsnotify/fsnotify/issues/62
- 参考実装: https://github.com/example/file-watcher
```

- 以下の操作時にTODO.mdを必ず更新する:
  1. ファイル編集開始時（対象ファイル名を記録）
  2. コンパイル/テストエラー発生時（エラーメッセージをコピー）
  3. テスト失敗時（失敗したテスト名とエラーメッセージ）
  4. 新たな問題点や課題発見時（詳細と対応策）
  5. 機能実装完了時（実装した内容のサマリー）
  
- Wampaを使用して、TODO.mdをリアルタイムに監視・結合し、AIエージェントにコンテキスト提供
- TODO.md自体がWampaのユースケース例となり、開発プロセスの改善にもフィードバック

### 作業フロー
1. 機能追加・バグ修正ごとにfeatureブランチを作成
2. TDDアプローチ: テスト → 実装 → リファクタリングの流れで開発
3. **操作ごと**にTODO.mdを更新し、現在の状況を詳細に記録
4. コミットする前に以下を実行:
   - テストの実行: `go test ./...`
   - リンターの実行: `golangci-lint run`
   - TODO.mdの更新: 完了した項目をマーク、新たな課題を追加
5. レビュー後、mainブランチにマージ

### CI/CD
- テスト、リンター、受入テストを自動実行
- `go test -race ./...` でレースコンディションの検出
- `go test -cover ./...` でカバレッジ計測
- GitHubの場合はGitHub Actionsを使用