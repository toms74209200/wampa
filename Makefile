.PHONY: all build test lint fmt clean

# デフォルトターゲット
all: lint test build

# ビルド
build:
	go build -v ./...

# テスト実行
test:
	go test -tags=small -v ./...

# テストカバレッジ計測
cover:
	go test -tags=small -race -coverprofile=coverage.txt -covermode=atomic `cat scripts/coverage_pkgs.txt`
	go tool cover -func=coverage.txt

# lintの実行
lint:
	golangci-lint run

# フォーマット
fmt:
	go fmt ./...
	goimports -w .

# クリーン
clean:
	rm -f coverage.txt
	rm -f wampa