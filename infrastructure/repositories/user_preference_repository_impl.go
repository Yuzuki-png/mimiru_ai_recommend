package repositories

import (
	"context"
	"mimiru-ai/domain/entities"
	"mimiru-ai/domain/repositories"
	"mimiru-ai/infrastructure/database"
)

// UserPreferenceRepositoryImpl ユーザー好みリポジトリの実装
type UserPreferenceRepositoryImpl struct {
	db *database.Client
}

// NewUserPreferenceRepositoryImpl コンストラクタ
func NewUserPreferenceRepositoryImpl(db *database.Client) repositories.UserPreferenceRepository {
	return &UserPreferenceRepositoryImpl{
		db: db,
	}
}

// GetUserPreferences ユーザーの好みを取得
func (r *UserPreferenceRepositoryImpl) GetUserPreferences(ctx context.Context, userID int) ([]*entities.UserPreference, error) {
	query := `
		SELECT ac.category_id, 
			   COUNT(*) as play_count,
			   AVG(CASE WHEN lh.completed THEN 1.0 ELSE 0.5 END) as completion_rate
		FROM "ListenHistory" lh
		JOIN "AudioContent" ac ON lh.audio_content_id = ac.id
		WHERE lh.user_id = $1
		  AND lh.created_at > NOW() - INTERVAL '90 days'
		GROUP BY ac.category_id
		HAVING COUNT(*) >= 3
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []*entities.UserPreference
	for rows.Next() {
		var categoryID int
		var playCount int
		var completionRate float64

		if err := rows.Scan(&categoryID, &playCount, &completionRate); err != nil {
			return nil, err
		}

		// スコア計算: 再生回数 × 完了率を正規化
		score := (float64(playCount) * completionRate) / 100.0
		if score > 1.0 {
			score = 1.0
		}

		preference := &entities.UserPreference{
			UserID:     userID,
			CategoryID: categoryID,
			Score:      score,
		}
		preferences = append(preferences, preference)
	}

	return preferences, rows.Err()
}

// SaveUserPreference ユーザーの好みを保存
func (r *UserPreferenceRepositoryImpl) SaveUserPreference(ctx context.Context, preference *entities.UserPreference) error {
	query := `
		INSERT INTO "UserPreference" (user_id, category_id, score, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, category_id) DO UPDATE SET
			score = EXCLUDED.score,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Pool.Exec(ctx, query, preference.UserID, preference.CategoryID, preference.Score)
	return err
}

// UpdateUserPreferences ユーザーの好みを一括更新
func (r *UserPreferenceRepositoryImpl) UpdateUserPreferences(ctx context.Context, userID int, preferences []*entities.UserPreference) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 既存の好みを削除
	_, err = tx.Exec(ctx, `DELETE FROM "UserPreference" WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// 新しい好みを挿入
	for _, preference := range preferences {
		_, err = tx.Exec(ctx, `
			INSERT INTO "UserPreference" (user_id, category_id, score, updated_at)
			VALUES ($1, $2, $3, NOW())
		`, preference.UserID, preference.CategoryID, preference.Score)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}