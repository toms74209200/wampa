Feature: リモートファイルの取得と結合
  Wampaはリモート(HTTP/HTTPS)上のファイルを取得し、出力ファイルに結合できる

  Scenario: リモートファイルの取得と出力
    Given HTTPモックサーバーが以下の応答を返すよう設定:
      | URL                      | Content                                  |
      | /remote/spec.md          | # リモート仕様\n- 機能D: リモート処理を行う |
    When wampaを以下のコマンドで実行:
      """
      wampa -i http://localhost:8080/remote/spec.md -o output.md
      """
    Then output.mdは以下の内容を含む:
      """
      [//]: # "filepath: http://localhost:8080/remote/spec.md"
      # リモート仕様
      - 機能D: リモート処理を行う
      """

  Scenario: ローカルファイルとリモートファイルの組み合わせ
    Given 以下の内容のlocal.mdが存在する:
      """
      # ローカル仕様
      - 機能L: ローカル処理を行う
      """
    And HTTPモックサーバーが以下の応答を返すよう設定:
      | URL                      | Content                                  |
      | /remote/spec.md          | # リモート仕様\n- 機能R: リモート処理を行う |
    When wampaを以下のコマンドで実行:
      """
      wampa -i local.md http://localhost:8080/remote/spec.md -o output.md
      """
    Then output.mdは以下の内容を含む:
      """
      [//]: # "filepath: local.md"
      # ローカル仕様
      - 機能L: ローカル処理を行う

      [//]: # "filepath: http://localhost:8080/remote/spec.md"
      # リモート仕様
      - 機能R: リモート処理を行う
      """

  Scenario: リモートファイルの定期的な更新確認
    Given HTTPモックサーバーが以下の応答を返すよう設定:
      | URL                      | Content                                  |
      | /remote/spec.md          | # リモート仕様\n- バージョン1             |
    When wampaを以下のコマンドで実行:
      """
      wampa -i http://localhost:8080/remote/spec.md -o output.md
      """
    Then output.mdは以下の内容を含む:
      """
      [//]: # "filepath: http://localhost:8080/remote/spec.md"
      # リモート仕様
      - バージョン1
      """
    When HTTPモックサーバーが以下の応答を返すよう設定を更新:
      | URL                      | Content                                  |
      | /remote/spec.md          | # リモート仕様\n- バージョン2             |
    Then 60秒以内にoutput.mdは以下の内容に更新される:
      """
      [//]: # "filepath: http://localhost:8080/remote/spec.md"
      # リモート仕様
      - バージョン2
      """