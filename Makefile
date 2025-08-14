.PHONY: test test-unit test-integration test-coverage build run clean

# ビルド
build:
	go build -o bin/recommendation ./cmd/main.go

# 開発サーバー起動
run:
	go run cmd/main.go

# 全テスト実行
test: test-unit test-integration

# ユニットテスト実行
test-unit:
	go test -v -short ./...

# 統合テスト実行
test-integration:
	INTEGRATION_TEST=1 go test -v -tags=integration .

# テストカバレッジ取得
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "カバレッジレポートが coverage.html に生成されました"

# ベンチマークテスト実行
test-benchmark:
	go test -v -bench=. ./...

# 統合ベンチマークテスト実行
test-benchmark-integration:
	INTEGRATION_TEST=1 go test -v -bench=. -tags=integration .

# Dockerビルド
docker-build:
	docker build -t mimiru-recommendation .

# Dockerテスト実行
docker-test:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down

# 依存関係更新
deps:
	go mod tidy
	go mod download

# 静的解析
lint:
	golangci-lint run

# フォーマット
fmt:
	go fmt ./...

# クリーンアップ
clean:
	rm -f bin/recommendation
	rm -f coverage.out coverage.html
	go clean -testcache

# ヘルプ
help:
	@echo "利用可能なコマンド:"
	@echo "  build                   - バイナリをビルド"
	@echo "  run                     - 開発サーバー起動"
	@echo "  test                    - 全テスト実行"
	@echo "  test-unit               - ユニットテストのみ実行"
	@echo "  test-integration        - 統合テストのみ実行"
	@echo "  test-coverage           - テストカバレッジ取得"
	@echo "  test-benchmark          - ベンチマークテスト実行"
	@echo "  docker-build            - Dockerイメージビルド"
	@echo "  docker-test             - Docker環境でテスト実行"
	@echo "  deps                    - 依存関係更新"
	@echo "  lint                    - 静的解析実行"
	@echo "  fmt                     - コードフォーマット"
	@echo "  clean                   - ビルド成果物削除"