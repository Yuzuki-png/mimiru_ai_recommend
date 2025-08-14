package repositories

import (
	"context"
	"mimiru-ai/domain/entities"
	"mimiru-ai/domain/repositories"
	"mimiru-ai/infrastructure/database"
)

// PlaybackRepositoryImpl 再生履歴リポジトリの実装
type PlaybackRepositoryImpl struct {
	db *database.Client
}

// NewPlaybackRepositoryImpl コンストラクタ
func NewPlaybackRepositoryImpl(db *database.Client) repositories.PlaybackRepository {
	return &PlaybackRepositoryImpl{
		db: db,
	}
}

// GetUserHistory ユーザーの再生履歴を取得
func (r *PlaybackRepositoryImpl) GetUserHistory(ctx context.Context, userID int, limit int) ([]*entities.PlaybackHistory, error) {
	query := `
		SELECT lh.user_id, lh.audio_content_id, lh.created_at, 
			   COALESCE(lh.duration, 0) as duration, lh.completed
		FROM "ListenHistory" lh
		WHERE lh.user_id = $1
		ORDER BY lh.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*entities.PlaybackHistory
	for rows.Next() {
		var h entities.PlaybackHistory
		var duration float64
		if err := rows.Scan(
			&h.UserID,
			&h.AudioContentID,
			&h.PlayedAt,
			&duration,
			&h.Completed,
		); err != nil {
			return nil, err
		}
		h.Duration = int(duration)
		history = append(history, &h)
	}

	return history, rows.Err()
}

// SavePlayback 再生履歴を保存
func (r *PlaybackRepositoryImpl) SavePlayback(ctx context.Context, history *entities.PlaybackHistory) error {
	if !history.IsValid() {
		return ErrInvalidEntity
	}

	query := `
		INSERT INTO "ListenHistory" (user_id, audio_content_id, created_at, duration, completed)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		history.UserID,
		history.AudioContentID,
		history.PlayedAt,
		history.Duration,
		history.Completed,
	)

	return err
}

// GetRecentPlaybacks 最近の再生履歴を取得
func (r *PlaybackRepositoryImpl) GetRecentPlaybacks(ctx context.Context, userID int, days int) ([]*entities.PlaybackHistory, error) {
	query := `
		SELECT lh.user_id, lh.audio_content_id, lh.created_at, 
			   COALESCE(lh.duration, 0) as duration, lh.completed
		FROM "ListenHistory" lh
		WHERE lh.user_id = $1
		  AND lh.created_at > NOW() - INTERVAL '%d days'
		ORDER BY lh.created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*entities.PlaybackHistory
	for rows.Next() {
		var h entities.PlaybackHistory
		var duration float64
		if err := rows.Scan(
			&h.UserID,
			&h.AudioContentID,
			&h.PlayedAt,
			&duration,
			&h.Completed,
		); err != nil {
			return nil, err
		}
		h.Duration = int(duration)
		history = append(history, &h)
	}

	return history, rows.Err()
}