Feature: 設定ファイルの読み込みと適用
  Wampaは設定ファイル(wampa.json)から入出力ファイルの設定を読み込む

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

  @medium
  Scenario: 設定ファイルからの入出力ファイル読み込み
    Given 以下の内容のwampa.jsonが存在する:
      """
      {
        "input_files": ["spec.md", "rules.md"],
        "output_file": "combined.md"
      }
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

  @medium
  Scenario: 設定ファイルとコマンドラインパラメータの優先順位
    Given 以下の内容のwampa.jsonが存在する:
      """
      {
        "input_files": ["spec.md", "rules.md"],
        "output_file": "combined.md"
      }
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

  @medium
  Scenario: 設定ファイル不在時のエラー処理
    When カレントディレクトリにwampa.jsonが存在しない状態でwampaをパラメータなしで実行:
      """
      wampa
      """
    Then 以下のエラーメッセージが表示される:
      """
      Configuration file wampa.json not found. Please specify -i and -o options or create a configuration file.
      """
    Then 以下のヘルプメッセージが表示される:
      """
      Usage: wampa [options]

      Options:
        -i, --input   Specify input file(s) (can be specified multiple times)
        -o, --output  Specify output file
        -h, --help    Display this help message
      """
    And プロセスは非ゼロの終了コードで終了する

  @medium
  Scenario Outline: ヘルプフラグによるヘルプ表示
    When wampaを以下のコマンドで実行:
      """
      wampa <flag>
      """
    Then 以下のヘルプメッセージが表示される:
      """
      Usage: wampa [options]

      Options:
        -i, --input   Specify input file(s) (can be specified multiple times)
        -o, --output  Specify output file
        -h, --help    Display this help message
      """
    And プロセスはゼロの終了コードで終了する

    Examples:
      | flag   |
      | -h     |
      | --help |

  @medium
  Scenario: 不正な引数指定時のヘルプ表示
    When wampaを以下のコマンドで実行:
      """
      wampa -x
      """
    Then 以下のエラーメッセージが表示される:
      """
      Unknown option: -x
      """
    Then 以下のヘルプメッセージが表示される:
      """
      Usage: wampa [options]

      Options:
        -i, --input   Specify input file(s) (can be specified multiple times)
        -o, --output  Specify output file
        -h, --help    Display this help message
      """
    And プロセスは非ゼロの終了コードで終了する

  @medium
  Scenario: 出力ファイル未指定時のエラー処理
    When カレントディレクトリにwampa.jsonが存在しない状態でwampaを以下のコマンドで実行:
      """
      wampa -i spec.md
      """
    Then 以下のエラーメッセージが表示される:
      """
      Output file not specified. Please specify -o option or create a configuration file.
      """
    Then 以下のヘルプメッセージが表示される:
      """
      Usage: wampa [options]

      Options:
        -i, --input   Specify input file(s) (can be specified multiple times)
        -o, --output  Specify output file
        -h, --help    Display this help message
      """
    And プロセスは非ゼロの終了コードで終了する