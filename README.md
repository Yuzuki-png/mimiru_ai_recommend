# Mimiru AI Recommendation Service

MimiruのレコメンドAI機能を提供するGoサービスです。

## 🚀 特徴

- **高性能**: Goによる高速レコメンド計算
- **多角的アルゴリズム**: 協調フィルタリング、コンテンツベース、人気度、新着を組み合わせ
- **キャッシュ**: Redisによる高速レスポンス
- **軽量**: Dockerコンテナで軽量実行

## 📋 API

### GET /recommendations/:userId
指定ユーザーのレコメンドを取得

**Response:**
```json
{
  "recommendations": [
    {
      "audioContentId": 123,
      "score": 4.5,
      "reason": "similar_users"
    }
  ],
  "userId": 1,
  "timestamp": 1640995200
}
```

### POST /events
ユーザーイベントを追跡

**Request:**
```json
{
  "userId": 1,
  "audioContentId": 123,
  "eventType": "play",
  "duration": 180
}
```

### GET /health
ヘルスチェック

## 🛠️ 開発

### 前提条件
- Go 1.21+
- PostgreSQL 15+
- Redis 7+

### セットアップ
```bash
# 依存関係インストール
go mod tidy

# 環境変数設定
cp .env.example .env

# 開発サーバー起動
go run cmd/main.go
```

### Docker実行
```bash
docker build -t mimiru-recommendation .
docker run -p 8080:8080 mimiru-recommendation
```

## 🏗️ アーキテクチャ

### ディレクトリ構造
```
cmd/           # エントリーポイント
internal/      # 内部パッケージ
├── handler/   # HTTPハンドラー
├── service/   # ビジネスロジック
├── repository/ # データアクセス
└── model/     # データモデル
pkg/           # 公開パッケージ
├── cache/     # Redisクライアント
└── database/  # PostgreSQLクライアント
```

### レコメンドアルゴリズム
1. **協調フィルタリング (40%)**: 類似ユーザーベース
2. **コンテンツベース (30%)**: カテゴリ・作者類似
3. **人気度ベース (20%)**: トレンディングコンテンツ
4. **新着コンテンツ (10%)**: 新規コンテンツ

## 🔧 設定

### 環境変数
- `DATABASE_URL`: PostgreSQL接続文字列
- `REDIS_URL`: Redis接続文字列
- `PORT`: サーバーポート（デフォルト: 8080）

### キャッシュ戦略
- レコメンド結果: 1時間キャッシュ
- ユーザー履歴: 30分キャッシュ
- 人気コンテンツ: 1時間キャッシュ

## 📊 パフォーマンス

### 目標
- レスポンス時間: < 100ms（キャッシュヒット時）
- スループット: 1000 req/s
- メモリ使用量: < 256MB

### 最適化
- 並行処理でDBクエリ高速化
- インメモリキャッシュでレスポンス向上
- 接続プールでDB効率化

## 🚀 デプロイ

### ECS統合
NestJS APIと同一タスクでサイドカーパターン実行

### リソース配分
- CPU: 0.125 vCPU (25%)
- メモリ: 256MB (25%)
- ポート: 8080

### 監視
- ヘルスチェック: `/health`
- ログ: CloudWatch Logs
- メトリクス: ECS標準メトリクス# mimiru_ai_recommend
