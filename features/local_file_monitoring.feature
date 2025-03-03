Feature: ローカルファイルの監視と結合
  Wampaはローカルのファイルを監視し、変更があった場合に出力ファイルを更新する

  Background:
    Given 以下の内容のspec.mdが存在する:
      """
      # 製品仕様
      - 機能A: Xを行う
      - 機能B: Yを行う
      """
    And 以下の内容のrules.mdが存在する:
      """
      # コーディング規則
      1. 変数名はcamelCaseを使用
      2. 公開関数にはコメントを追加
      """

  Scenario: 単一ファイルの監視と出力
    When wampaを以下のコマンドで実行:
      """
      wampa -i spec.md -o output.md
      """
    Then output.mdは以下の内容を含む:
      """
      [//]: # "filepath: spec.md"
      # 製品仕様
      - 機能A: Xを行う
      - 機能B: Yを行う
      """
    When spec.mdを以下の内容に変更:
      """
      # 製品仕様
      - 機能A: Xを行う
      - 機能B: Yを行う
      - 機能C: Zを行う
      """
    Then 5秒以内にoutput.mdは以下の内容に更新される:
      """
      [//]: # "filepath: spec.md"
      # 製品仕様
      - 機能A: Xを行う
      - 機能B: Yを行う
      - 機能C: Zを行う
      """

  Scenario: 複数ファイルの監視と出力
    When wampaを以下のコマンドで実行:
      """
      wampa -i spec.md rules.md -o output.md
      """
    Then output.mdは以下の内容を含む:
      """
      [//]: # "filepath: spec.md"
      # 製品仕様
      - 機能A: Xを行う
      - 機能B: Yを行う

      [//]: # "filepath: rules.md"
      # コーディング規則
      1. 変数名はcamelCaseを使用
      2. 公開関数にはコメントを追加
      """
    When rules.mdを以下の内容に変更:
      """
      # コーディング規則
      1. 変数名はcamelCaseを使用
      2. 公開関数にはコメントを追加
      3. テストカバレッジは80%以上
      """
    Then 5秒以内にoutput.mdは以下の内容に更新される:
      """
      [//]: # "filepath: spec.md"
      # 製品仕様
      - 機能A: Xを行う
      - 機能B: Yを行う

      [//]: # "filepath: rules.md"
      # コーディング規則
      1. 変数名はcamelCaseを使用
      2. 公開関数にはコメントを追加
      3. テストカバレッジは80%以上
      """