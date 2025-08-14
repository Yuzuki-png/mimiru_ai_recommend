# CLAUDE.md

このファイルは、Claude Code (claude.ai/code) がこのリポジトリのコードを扱う際のガイダンスを提供します。
必ず日本語で回答してください。

## 概要

Mimiru AIレコメンドサービスは、NestJS（mimiru_api）と統合されるGo言語ベースのマイクロサービスです。PostgreSQLに直接接続し、高負荷な推薦アルゴリズム（協調フィルタリング、コンテンツベース、人気度ベース、新着コンテンツ）を実行します。リアルタイムのデータベース監視機能とRedisキャッシュを組み合わせて最適なパフォーマンスを実現します。

## 開発コマンド

### セットアップ
```bash
# 依存関係のインストール
go mod tidy

# 環境変数設定
cp .env.example .env
# .envファイルの設定例:
# DATABASE_URL=postgresql://postgres:postgres@localhost:5432/radio_site?schema=public&sslmode=disable
# REDIS_URL=localhost:6379
# PORT=8080
```

### サービス実行
```bash
# 開発サーバー起動（クリーンアーキテクチャ版）
go run cmd/main.go

# バイナリビルド
go build -o recommendation ./cmd/main.go

# Dockerビルドと実行
docker build -t mimiru-recommendation .
docker run -p 8080:8080 mimiru-recommendation
```

### テスト
```bash
# ユニットテストのみ実行
make test-unit

# 統合テストのみ実行（要DB・Redisセットアップ）
make test-integration

# 全テスト実行
make test

# テストカバレッジ取得
make test-coverage

# ベンチマークテスト実行
make test-benchmark
```

テスト構成：
- **domain/entities/**: ドメインエンティティのユニットテスト
- **usecases/**: ユースケースのテスト（モック使用）
- **controllers/**: HTTPコントローラーのテスト（必要に応じて追加）
- **integration_test.go**: エンドツーエンドの統合テスト

## アーキテクチャ

このサービスはクリーンアーキテクチャパターンに基づき、以下の4層構造で設計されています：

### 1. ドメイン層（最内層）
- **domain/entities/**: 純粋なビジネスエンティティ（User, AudioContent, Recommendation等）
- **domain/repositories/**: リポジトリインターフェース定義
- **domain/services/**: ドメインサービス（RecommendationAlgorithmService）

### 2. ユースケース層
- **usecases/**: アプリケーションのビジネスルール（GetRecommendationsUsecase, TrackEventUsecase）
- 各ユースケースは1つのパブリックメソッド（Execute）のみを持つ

### 3. インターフェース層
- **controllers/**: HTTPコントローラー（Ginフレームワーク使用）
- **cmd/**: エントリーポイントと依存性注入コンテナ

### 4. インフラストラクチャ層（最外層）
- **infrastructure/database/**: PostgreSQL接続クライアント
- **infrastructure/cache/**: Redis接続クライアント
- **infrastructure/repositories/**: リポジトリの具体実装

### 依存性の方向
外側の層は内側の層に依存できるが、内側の層は外側の層に依存しない。
インフラ層はインターフェースを通してドメイン層と疎結合。

### レコメンドエンジン
コアレコメンドシステムは重み付きスコアリングアプローチを実装：
- **協調フィルタリング（40%）**: 類似ユーザーの好み
- **コンテンツベース（30%）**: カテゴリと作者の類似性
- **人気度ベース（20%）**: トレンディングコンテンツ（7日間）
- **新着コンテンツ（10%）**: 最近追加されたコンテンツ（3日間）

すべての推薦は重複排除され、合計スコア順にソートされ、Redis で1時間TTLでキャッシュされます。

### データフロー
1. HTTPリクエスト → ハンドラー層
2. ハンドラー → ビジネスロジック用サービス層
3. サービス → データベースクエリ用リポジトリ層
4. 結果をRedisにキャッシュしてJSONで返却

### 依存関係
- **Webフレームワーク**: Gin (github.com/gin-gonic/gin)
- **データベース**: PostgreSQL with pgx driver (github.com/jackc/pgx/v5)
- **キャッシュ**: Redis with go-redis client (github.com/go-redis/redis/v8)
- **設定**: 環境変数用 godotenv

## 環境設定

必要な環境変数：
- `DATABASE_URL`: PostgreSQL接続文字列
- `REDIS_URL`: Redisサーバーアドレス（デフォルト: localhost:6379）
- `PORT`: サーバーポート（デフォルト: 8080）

## APIエンドポイント

### GET /health
ヘルスチェックエンドポイント（DB接続確認含む）
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "database": "connected",
    "version": "2.0.0-clean-arch"
  },
  "timestamp": 1640995200,
  "service": "mimiru-recommendation"
}
```

### GET /recommendations
ユーザー推薦取得（NestJS統合用）
- **Query Parameters**: `user_id` (required), `limit` (optional, default: 20)
- **Example**: `/recommendations?user_id=123&limit=10`
```json
{
  "success": true,
  "data": {
    "userId": 123,
    "recommendations": [...],
    "timestamp": 1640995200
  },
  "timestamp": 1640995200,
  "service": "mimiru-recommendation"
}
```

### POST /events  
ユーザーインタラクションイベント追跡
```json
{
  "userId": 1,
  "audioContentId": 123,
  "eventType": "play",
  "duration": 180
}
```

## パフォーマンス目標

- レスポンス時間: <100ms（キャッシュヒット時）
- スループット: 1000 req/s
- メモリ使用量: <256MB
- キャッシュ戦略: 推薦結果1時間TTL

## NestJS統合

### 接続設定
```typescript
// NestJS側の環境変数
GO_RECOMMENDATION_SERVICE_URL="http://localhost:8080"

// サービス呼び出し例
const response = await this.httpClientService.get<RecommendationResponse>(
  `${this.goServiceBaseUrl}/recommendations`,
  {
    params: { user_id: userId, limit },
    timeout: 5000,
  }
);
```

### エラーハンドリング
統一されたエラーレスポンス形式:
```json
{
  "error": "Bad Request",
  "message": "user_id parameter is required",
  "details": "additional error details",
  "timestamp": 1640995200,
  "service": "mimiru-recommendation"
}
```

## バックグラウンド監視

### データベース監視機能
- **playback_sessions** テーブルの変更を自動監視
- **user_ratings** テーブルの変更を自動監視  
- ユーザーのレコメンドキャッシュを自動無効化
- ポーリング監視（30秒間隔）またはLISTEN/NOTIFY使用

### キャッシュ戦略
- レコメンド結果: 1時間TTL
- データ変更時の自動キャッシュ無効化
- 関連ユーザーのキャッシュも部分的に無効化

## Dockerデプロイ

このサービスは、ECSでNestJS APIと並んでサイドカーコンテナとして実行されるよう設計されており、タスクリソースの25%（0.125 vCPU、256MBメモリ）を消費します。バックグラウンド監視サービスも自動起動されます。