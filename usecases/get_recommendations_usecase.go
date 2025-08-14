package usecases

import (
	"context"
	"fmt"
	"mimiru-ai/domain/entities"
	"mimiru-ai/domain/repositories"
	"time"
)

// RecommendationAlgorithmServiceInterface ドメインサービスのインターフェース
type RecommendationAlgorithmServiceInterface interface {
	GenerateCollaborativeRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error)
	GenerateContentBasedRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error)
	GeneratePopularityBasedRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error)
	GenerateNewContentRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error)
}

// GetRecommendationsInput レコメンド取得の入力
type GetRecommendationsInput struct {
	UserID int
	Limit  int
}

// GetRecommendationsOutput レコメンド取得の出力
type GetRecommendationsOutput struct {
	UserID          int                       `json:"userId"`
	Recommendations []*entities.Recommendation `json:"recommendations"`
	Timestamp       int64                     `json:"timestamp"`
}

// GetRecommendationsUsecase レコメンド取得ユースケース
type GetRecommendationsUsecase struct {
	algorithmService RecommendationAlgorithmServiceInterface
	cacheRepo        repositories.CacheRepository
	userRepo         repositories.UserRepository
}

// NewGetRecommendationsUsecase コンストラクタ
func NewGetRecommendationsUsecase(
	algorithmService RecommendationAlgorithmServiceInterface,
	cacheRepo repositories.CacheRepository,
	userRepo repositories.UserRepository,
) *GetRecommendationsUsecase {
	return &GetRecommendationsUsecase{
		algorithmService: algorithmService,
		cacheRepo:        cacheRepo,
		userRepo:         userRepo,
	}
}

// Execute ユースケース実行
func (uc *GetRecommendationsUsecase) Execute(ctx context.Context, input *GetRecommendationsInput) (*GetRecommendationsOutput, error) {
	// 入力検証
	if input.UserID <= 0 {
		return nil, fmt.Errorf("無効なユーザーID: %d", input.UserID)
	}
	
	if input.Limit <= 0 {
		input.Limit = 20 // デフォルト値
	}

	// ユーザーの存在確認
	user, err := uc.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("ユーザー情報の取得に失敗しました: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("ユーザーが見つかりません: %d", input.UserID)
	}

	// キャッシュ確認
	cacheKey := fmt.Sprintf("recommendations:user:%d", input.UserID)
	var cachedOutput GetRecommendationsOutput
	if err := uc.cacheRepo.Get(ctx, cacheKey, &cachedOutput); err == nil {
		return &cachedOutput, nil
	}

	// レコメンド生成セット作成
	recSet := &entities.RecommendationSet{
		UserID:      input.UserID,
		GeneratedAt: time.Now(),
	}

	// 1. 協調フィルタリング (40%)
	collaborative, err := uc.algorithmService.GenerateCollaborativeRecommendations(
		ctx, 
		input.UserID, 
		input.Limit/2,
	)
	if err == nil {
		for _, rec := range collaborative {
			rec.GeneratedAt = time.Now()
			recSet.AddRecommendation(rec)
		}
	}

	// 2. コンテンツベース (30%)
	contentBased, err := uc.algorithmService.GenerateContentBasedRecommendations(
		ctx, 
		input.UserID, 
		input.Limit/3,
	)
	if err == nil {
		for _, rec := range contentBased {
			rec.GeneratedAt = time.Now()
			recSet.AddRecommendation(rec)
		}
	}

	// 3. 人気度ベース (20%)
	popular, err := uc.algorithmService.GeneratePopularityBasedRecommendations(
		ctx, 
		input.UserID, 
		input.Limit/5,
	)
	if err == nil {
		for _, rec := range popular {
			rec.GeneratedAt = time.Now()
			recSet.AddRecommendation(rec)
		}
	}

	// 4. 新着コンテンツ (10%)
	newContent, err := uc.algorithmService.GenerateNewContentRecommendations(
		ctx, 
		input.UserID, 
		input.Limit/10,
	)
	if err == nil {
		for _, rec := range newContent {
			rec.GeneratedAt = time.Now()
			recSet.AddRecommendation(rec)
		}
	}

	// 重複除去とソート
	recSet.SortByScore()
	finalRecommendations := recSet.Limit(input.Limit)

	// 結果作成
	output := &GetRecommendationsOutput{
		UserID:          input.UserID,
		Recommendations: finalRecommendations,
		Timestamp:       time.Now().Unix(),
	}

	// キャッシュに保存
	if err := uc.cacheRepo.Set(ctx, cacheKey, output, time.Hour); err != nil {
		// ログ出力のみで続行
	}

	return output, nil
}