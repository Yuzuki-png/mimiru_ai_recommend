package entities

import "time"

// RecommendationReason レコメンドの理由
type RecommendationReason string

const (
	ReasonSimilarUsers   RecommendationReason = "similar_users"
	ReasonContentBased   RecommendationReason = "content_based"
	ReasonPopular        RecommendationReason = "popular"
	ReasonNewContent     RecommendationReason = "new_content"
)

// Recommendation レコメンドエンティティ
type Recommendation struct {
	UserID         int
	AudioContentID int
	Score          float64
	Reason         RecommendationReason
	GeneratedAt    time.Time
}

// IsValid レコメンドの妥当性をチェック
func (r *Recommendation) IsValid() bool {
	return r.UserID > 0 && r.AudioContentID > 0 && r.Score > 0
}

// IsHighQuality 高品質なレコメンドかどうか判定
func (r *Recommendation) IsHighQuality() bool {
	return r.Score >= 3.0
}

// RecommendationSet レコメンドセット
type RecommendationSet struct {
	UserID          int
	Recommendations []*Recommendation
	GeneratedAt     time.Time
}

// AddRecommendation レコメンドを追加
func (rs *RecommendationSet) AddRecommendation(rec *Recommendation) {
	if rec.IsValid() {
		rs.Recommendations = append(rs.Recommendations, rec)
	}
}

// SortByScore スコア順にソート
func (rs *RecommendationSet) SortByScore() {
	for i := 0; i < len(rs.Recommendations); i++ {
		for j := i + 1; j < len(rs.Recommendations); j++ {
			if rs.Recommendations[i].Score < rs.Recommendations[j].Score {
				rs.Recommendations[i], rs.Recommendations[j] = rs.Recommendations[j], rs.Recommendations[i]
			}
		}
	}
}

// FilterHighQuality 高品質なレコメンドのみを取得
func (rs *RecommendationSet) FilterHighQuality() []*Recommendation {
	var highQuality []*Recommendation
	for _, rec := range rs.Recommendations {
		if rec.IsHighQuality() {
			highQuality = append(highQuality, rec)
		}
	}
	return highQuality
}

// Limit 指定した数だけレコメンドを取得
func (rs *RecommendationSet) Limit(count int) []*Recommendation {
	if len(rs.Recommendations) <= count {
		return rs.Recommendations
	}
	return rs.Recommendations[:count]
}