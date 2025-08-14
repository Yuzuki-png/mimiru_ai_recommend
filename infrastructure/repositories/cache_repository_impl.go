package repositories

import (
	"context"
	"mimiru-ai/domain/repositories"
	"mimiru-ai/infrastructure/cache"
	"time"
)

// CacheRepositoryImpl キャッシュリポジトリの実装
type CacheRepositoryImpl struct {
	cache *cache.Client
}

// NewCacheRepositoryImpl コンストラクタ
func NewCacheRepositoryImpl(cache *cache.Client) repositories.CacheRepository {
	return &CacheRepositoryImpl{
		cache: cache,
	}
}

// Get キャッシュから値を取得
func (r *CacheRepositoryImpl) Get(ctx context.Context, key string, dest interface{}) error {
	return r.cache.Get(key, dest)
}

// Set キャッシュに値を設定
func (r *CacheRepositoryImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.cache.Set(key, value, expiration)
}

// Delete キャッシュから値を削除
func (r *CacheRepositoryImpl) Delete(ctx context.Context, key string) error {
	return r.cache.Delete(key)
}

// Exists キーの存在確認
func (r *CacheRepositoryImpl) Exists(ctx context.Context, key string) (bool, error) {
	return r.cache.Exists(key)
}