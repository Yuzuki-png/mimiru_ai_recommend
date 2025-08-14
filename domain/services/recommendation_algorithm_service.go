package services

import (
	"context"
	"mimiru-ai/domain/entities"
	"mimiru-ai/domain/repositories"
)

// RecommendationAlgorithmService レコメンドアルゴリズムのドメインサービス
type RecommendationAlgorithmService struct {
	userRepo         repositories.UserRepository
	audioContentRepo repositories.AudioContentRepository
	playbackRepo     repositories.PlaybackRepository
	userPrefRepo     repositories.UserPreferenceRepository
}

// NewRecommendationAlgorithmService コンストラクタ
func NewRecommendationAlgorithmService(
	userRepo repositories.UserRepository,
	audioContentRepo repositories.AudioContentRepository,
	playbackRepo repositories.PlaybackRepository,
	userPrefRepo repositories.UserPreferenceRepository,
) *RecommendationAlgorithmService {
	return &RecommendationAlgorithmService{
		userRepo:         userRepo,
		audioContentRepo: audioContentRepo,
		playbackRepo:     playbackRepo,
		userPrefRepo:     userPrefRepo,
	}
}

// GenerateCollaborativeRecommendations 協調フィルタリングによるレコメンド生成
func (s *RecommendationAlgorithmService) GenerateCollaborativeRecommendations(
	ctx context.Context,
	targetUserID int,
	limit int,
) ([]*entities.Recommendation, error) {
	// 類似ユーザーを取得
	similarUsers, err := s.userRepo.GetSimilarUsers(ctx, targetUserID, 10)
	if err != nil {
		return nil, err
	}

	if len(similarUsers) == 0 {
		return []*entities.Recommendation{}, nil
	}

	// 対象ユーザーの既視聴コンテンツを取得
	userHistory, err := s.playbackRepo.GetUserHistory(ctx, targetUserID, 100)
	if err != nil {
		return nil, err
	}

	watchedContent := make(map[int]bool)
	for _, history := range userHistory {
		watchedContent[history.AudioContentID] = true
	}

	// 類似ユーザーの視聴履歴を分析
	contentScores := make(map[int]float64)
	for _, similarUser := range similarUsers {
		history, err := s.playbackRepo.GetUserHistory(ctx, similarUser.ID, 20)
		if err != nil {
			continue
		}

		for _, playback := range history {
			if watchedContent[playback.AudioContentID] {
				continue // 既に視聴済み
			}

			score := playback.CalculateEngagementScore()
			contentScores[playback.AudioContentID] += score
		}
	}

	// レコメンドを生成
	var recommendations []*entities.Recommendation
	for contentID, score := range contentScores {
		if score >= 2.0 { // 最低スコア閾値
			recommendation := &entities.Recommendation{
				UserID:         targetUserID,
				AudioContentID: contentID,
				Score:          score * 0.4, // 協調フィルタリングの重み
				Reason:         entities.ReasonSimilarUsers,
			}
			recommendations = append(recommendations, recommendation)
		}

		if len(recommendations) >= limit {
			break
		}
	}

	return recommendations, nil
}

// GenerateContentBasedRecommendations コンテンツベースレコメンド生成
func (s *RecommendationAlgorithmService) GenerateContentBasedRecommendations(
	ctx context.Context,
	userID int,
	limit int,
) ([]*entities.Recommendation, error) {
	// ユーザーの好みを取得
	preferences, err := s.userPrefRepo.GetUserPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(preferences) == 0 {
		return []*entities.Recommendation{}, nil
	}

	// 視聴済みコンテンツを取得
	history, err := s.playbackRepo.GetUserHistory(ctx, userID, 50)
	if err != nil {
		return nil, err
	}

	excludeIDs := make([]int, len(history))
	for i, h := range history {
		excludeIDs[i] = h.AudioContentID
	}

	var recommendations []*entities.Recommendation
	
	// 各好みカテゴリから類似コンテンツを取得
	for _, preference := range preferences {
		if !preference.IsStrong() {
			continue // 強い好みのみ対象
		}

		similarContent, err := s.audioContentRepo.GetSimilarContent(
			ctx, 
			preference.CategoryID, 
			0, 
			excludeIDs, 
			5,
		)
		if err != nil {
			continue
		}

		for _, content := range similarContent {
			score := preference.Score * 0.3 // コンテンツベースの重み
			recommendation := &entities.Recommendation{
				UserID:         userID,
				AudioContentID: content.ID,
				Score:          score,
				Reason:         entities.ReasonContentBased,
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	return recommendations, nil
}

// GeneratePopularityBasedRecommendations 人気度ベースレコメンド生成
func (s *RecommendationAlgorithmService) GeneratePopularityBasedRecommendations(
	ctx context.Context,
	userID int,
	limit int,
) ([]*entities.Recommendation, error) {
	// 人気コンテンツを取得
	popularContent, err := s.audioContentRepo.GetPopularContent(ctx, 7, limit+10)
	if err != nil {
		return nil, err
	}

	// 視聴済みコンテンツを除外
	history, err := s.playbackRepo.GetUserHistory(ctx, userID, 100)
	if err != nil {
		return nil, err
	}

	watchedContent := make(map[int]bool)
	for _, h := range history {
		watchedContent[h.AudioContentID] = true
	}

	var recommendations []*entities.Recommendation
	for i, content := range popularContent {
		if watchedContent[content.ID] {
			continue
		}

		// 順位に応じてスコア調整
		score := float64(limit-i) * 0.2 // 人気度の重み
		recommendation := &entities.Recommendation{
			UserID:         userID,
			AudioContentID: content.ID,
			Score:          score,
			Reason:         entities.ReasonPopular,
		}
		recommendations = append(recommendations, recommendation)

		if len(recommendations) >= limit {
			break
		}
	}

	return recommendations, nil
}

// GenerateNewContentRecommendations 新着コンテンツレコメンド生成
func (s *RecommendationAlgorithmService) GenerateNewContentRecommendations(
	ctx context.Context,
	userID int,
	limit int,
) ([]*entities.Recommendation, error) {
	// 新着コンテンツを取得
	newContent, err := s.audioContentRepo.GetNewContent(ctx, 3, limit)
	if err != nil {
		return nil, err
	}

	var recommendations []*entities.Recommendation
	for _, content := range newContent {
		recommendation := &entities.Recommendation{
			UserID:         userID,
			AudioContentID: content.ID,
			Score:          0.1, // 新着の重み
			Reason:         entities.ReasonNewContent,
		}
		recommendations = append(recommendations, recommendation)
	}

	return recommendations, nil
}