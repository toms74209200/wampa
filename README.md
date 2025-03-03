## Motivation

- AIコーディングエージェントが、プログラムのコードを生成する際に必要となるコンテキスト情報やルールなどの情報を統合して一つのファイルとして提供する
- AIコーディングエージェントには.clinerulesや.cursorrulesなどのファイルによって、コード生成のルールを提供できるが、使っているエージェントによって異なることや個々人が設定したいルールを別に持ちたいなど、ルールファイルが散逸してしまう問題がある
- 人間のエンジニアがコードベースを参照する場合、プロジェクトの仕様やコーディングルールはコンテキストごとに分かれているほうがわかりやすい

## Usage

wampa を実行すると<input_files> を監視して <input_files> の変更のたびに <output_file> を更新する。 <input_files> は複数のファイルをスペース区切りで指定できる。

```bash
wampa -i <input_files> -o <output_file>
```
wampa を実行するカレントディレクトリに設定ファイルwampa.tomlが存在する場合、設定ファイルに記述されたパラメータを使用して処理を行う。この場合引数なしで実行できる。

```bash
wampa
```

wampa.tomlの例
```toml
input_files = ["input1.md", "input2.txt"]
output_file = "output.txt"
```

wampa.tomlが存在し、なおかつコマンドライン引数が指定されている場合、コマンドライン引数が優先される。

ネットワークを通じてオンライン上にあるファイルを指定することもできる

```bash
wampa -i https://example.com/input1.md -o output.txt
```

## Technical requirements

- インストールが容易であること
- 実行が容易であること
- 依存関係が限りなく0に近いこと
- バージョン更新がほとんど不要なこと
- メンテナンスが容易であること
- テストが容易であること

## Developer Information

### Setup Development Environment

```bash
# Go環境のセットアップ（要Go 1.21以上）
go mod download

# golangci-lintのインストール
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.54.2

# goimportsのインストール
go install golang.org/x/tools/cmd/goimports@latest
```

### Development Commands

プロジェクトルートのMakefileで以下のコマンドが利用可能です：

```bash
# すべての検証とビルドを実行
make all

# コードのビルドのみ
make build

# テストの実行（small tests）
make test

# カバレッジレポートの生成
make cover

# linterの実行
make lint

# コードフォーマット
make fmt
```

### Development Workflow

1. 機能追加・バグ修正を始める前に最新のmainブランチを取得
2. 新しいfeatureブランチを作成
3. TDDアプローチで開発：テスト → 実装 → リファクタリング
4. コードの変更に合わせてTODO.mdを更新
5. コミット前に `make all` を実行して問題がないか確認
6. コミットメッセージは[gitmoji](https://gitmoji.dev/)を使用（例: `✨ Add config file parsing`）
7. Pull Requestを作成