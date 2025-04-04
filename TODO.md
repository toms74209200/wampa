# Wampa 開発状況

## 現在の作業コンテキスト
- **作業中のファイル**: なし（実装完了）
- **実装中の機能**: 標準入力処理
- **関連するテスト**: features/standard_input_handling.feature
- **現在の課題**: 標準入力機能の実装

## エラーとバグ追跡
- **コンパイルエラー**: なし
- **テストエラー**: なし
- **リントエラー**: なし

## 実装ステータス
### 完了した機能
- [x] プロジェクト仕様書の作成 (spec.md)
- [x] 技術要件の定義 (requirements.md)
  - [x] TOML対応の課題を明確化し、暫定対応としてJSONを使用
- [x] 受入テスト定義 (features/*.feature)
  - [x] config_file_handling.feature
  - [x] local_file_monitoring.feature
  - [x] remote_file_handling.feature
  - [x] standard_input_handling.feature
- [x] プロジェクト構造の設定
  - [x] ディレクトリ構造の作成
  - [x] 主要インターフェースの定義
  - [x] 依存関係の確認
- [x] 設定ファイル処理
  - [x] 設定ファイル読み込み（JSONベース）
  - [x] テーブル駆動テストの実装
  - [x] コマンドラインオプションと設定ファイルの責務分離
  - [ ] TOMLサポートへの移行（Go 2.0で導入予定のencoding/tomlパッケージ待ち）
- [x] ファイル監視モジュール
  - [x] ローカルファイル監視インターフェース定義
  - [x] fsnotifyを使用したローカルファイル監視実装
  - [x] ローカルファイル監視のテスト実装
  - [x] リモートファイル処理の基本実装
    - [x] HTTPリクエスト生成（純粋関数）
    - [x] レスポンス処理（純粋関数）
    - [x] ストリーミング処理（純粋関数）
    - [x] テストの実装（small tests）
  - [x] リモートファイル処理のLarge Testの実装
    - [x] remote_file_handling.featureの実装
    - [x] GitHub Actionsワークフローの追加 (test-large.yml)
  - [ ] リモートファイル監視（スコープ外として保留）
- [x] ファイル結合機能
  - [x] フォーマッターインターフェース設計
  - [x] Markdown対応フォーマット処理の実装
  - [x] フォーマッターのテスト実装
- [x] local_file_monitoring.featureの受け入れテスト実装
- [x] コマンドライン引数のパース処理
  - [x] `flag`パッケージを使用した引数パース処理の実装
  - [x] テスト駆動開発でユニットテストを実装
  - [x] 純粋関数としてテスト可能な設計を実現
- [x] ヘルプ機能
  - [x] ヘルプフラグ（-h, --help）の実装
  - [x] ヘルプメッセージの定義（定数として実装）
  - [x] wampa.Run関数でのヘルプフラグチェックと表示の実装
  - [x] ヘルプメッセージ表示のアクセプタンステストの実装
- [x] メインパッケージの実装
  - [x] シグナル処理によるグレースフルシャットダウン
  - [x] コンテキストを使ったリソース管理
  - [x] エラー処理の実装
- [x] エラー処理の改善
  - [x] 無効なパラメータ入力時のエラーメッセージとヘルプ表示
  - [x] 設定ファイル不在時のエラー処理
- [x] config_file_handling.featureの実装
  - [x] ヘルプメッセージ表示のアクセプタンステスト
  - [x] エラーメッセージ表示のアクセプタンステスト

### 保留中の機能
- [ ] 標準入力サポート
  - [ ] コマンドライン引数に標準入力フラグ(-s, --stdin)を追加
  - [ ] 標準入力の読み取り処理の実装
  - [ ] 標準入力コンテンツと他の入力ファイルの結合処理
  - [ ] standard_input_handling.featureの受け入れテスト実装
- [ ] TOMLサポートへの移行（標準ライブラリのencoding/tomlパッケージ対応待ち）
- [ ] リモートファイル監視（スコープ外）

## メモと参考情報
- TOMLサポートについて：
  - Go 2.0でencoding/tomlパッケージが導入予定
  - 暫定対応としてJSONを使用し、後でTOMLに移行
  - 移行時の影響を最小限にするため、設定ファイル処理を独立したパッケージとして実装
- テストカバレッジ
  - config パッケージ: テーブル駆動テストで主要パスをカバー
    - 設定ファイル処理とコマンドラインオプションの責務を分離
    - LoadFromFileはmedium testで検証予定
    - コマンドライン引数のパース処理は純粋関数としてテスト可能
  - watcher パッケージ: モックを使用したテストで複数のイベントケースをカバー
  - formatter パッケージ: テーブル駆動テストで実装完了
  - acceptance テスト: 
    - local_file_monitoring.featureとconfig_file_handling.featureはmedium testで実装
    - remote_file_handling.featureはlarge testで実装（GitHub上のリモートファイル取得）

## 直近の作業履歴
- 標準入力機能のサポートに関する計画を追加
  - standard_input_handling.featureファイルの作成
  - TODO.mdの更新
- リモートファイル処理のLarge Test実装が完了
  - GitHub上のリモートファイル取得テストの実装
  - Large Test用のGitHub Actionsワークフロー作成
  - テストがすべて通過していることを確認
- 次のステップ: 標準入力機能の実装