package main

import (
	"context"
	"log"
	"mimiru-ai/controllers"
	"mimiru-ai/domain/repositories"
	"mimiru-ai/domain/services"
	"mimiru-ai/infrastructure/cache"
	"mimiru-ai/infrastructure/database"
	infraRepos "mimiru-ai/infrastructure/repositories"
	"mimiru-ai/usecases"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type DIContainer struct {
	db          *database.Client
	cacheClient *cache.Client

	userRepo         repositories.UserRepository
	userPrefRepo     repositories.UserPreferenceRepository
	audioContentRepo repositories.AudioContentRepository
	playbackRepo     repositories.PlaybackRepository
	cacheRepo        repositories.CacheRepository

	algorithmService      *services.RecommendationAlgorithmService
	monitorService        *services.DatabaseMonitorService
	recommendationUpdater *services.RecommendationUpdaterService

	getRecommendationsUC *usecases.GetRecommendationsUsecase

	recommendationController *controllers.RecommendationController
}

func NewDIContainer() (*DIContainer, error) {
	container := &DIContainer{}

	if err := container.initInfrastructure(); err != nil {
		return nil, err
	}

	container.initRepositories()

	container.initDomainServices()

	container.initUsecases()

	container.initControllers()

	return container, nil
}

func (c *DIContainer) initInfrastructure() error {
	db, err := database.NewPostgresClient()
	if err != nil {
		return err
	}
	c.db = db

	c.cacheClient = cache.NewRedisClient()

	return nil
}

func (c *DIContainer) initRepositories() {
	c.userRepo = infraRepos.NewUserRepositoryImpl(c.db)
	c.userPrefRepo = infraRepos.NewUserPreferenceRepositoryImpl(c.db)
	c.audioContentRepo = infraRepos.NewAudioContentRepositoryImpl(c.db)
	c.playbackRepo = infraRepos.NewPlaybackRepositoryImpl(c.db)
	c.cacheRepo = infraRepos.NewCacheRepositoryImpl(c.cacheClient)
}

func (c *DIContainer) initDomainServices() {
	c.algorithmService = services.NewRecommendationAlgorithmService(
		c.userRepo,
		c.audioContentRepo,
		c.playbackRepo,
		c.userPrefRepo,
	)

	c.monitorService = services.NewDatabaseMonitorService(c.db)
	c.recommendationUpdater = services.NewRecommendationUpdaterService(c.cacheRepo)
}

func (c *DIContainer) initUsecases() {
	c.getRecommendationsUC = usecases.NewGetRecommendationsUsecase(
		c.algorithmService,
		c.cacheRepo,
		c.userRepo,
	)

}

func (c *DIContainer) initControllers() {
	c.recommendationController = controllers.NewRecommendationController(
		c.getRecommendationsUC,
		c.db,
	)
}

func (c *DIContainer) Close() {
	if c.db != nil {
		c.db.Close()
	}
	if c.cacheClient != nil {
		c.cacheClient.Close()
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		// .envファイルが見つからないため、環境変数を使用
	}

	container, err := NewDIContainer()
	if err != nil {
		log.Fatal("DIコンテナの初期化に失敗しました:", err)
	}
	defer container.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET("/health", container.recommendationController.HealthCheck)
	r.GET("/recommendations", container.recommendationController.GetRecommendations)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	monitorCtx, cancelMonitor := context.WithCancel(context.Background())
	go func() {
		container.recommendationUpdater.StartRecommendationUpdater(monitorCtx, container.monitorService)

		container.monitorService.PollingMonitor(monitorCtx, 30*time.Second)
	}()

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("リッスンエラー: %s\n", err)
		}
	}()


	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	cancelMonitor()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("サーバーが強制的にシャットダウンされました:", err)
	}

}
