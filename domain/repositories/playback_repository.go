package repositories

import (
	"context"
	"mimiru-ai/domain/entities"
)

// PlaybackRepository 再生履歴リポジトリのインターフェース
type PlaybackRepository interface {
	GetUserHistory(ctx context.Context, userID int, limit int) ([]*entities.PlaybackHistory, error)
	SavePlayback(ctx context.Context, history *entities.PlaybackHistory) error
	GetRecentPlaybacks(ctx context.Context, userID int, days int) ([]*entities.PlaybackHistory, error)
}

// RecommendationRepository レコメンドリポジトリのインターフェース
type RecommendationRepository interface {
	SaveRecommendationSet(ctx context.Context, recSet *entities.RecommendationSet) error
	GetRecommendationSet(ctx context.Context, userID int) (*entities.RecommendationSet, error)
}