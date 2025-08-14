package entities

import "time"

// User ドメインエンティティ
type User struct {
	ID               int
	Email            string
	CreatedAt        time.Time
	PreferredCategories []int
}

// IsValid ユーザーの妥当性をチェック
func (u *User) IsValid() bool {
	return u.ID > 0 && u.Email != ""
}

// HasPreferenceFor 特定のカテゴリを好むかチェック
func (u *User) HasPreferenceFor(categoryID int) bool {
	for _, prefID := range u.PreferredCategories {
		if prefID == categoryID {
			return true
		}
	}
	return false
}