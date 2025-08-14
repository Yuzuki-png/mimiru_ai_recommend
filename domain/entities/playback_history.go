package entities

import "time"

// PlaybackHistory 再生履歴のドメインエンティティ
type PlaybackHistory struct {
	UserID         int
	AudioContentID int
	PlayedAt       time.Time
	Duration       int  // 実際の再生時間（秒）
	Completed      bool // 最後まで聞いたか
}

// IsValid 再生履歴の妥当性をチェック
func (ph *PlaybackHistory) IsValid() bool {
	return ph.UserID > 0 && ph.AudioContentID > 0 && !ph.PlayedAt.IsZero()
}

// IsRecentPlay 最近の再生かどうか判定
func (ph *PlaybackHistory) IsRecentPlay(days int) bool {
	return time.Since(ph.PlayedAt) <= time.Duration(days)*24*time.Hour
}

// CalculateEngagementScore エンゲージメントスコアを計算
func (ph *PlaybackHistory) CalculateEngagementScore() float64 {
	baseScore := 1.0
	
	if ph.Completed {
		baseScore *= 2.0 // 完了した場合は2倍
	} else if ph.Duration > 60 { // 1分以上聞いた場合
		baseScore *= 1.5
	}
	
	// 最近の再生は高スコア
	if ph.IsRecentPlay(7) {
		baseScore *= 1.2
	}
	
	return baseScore
}

// UserPreference ユーザーの好みを表現
type UserPreference struct {
	UserID     int
	CategoryID int
	Score      float64 // 0.0-1.0の好み度
	UpdatedAt  time.Time
}

// IsStrong 強い好みかどうか判定
func (up *UserPreference) IsStrong() bool {
	return up.Score >= 0.7
}

// IsWeak 弱い好みかどうか判定
func (up *UserPreference) IsWeak() bool {
	return up.Score < 0.3
}