package services

import (
	"context"
	"fmt"
	"mimiru-ai/domain/repositories"
	"strconv"
)

type RecommendationUpdaterService struct {
	cacheRepo repositories.CacheRepository
}

func NewRecommendationUpdaterService(cacheRepo repositories.CacheRepository) *RecommendationUpdaterService {
	return &RecommendationUpdaterService{
		cacheRepo: cacheRepo,
	}
}

func (s *RecommendationUpdaterService) HandlePlaybackEvent(event DatabaseEvent) {

	userID, ok := event.Data["user_id"].(int)
	if !ok {
		if userIDFloat, ok := event.Data["user_id"].(float64); ok {
			userID = int(userIDFloat)
		} else if userIDStr, ok := event.Data["user_id"].(string); ok {
			var err error
			userID, err = strconv.Atoi(userIDStr)
			if err != nil {
				return
			}
		} else {
			return
		}
	}

	cacheKey := fmt.Sprintf("recommendations:user:%d", userID)
	ctx := context.Background()
	
	s.cacheRepo.Delete(ctx, cacheKey)

	s.invalidateRelatedUserCaches(ctx, userID)
}

func (s *RecommendationUpdaterService) HandleUserRatingEvent(event DatabaseEvent) {

	s.HandlePlaybackEvent(event)
}

func (s *RecommendationUpdaterService) invalidateRelatedUserCaches(ctx context.Context, userID int) {
	
	
	
}

func (s *RecommendationUpdaterService) StartRecommendationUpdater(
	ctx context.Context,
	monitorService *DatabaseMonitorService,
) {
	monitorService.RegisterEventHandler("playback_sessions", s.HandlePlaybackEvent)
	
	monitorService.RegisterEventHandler("user_ratings", s.HandleUserRatingEvent)
	
}