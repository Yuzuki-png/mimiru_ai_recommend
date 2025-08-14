package controllers

import (
	"context"
	"mimiru-ai/common"
	"mimiru-ai/infrastructure/database"
	"mimiru-ai/usecases"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RecommendationController struct {
	getRecommendationsUC *usecases.GetRecommendationsUsecase
	db                   *database.Client
}

func NewRecommendationController(
	getRecommendationsUC *usecases.GetRecommendationsUsecase,
	db *database.Client,
) *RecommendationController {
	return &RecommendationController{
		getRecommendationsUC: getRecommendationsUC,
		db:                   db,
	}
}

func (c *RecommendationController) GetRecommendations(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		common.RespondWithError(ctx, common.ErrUserIDRequired)
		return
	}
	
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		appErr := common.NewBadRequestError("ユーザーIDの形式が正しくありません", err.Error())
		common.RespondWithError(ctx, appErr)
		return
	}

	limitStr := ctx.Query("limit")
	limit := 20
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	input := &usecases.GetRecommendationsInput{
		UserID: userID,
		Limit:  limit,
	}

	output, err := c.getRecommendationsUC.Execute(ctx.Request.Context(), input)
	if err != nil {
		appErr := common.NewInternalServerError("レコメンド取得に失敗しました", err.Error())
		common.RespondWithError(ctx, appErr)
		return
	}

	common.RespondWithSuccess(ctx, output)
}


func (c *RecommendationController) HealthCheck(ctx *gin.Context) {
	if err := c.db.Pool.Ping(context.Background()); err != nil {
		appErr := common.NewServiceUnavailableError("データベース接続に失敗しました", err.Error())
		common.RespondWithError(ctx, appErr)
		return
	}

	healthData := gin.H{
		"status":   "正常",
		"database": "接続中",
		"version":  "2.0.0-clean-arch",
	}

	common.RespondWithSuccess(ctx, healthData)
}