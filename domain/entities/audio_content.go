package entities

import "time"

// AudioContent 音声コンテンツのドメインエンティティ
type AudioContent struct {
	ID          int
	Title       string
	Description string
	CategoryID  int
	AuthorID    int
	Duration    int // 秒
	CreatedAt   time.Time
	PlayCount   int
	LikeCount   int
}

// IsValid コンテンツの妥当性をチェック
func (ac *AudioContent) IsValid() bool {
	return ac.ID > 0 && ac.Title != "" && ac.Duration > 0
}

// IsPopular 人気コンテンツかどうかを判定
func (ac *AudioContent) IsPopular() bool {
	return ac.PlayCount > 100 || ac.LikeCount > 50
}

// IsNew 新しいコンテンツかどうかを判定
func (ac *AudioContent) IsNew() bool {
	return time.Since(ac.CreatedAt) <= 7*24*time.Hour // 7日以内
}

// CalculatePopularityScore 人気度スコアを計算
func (ac *AudioContent) CalculatePopularityScore() float64 {
	playScore := float64(ac.PlayCount) * 0.7
	likeScore := float64(ac.LikeCount) * 1.5
	
	// 新しいコンテンツには追加のスコア
	var recencyBonus float64
	if ac.IsNew() {
		recencyBonus = 10.0
	}
	
	return playScore + likeScore + recencyBonus
}