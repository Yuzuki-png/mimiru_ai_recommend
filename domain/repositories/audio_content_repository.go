package repositories

import (
	"context"
	"mimiru-ai/domain/entities"
)

// AudioContentRepository 音声コンテンツリポジトリのインターフェース
type AudioContentRepository interface {
	GetByID(ctx context.Context, contentID int) (*entities.AudioContent, error)
	GetByIDs(ctx context.Context, contentIDs []int) ([]*entities.AudioContent, error)
	GetSimilarContent(ctx context.Context, categoryID, authorID int, excludeIDs []int, limit int) ([]*entities.AudioContent, error)
	GetNewContent(ctx context.Context, days int, limit int) ([]*entities.AudioContent, error)
	GetPopularContent(ctx context.Context, days int, limit int) ([]*entities.AudioContent, error)
	Save(ctx context.Context, content *entities.AudioContent) error
}