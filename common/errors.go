package common

import (
	"errors"
	"net/http"
)

type ErrorCode string

const (
	ErrorCodeBadRequest          ErrorCode = "BAD_REQUEST"
	ErrorCodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	ErrorCodeNotFound           ErrorCode = "NOT_FOUND"
	ErrorCodeInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrorCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
)

type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	HTTPStatus int       `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewBadRequestError(message string, details ...string) *AppError {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	return &AppError{
		Code:       ErrorCodeBadRequest,
		Message:    message,
		Details:    detail,
		HTTPStatus: http.StatusBadRequest,
	}
}

func NewNotFoundError(message string, details ...string) *AppError {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	return &AppError{
		Code:       ErrorCodeNotFound,
		Message:    message,
		Details:    detail,
		HTTPStatus: http.StatusNotFound,
	}
}

func NewInternalServerError(message string, details ...string) *AppError {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	return &AppError{
		Code:       ErrorCodeInternalServer,
		Message:    message,
		Details:    detail,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func NewServiceUnavailableError(message string, details ...string) *AppError {
	detail := ""
	if len(details) > 0 {
		detail = details[0]
	}
	return &AppError{
		Code:       ErrorCodeServiceUnavailable,
		Message:    message,
		Details:    detail,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

var (
	ErrUserIDRequired       = NewBadRequestError("ユーザーIDが必要です")
	ErrInvalidUserIDFormat  = NewBadRequestError("ユーザーIDの形式が正しくありません")
	ErrInvalidEventData     = NewBadRequestError("イベントデータが正しくありません")
	ErrUserNotFound         = NewNotFoundError("ユーザーが見つかりません")
	ErrRecommendationFailed = NewInternalServerError("レコメンド取得に失敗しました")
	ErrEventTrackingFailed  = NewInternalServerError("イベント追跡に失敗しました")
	ErrDatabaseConnection   = NewServiceUnavailableError("データベース接続に失敗しました")
)

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}