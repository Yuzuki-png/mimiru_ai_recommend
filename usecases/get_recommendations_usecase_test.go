package usecases

import (
	"context"
	"errors"
	"mimiru-ai/domain/entities"
	"testing"
	"time"
)

// モックリポジトリとサービス
type mockCacheRepository struct {
	data map[string]interface{}
	err  error
}

func (m *mockCacheRepository) Get(ctx context.Context, key string, dest interface{}) error {
	if m.err != nil {
		return m.err
	}
	if data, exists := m.data[key]; exists {
		if output, ok := data.(*GetRecommendationsOutput); ok {
			*dest.(*GetRecommendationsOutput) = *output
			return nil
		}
	}
	return errors.New("キャッシュミス")
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.err != nil {
		return m.err
	}
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
	return nil
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	if m.data != nil {
		delete(m.data, key)
	}
	return nil
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	if m.data == nil {
		return false, nil
	}
	_, exists := m.data[key]
	return exists, nil
}

type mockUserRepository struct {
	user *entities.User
	err  error
}

func (m *mockUserRepository) GetByID(ctx context.Context, userID int) (*entities.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}

func (m *mockUserRepository) GetSimilarUsers(ctx context.Context, userID int, limit int) ([]*entities.User, error) {
	return []*entities.User{}, nil
}

func (m *mockUserRepository) Save(ctx context.Context, user *entities.User) error {
	return nil
}

type mockRecommendationAlgorithmService struct {
	collaborative []*entities.Recommendation
	contentBased  []*entities.Recommendation
	popular       []*entities.Recommendation
	newContent    []*entities.Recommendation
	err           error
}

func (m *mockRecommendationAlgorithmService) GenerateCollaborativeRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.collaborative, nil
}

func (m *mockRecommendationAlgorithmService) GenerateContentBasedRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.contentBased, nil
}

func (m *mockRecommendationAlgorithmService) GeneratePopularityBasedRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.popular, nil
}

func (m *mockRecommendationAlgorithmService) GenerateNewContentRecommendations(ctx context.Context, userID int, limit int) ([]*entities.Recommendation, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.newContent, nil
}

func TestGetRecommendationsUsecase_Execute_WithCache(t *testing.T) {
	// キャッシュされたレスポンス
	cachedOutput := &GetRecommendationsOutput{
		UserID: 123,
		Recommendations: []*entities.Recommendation{
			{
				UserID:         123,
				AudioContentID: 1,
				Score:          4.5,
				Reason:         entities.ReasonSimilarUsers,
			},
		},
		Timestamp: time.Now().Unix(),
	}

	mockCache := &mockCacheRepository{
		data: map[string]interface{}{
			"recommendations:user:123": cachedOutput,
		},
	}

	mockUser := &mockUserRepository{
		user: &entities.User{ID: 123, Email: "test@example.com"},
	}

	mockAlgorithm := &mockRecommendationAlgorithmService{}

	usecase := NewGetRecommendationsUsecase(
		mockAlgorithm,
		mockCache,
		mockUser,
	)

	input := &GetRecommendationsInput{
		UserID: 123,
		Limit:  20,
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("エラーがないことを期待しましたが、%vを取得しました", err)
	}

	if output.UserID != 123 {
		t.Errorf("ユーザーID 123を期待しましたが、%dを取得しました", output.UserID)
	}

	if len(output.Recommendations) != 1 {
		t.Errorf("1件のレコメンドを期待しましたが、%d件を取得しました", len(output.Recommendations))
	}
}

func TestGetRecommendationsUsecase_Execute_InvalidInput(t *testing.T) {
	mockCache := &mockCacheRepository{}
	mockUser := &mockUserRepository{}
	mockAlgorithm := &mockRecommendationAlgorithmService{}

	usecase := NewGetRecommendationsUsecase(
		mockAlgorithm,
		mockCache,
		mockUser,
	)

	// 無効なユーザーID
	input := &GetRecommendationsInput{
		UserID: 0,
		Limit:  20,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Error("無効なユーザーIDに対するエラーを期待しましたが、エラーがありませんでした")
	}
}

func TestGetRecommendationsUsecase_Execute_UserNotFound(t *testing.T) {
	mockCache := &mockCacheRepository{
		err: errors.New("キャッシュミス"),
	}

	mockUser := &mockUserRepository{
		user: nil, // ユーザーが見つからない
	}

	mockAlgorithm := &mockRecommendationAlgorithmService{}

	usecase := NewGetRecommendationsUsecase(
		mockAlgorithm,
		mockCache,
		mockUser,
	)

	input := &GetRecommendationsInput{
		UserID: 999,
		Limit:  20,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Error("ユーザーが見つからない場合のエラーを期待しましたが、エラーがありませんでした")
	}
}

func TestGetRecommendationsUsecase_Execute_GenerateRecommendations(t *testing.T) {
	mockCache := &mockCacheRepository{
		err: errors.New("キャッシュミス"),
	}

	mockUser := &mockUserRepository{
		user: &entities.User{ID: 123, Email: "test@example.com"},
	}

	mockAlgorithm := &mockRecommendationAlgorithmService{
		collaborative: []*entities.Recommendation{
			{UserID: 123, AudioContentID: 1, Score: 4.0, Reason: entities.ReasonSimilarUsers},
		},
		contentBased: []*entities.Recommendation{
			{UserID: 123, AudioContentID: 2, Score: 3.5, Reason: entities.ReasonContentBased},
		},
		popular: []*entities.Recommendation{
			{UserID: 123, AudioContentID: 3, Score: 3.0, Reason: entities.ReasonPopular},
		},
		newContent: []*entities.Recommendation{
			{UserID: 123, AudioContentID: 4, Score: 2.5, Reason: entities.ReasonNewContent},
		},
	}

	usecase := NewGetRecommendationsUsecase(
		mockAlgorithm,
		mockCache,
		mockUser,
	)

	input := &GetRecommendationsInput{
		UserID: 123,
		Limit:  20,
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Errorf("エラーがないことを期待しましたが、%vを取得しました", err)
	}

	if output.UserID != 123 {
		t.Errorf("ユーザーID 123を期待しましたが、%dを取得しました", output.UserID)
	}

	if len(output.Recommendations) != 4 {
		t.Errorf("4件のレコメンドを期待しましたが、%d件を取得しました", len(output.Recommendations))
	}

	// スコア順にソートされているかチェック
	for i := 0; i < len(output.Recommendations)-1; i++ {
		if output.Recommendations[i].Score < output.Recommendations[i+1].Score {
			t.Error("レコメンドがスコアの降順でソートされていません")
		}
	}
}