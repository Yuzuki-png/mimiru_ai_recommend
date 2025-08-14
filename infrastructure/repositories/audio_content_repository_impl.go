package repositories

import (
	"context"
	"mimiru-ai/domain/entities"
	"mimiru-ai/domain/repositories"
	"mimiru-ai/infrastructure/database"
)

// AudioContentRepositoryImpl 音声コンテンツリポジトリの実装
type AudioContentRepositoryImpl struct {
	db *database.Client
}

// NewAudioContentRepositoryImpl コンストラクタ
func NewAudioContentRepositoryImpl(db *database.Client) repositories.AudioContentRepository {
	return &AudioContentRepositoryImpl{
		db: db,
	}
}

// GetByID IDで音声コンテンツを取得
func (r *AudioContentRepositoryImpl) GetByID(ctx context.Context, contentID int) (*entities.AudioContent, error) {
	query := `
		SELECT id, title, description, category_id, author_id, COALESCE(duration, 0), created_at,
			   COALESCE(play_count.count, 0) as play_count,
			   COALESCE(like_count.count, 0) as like_count
		FROM "AudioContent" ac
		LEFT JOIN (
			SELECT audio_content_id, COUNT(*) as count
			FROM "ListenHistory"
			GROUP BY audio_content_id
		) play_count ON ac.id = play_count.audio_content_id
		LEFT JOIN (
			SELECT content_id, COUNT(*) as count
			FROM "Like"
			GROUP BY content_id
		) like_count ON ac.id = like_count.content_id
		WHERE ac.id = $1
	`

	var content entities.AudioContent
	err := r.db.Pool.QueryRow(ctx, query, contentID).Scan(
		&content.ID,
		&content.Title,
		&content.Description,
		&content.CategoryID,
		&content.AuthorID,
		&content.Duration,
		&content.CreatedAt,
		&content.PlayCount,
		&content.LikeCount,
	)

	if err != nil {
		return nil, err
	}

	return &content, nil
}

// GetByIDs 複数IDで音声コンテンツを取得
func (r *AudioContentRepositoryImpl) GetByIDs(ctx context.Context, contentIDs []int) ([]*entities.AudioContent, error) {
	if len(contentIDs) == 0 {
		return []*entities.AudioContent{}, nil
	}

	query := `
		SELECT id, title, description, category_id, author_id, COALESCE(duration, 0), created_at,
			   COALESCE(play_count.count, 0) as play_count,
			   COALESCE(like_count.count, 0) as like_count
		FROM "AudioContent" ac
		LEFT JOIN (
			SELECT audio_content_id, COUNT(*) as count
			FROM "ListenHistory"
			GROUP BY audio_content_id
		) play_count ON ac.id = play_count.audio_content_id
		LEFT JOIN (
			SELECT content_id, COUNT(*) as count
			FROM "Like"
			GROUP BY content_id
		) like_count ON ac.id = like_count.content_id
		WHERE ac.id = ANY($1)
	`

	rows, err := r.db.Pool.Query(ctx, query, contentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*entities.AudioContent
	for rows.Next() {
		var content entities.AudioContent
		if err := rows.Scan(
			&content.ID,
			&content.Title,
			&content.Description,
			&content.CategoryID,
			&content.AuthorID,
			&content.Duration,
			&content.CreatedAt,
			&content.PlayCount,
			&content.LikeCount,
		); err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}

	return contents, rows.Err()
}

// GetSimilarContent 類似コンテンツを取得
func (r *AudioContentRepositoryImpl) GetSimilarContent(ctx context.Context, categoryID, authorID int, excludeIDs []int, limit int) ([]*entities.AudioContent, error) {
	query := `
		SELECT id, title, description, category_id, author_id, COALESCE(duration, 0), created_at,
			   0 as play_count, 0 as like_count
		FROM "AudioContent"
		WHERE (category_id = $1 OR author_id = $2)
		  AND id != ALL($3)
		  AND created_at > NOW() - INTERVAL '180 days'
		ORDER BY created_at DESC
		LIMIT $4
	`

	rows, err := r.db.Pool.Query(ctx, query, categoryID, authorID, excludeIDs, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*entities.AudioContent
	for rows.Next() {
		var content entities.AudioContent
		if err := rows.Scan(
			&content.ID,
			&content.Title,
			&content.Description,
			&content.CategoryID,
			&content.AuthorID,
			&content.Duration,
			&content.CreatedAt,
			&content.PlayCount,
			&content.LikeCount,
		); err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}

	return contents, rows.Err()
}

// GetNewContent 新着コンテンツを取得
func (r *AudioContentRepositoryImpl) GetNewContent(ctx context.Context, days int, limit int) ([]*entities.AudioContent, error) {
	query := `
		SELECT id, title, description, category_id, author_id, COALESCE(duration, 0), created_at,
			   0 as play_count, 0 as like_count
		FROM "AudioContent"
		WHERE created_at > NOW() - INTERVAL '3 days'
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*entities.AudioContent
	for rows.Next() {
		var content entities.AudioContent
		if err := rows.Scan(
			&content.ID,
			&content.Title,
			&content.Description,
			&content.CategoryID,
			&content.AuthorID,
			&content.Duration,
			&content.CreatedAt,
			&content.PlayCount,
			&content.LikeCount,
		); err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}

	return contents, rows.Err()
}

// GetPopularContent 人気コンテンツを取得
func (r *AudioContentRepositoryImpl) GetPopularContent(ctx context.Context, days int, limit int) ([]*entities.AudioContent, error) {
	query := `
		SELECT ac.id, ac.title, ac.description, ac.category_id, ac.author_id, 
			   COALESCE(ac.duration, 0), ac.created_at, COUNT(*) as play_count, 0 as like_count
		FROM "ListenHistory" lh
		JOIN "AudioContent" ac ON lh.audio_content_id = ac.id
		WHERE lh.created_at > NOW() - INTERVAL '7 days'
		GROUP BY ac.id, ac.title, ac.description, ac.category_id, ac.author_id, ac.duration, ac.created_at
		ORDER BY play_count DESC, ac.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contents []*entities.AudioContent
	for rows.Next() {
		var content entities.AudioContent
		if err := rows.Scan(
			&content.ID,
			&content.Title,
			&content.Description,
			&content.CategoryID,
			&content.AuthorID,
			&content.Duration,
			&content.CreatedAt,
			&content.PlayCount,
			&content.LikeCount,
		); err != nil {
			return nil, err
		}
		contents = append(contents, &content)
	}

	return contents, rows.Err()
}

// Save 音声コンテンツを保存
func (r *AudioContentRepositoryImpl) Save(ctx context.Context, content *entities.AudioContent) error {
	if !content.IsValid() {
		return ErrInvalidEntity
	}

	query := `
		INSERT INTO "AudioContent" (title, description, category_id, author_id, duration, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	err := r.db.Pool.QueryRow(ctx, query,
		content.Title,
		content.Description,
		content.CategoryID,
		content.AuthorID,
		content.Duration,
		content.CreatedAt,
	).Scan(&content.ID)

	return err
}