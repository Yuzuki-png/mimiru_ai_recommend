package repositories

import (
	"context"
	"mimiru-ai/domain/entities"
)

// UserRepository ユーザーリポジトリのインターフェース
type UserRepository interface {
	GetByID(ctx context.Context, userID int) (*entities.User, error)
	GetSimilarUsers(ctx context.Context, userID int, limit int) ([]*entities.User, error)
	Save(ctx context.Context, user *entities.User) error
}

// UserPreferenceRepository ユーザー好みリポジトリのインターフェース
type UserPreferenceRepository interface {
	GetUserPreferences(ctx context.Context, userID int) ([]*entities.UserPreference, error)
	SaveUserPreference(ctx context.Context, preference *entities.UserPreference) error
	UpdateUserPreferences(ctx context.Context, userID int, preferences []*entities.UserPreference) error
}