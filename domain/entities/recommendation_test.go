package entities

import (
	"testing"
	"time"
)

func TestRecommendation_IsValid(t *testing.T) {
	tests := []struct {
		name           string
		recommendation *Recommendation
		expected       bool
	}{
		{
			name: "有効なレコメンド",
			recommendation: &Recommendation{
				UserID:         1,
				AudioContentID: 100,
				Score:          4.5,
				Reason:         ReasonSimilarUsers,
			},
			expected: true,
		},
		{
			name: "無効なユーザーID",
			recommendation: &Recommendation{
				UserID:         0,
				AudioContentID: 100,
				Score:          4.5,
				Reason:         ReasonSimilarUsers,
			},
			expected: false,
		},
		{
			name: "無効なオーディオコンテンツID",
			recommendation: &Recommendation{
				UserID:         1,
				AudioContentID: 0,
				Score:          4.5,
				Reason:         ReasonSimilarUsers,
			},
			expected: false,
		},
		{
			name: "無効なスコア",
			recommendation: &Recommendation{
				UserID:         1,
				AudioContentID: 100,
				Score:          0,
				Reason:         ReasonSimilarUsers,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.recommendation.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, 期待値 %v", got, tt.expected)
			}
		})
	}
}

func TestRecommendation_IsHighQuality(t *testing.T) {
	tests := []struct {
		name           string
		recommendation *Recommendation
		expected       bool
	}{
		{
			name: "高品質のレコメンド",
			recommendation: &Recommendation{Score: 4.5},
			expected:       true,
		},
		{
			name: "低品質のレコメンド",
			recommendation: &Recommendation{Score: 2.5},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.recommendation.IsHighQuality(); got != tt.expected {
				t.Errorf("IsHighQuality() = %v, 期待値 %v", got, tt.expected)
			}
		})
	}
}

func TestRecommendationSet_AddRecommendation(t *testing.T) {
	recSet := &RecommendationSet{
		UserID:      1,
		GeneratedAt: time.Now(),
	}

	validRec := &Recommendation{
		UserID:         1,
		AudioContentID: 100,
		Score:          4.5,
		Reason:         ReasonSimilarUsers,
	}

	invalidRec := &Recommendation{
		UserID:         0, // 無効
		AudioContentID: 100,
		Score:          4.5,
		Reason:         ReasonSimilarUsers,
	}

	// 有効なレコメンドを追加
	recSet.AddRecommendation(validRec)
	if len(recSet.Recommendations) != 1 {
		t.Errorf("1件のレコメンドを期待しましたが、%d件を取得しました", len(recSet.Recommendations))
	}

	// 無効なレコメンドは追加されない
	recSet.AddRecommendation(invalidRec)
	if len(recSet.Recommendations) != 1 {
		t.Errorf("1件のレコメンドを期待しましたが、%d件を取得しました", len(recSet.Recommendations))
	}
}

func TestRecommendationSet_SortByScore(t *testing.T) {
	recSet := &RecommendationSet{
		UserID: 1,
		Recommendations: []*Recommendation{
			{Score: 2.0},
			{Score: 4.0},
			{Score: 1.0},
			{Score: 3.0},
		},
	}

	recSet.SortByScore()

	expectedScores := []float64{4.0, 3.0, 2.0, 1.0}
	for i, rec := range recSet.Recommendations {
		if rec.Score != expectedScores[i] {
			t.Errorf("インデックス %d: スコア %f を期待しましたが、%f を取得しました", i, expectedScores[i], rec.Score)
		}
	}
}

func TestRecommendationSet_FilterHighQuality(t *testing.T) {
	recSet := &RecommendationSet{
		UserID: 1,
		Recommendations: []*Recommendation{
			{Score: 2.0}, // 低品質
			{Score: 4.0}, // 高品質
			{Score: 1.0}, // 低品質
			{Score: 3.5}, // 高品質
		},
	}

	highQuality := recSet.FilterHighQuality()

	if len(highQuality) != 2 {
		t.Errorf("2件の高品質レコメンドを期待しましたが、%d件を取得しました", len(highQuality))
	}

	for _, rec := range highQuality {
		if !rec.IsHighQuality() {
			t.Errorf("高品質のレコメンドを期待しましたが、スコア %f を取得しました", rec.Score)
		}
	}
}

func TestRecommendationSet_Limit(t *testing.T) {
	recSet := &RecommendationSet{
		UserID: 1,
		Recommendations: []*Recommendation{
			{Score: 4.0},
			{Score: 3.0},
			{Score: 2.0},
			{Score: 1.0},
		},
	}

	// 制限数が配列の長さより少ない場合
	limited := recSet.Limit(2)
	if len(limited) != 2 {
		t.Errorf("2件のレコメンドを期待しましたが、%d件を取得しました", len(limited))
	}

	// 制限数が配列の長さより多い場合
	limited = recSet.Limit(10)
	if len(limited) != 4 {
		t.Errorf("4件のレコメンドを期待しましたが、%d件を取得しました", len(limited))
	}
}