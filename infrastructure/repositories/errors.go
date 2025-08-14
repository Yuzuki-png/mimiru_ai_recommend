package repositories

import "errors"

var (
	// ErrInvalidEntity 無効なエンティティエラー
	ErrInvalidEntity = errors.New("無効なエンティティ")
	
	// ErrNotFound エンティティが見つからないエラー
	ErrNotFound = errors.New("エンティティが見つかりません")
	
	// ErrDuplicateEntry 重複エラー
	ErrDuplicateEntry = errors.New("重複エントリ")
)