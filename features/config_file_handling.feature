Feature: 設定ファイルの読み込みと適用
  Wampaは設定ファイル(wampa.toml)から入出力ファイルの設定を読み込む

  Background:
    Given 以下の内容のspec.mdが存在する:
      """
      # 製品仕様書
      これは製品仕様書です
      """
    And 以下の内容のrules.mdが存在する:
      """
      # コーディングルール
      これはコーディングルールです
      """

  Scenario: 設定ファイルからの入出力ファイル読み込み
    Given 以下の内容のwampa.tomlが存在する:
      """
      input_files = ["spec.md", "rules.md"]
      output_file = "combined.md"
      """
    When wampaをパラメータなしで実行:
      """
      wampa
      """
    Then combined.mdは以下の内容を含む:
      """
      [//]: # "filepath: spec.md"
      # 製品仕様書
      これは製品仕様書です

      [//]: # "filepath: rules.md"
      # コーディングルール
      これはコーディングルールです
      """

  Scenario: 設定ファイルとコマンドラインパラメータの優先順位
    Given 以下の内容のwampa.tomlが存在する:
      """
      input_files = ["spec.md", "rules.md"]
      output_file = "combined.md"
      """
    When wampaを以下のコマンドで実行:
      """
      wampa -i spec.md -o override.md
      """
    Then combined.mdは作成されない
    And override.mdは以下の内容を含む:
      """
      [//]: # "filepath: spec.md"
      # 製品仕様書
      これは製品仕様書です
      """

  Scenario: 設定ファイル不在時のエラー処理
    When カレントディレクトリにwampa.tomlが存在しない状態でwampaをパラメータなしで実行:
      """
      wampa
      """
    Then 以下のエラーメッセージが表示される:
      """
      設定ファイル wampa.toml が見つかりません。-i および -o オプションを指定するか、設定ファイルを作成してください。
      """
    And プロセスは非ゼロの終了コードで終了する