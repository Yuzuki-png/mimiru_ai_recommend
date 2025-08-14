package repositories

import (
	"context"
	"mimiru-ai/domain/entities"
	"mimiru-ai/domain/repositories"
	"mimiru-ai/infrastructure/database"

	"github.com/jackc/pgx/v5"
)

// UserRepositoryImpl ユーザーリポジトリの実装
type UserRepositoryImpl struct {
	db *database.Client
}

// NewUserRepositoryImpl コンストラクタ
func NewUserRepositoryImpl(db *database.Client) repositories.UserRepository {
	return &UserRepositoryImpl{
		db: db,
	}
}

// GetByID IDでユーザーを取得
func (r *UserRepositoryImpl) GetByID(ctx context.Context, userID int) (*entities.User, error) {
	query := `
		SELECT id, email, created_at 
		FROM "User" 
		WHERE id = $1
	`

	var user entities.User
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// GetSimilarUsers 類似ユーザーを取得
func (r *UserRepositoryImpl) GetSimilarUsers(ctx context.Context, userID int, limit int) ([]*entities.User, error) {
	query := `
		WITH user_categories AS (
			SELECT DISTINCT ac.category_id
			FROM "ListenHistory" lh
			JOIN "AudioContent" ac ON lh.audio_content_id = ac.id
			WHERE lh.user_id = $1
			  AND lh.created_at > NOW() - INTERVAL '30 days'
		),
		similar_users AS (
			SELECT u.id, u.email, u.created_at, COUNT(*) as common_categories
			FROM "ListenHistory" lh
			JOIN "AudioContent" ac ON lh.audio_content_id = ac.id
			JOIN user_categories uc ON ac.category_id = uc.category_id
			JOIN "User" u ON lh.user_id = u.id
			WHERE lh.user_id != $1
			  AND lh.created_at > NOW() - INTERVAL '30 days'
			GROUP BY u.id, u.email, u.created_at
			HAVING COUNT(*) >= 2
		)
		SELECT id, email, created_at
		FROM similar_users
		ORDER BY common_categories DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

// Save ユーザーを保存
func (r *UserRepositoryImpl) Save(ctx context.Context, user *entities.User) error {
	if !user.IsValid() {
		return ErrInvalidEntity
	}

	query := `
		INSERT INTO "User" (email, created_at)
		VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET
			created_at = EXCLUDED.created_at
		RETURNING id
	`

	err := r.db.Pool.QueryRow(ctx, query, user.Email, user.CreatedAt).Scan(&user.ID)
	return err
}