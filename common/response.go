package common

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ResponseWidget struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Message   string      `json:"message,omitempty"`
	Details   string      `json:"details,omitempty"`
	Code      ErrorCode   `json:"code,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Service   string      `json:"service"`
}

func NewSuccessResponse(data interface{}) *ResponseWidget {
	return &ResponseWidget{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Unix(),
		Service:   "mimiru-recommendation",
	}
}

func NewErrorResponse(appErr *AppError) *ResponseWidget {
	return &ResponseWidget{
		Success:   false,
		Error:     string(appErr.Code),
		Message:   appErr.Message,
		Details:   appErr.Details,
		Code:      appErr.Code,
		Timestamp: time.Now().Unix(),
		Service:   "mimiru-recommendation",
	}
}

func RespondWithSuccess(ctx *gin.Context, data interface{}) {
	response := NewSuccessResponse(data)
	ctx.JSON(http.StatusOK, response)
}

func RespondWithError(ctx *gin.Context, appErr *AppError) {
	response := NewErrorResponse(appErr)
	ctx.JSON(appErr.HTTPStatus, response)
}

func RespondWithAppError(ctx *gin.Context, err error) {
	if appErr, ok := IsAppError(err); ok {
		RespondWithError(ctx, appErr)
		return
	}
	
	internalErr := NewInternalServerError("内部サーバーエラー", err.Error())
	RespondWithError(ctx, internalErr)
}