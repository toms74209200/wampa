Feature: リモートファイルの取得と結合
  Wampaはリモート(HTTP/HTTPS)上のファイルを取得し、出力ファイルに結合できる

  @large
  Scenario: リモートファイルの取得と出力
    When wampaを以下のコマンドで実行:
      """
      wampa -i https://raw.githubusercontent.com/toms74209200/wampa/028171afb7eefed15d055b4d82618280c9782f74/TODO.md -o output.md
      """
    Then output.mdは以下の内容を含む:
      """
      [//]: # "filepath: https://raw.githubusercontent.com/toms74209200/wampa/028171afb7eefed15d055b4d82618280c9782f74/TODO.md"
      # Wampa 開発状況
      """

  @large
  Scenario: ローカルファイルとリモートファイルの組み合わせ
    Given 以下の内容のlocal.mdが存在する:
      """
      # ローカル仕様
      - 機能L: ローカル処理を行う
      """
    When wampaを以下のコマンドで実行:
      """
      wampa -i local.md https://raw.githubusercontent.com/toms74209200/wampa/028171afb7eefed15d055b4d82618280c9782f74/TODO.md -o output.md
      """
    Then output.mdは以下の内容を含む:
      """
      [//]: # "filepath: local.md"
      # ローカル仕様
      - 機能L: ローカル処理を行う
      """
    And output.mdは以下の内容を含む:
      """
      [//]: # "filepath: https://raw.githubusercontent.com/toms74209200/wampa/028171afb7eefed15d055b4d82618280c9782f74/TODO.md"
      # Wampa 開発状況
      """