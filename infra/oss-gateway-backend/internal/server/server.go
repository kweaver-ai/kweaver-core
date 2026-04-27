package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"oss-gateway/internal/cache"
	"oss-gateway/internal/config"
	"oss-gateway/internal/handler"
	"oss-gateway/internal/repository"
	"oss-gateway/internal/router"
	"oss-gateway/internal/service"
	"oss-gateway/pkg/crypto"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Server struct {
	config *config.AppConfig
	log    *logrus.Entry
	db     *gorm.DB
	crypto *crypto.AESCrypto
}

func NewServer(cfg *config.AppConfig, log *logrus.Entry, db *gorm.DB, crypto *crypto.AESCrypto) *Server {
	return &Server{
		config: cfg,
		log:    log,
		db:     db,
		crypto: crypto,
	}
}

func (s *Server) Start() {
	storageRepo := repository.NewStorageRepository(s.db)
	uploadTaskRepo := repository.NewUploadTaskRepository(s.db)

	redisClient, err := cache.NewRedisClient(s.config, s.log.WithField("module", "redis"))
	if err != nil {
		s.log.WithError(err).Fatal("Failed to connect to Redis")
	}

	storageCache := cache.NewStorageCache(redisClient)

	storageService := service.NewStorageService(storageRepo, s.crypto, storageCache, s.config, s.log.WithField("module", "storage_service"))
	urlService := service.NewURLService(storageService, uploadTaskRepo, s.config, s.log.WithField("module", "url_service"))

	// 启动时清理超过7天未完成的上传任务
	s.cleanExpiredTasksOnStartup(urlService)

	healthHandler := handler.NewHealthHandler(s.db, redisClient)
	storageHandler := handler.NewStorageHandler(storageService)
	uploadHandler := handler.NewUploadHandler(urlService)
	downloadHandler := handler.NewDownloadHandler(urlService)
	deleteHandler := handler.NewDeleteHandler(urlService)
	headHandler := handler.NewHeadHandler(urlService)

	r := router.SetupRouter(&router.RouterConfig{
		Logger:          s.log,
		HealthHandler:   healthHandler,
		StorageHandler:  storageHandler,
		UploadHandler:   uploadHandler,
		DownloadHandler: downloadHandler,
		DeleteHandler:   deleteHandler,
		HeadHandler:     headHandler,
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.CommonConfig.Port),
		Handler: r,
	}

	go func() {
		s.log.Infof("Starting server on port %s", s.config.CommonConfig.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.WithError(err).Fatal("Failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		s.log.WithError(err).Error("Server forced to shutdown")
	}

	s.log.Info("Server exited")
}

// cleanExpiredTasksOnStartup 在服务启动时清理超过7天未完成的上传任务
func (s *Server) cleanExpiredTasksOnStartup(urlService service.URLService) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.log.Info("Cleaning expired upload tasks on startup (tasks older than 7 days)...")

	if err := urlService.CleanExpiredTasks(ctx); err != nil {
		s.log.WithError(err).Error("Failed to clean expired upload tasks on startup")
	} else {
		s.log.Info("Successfully cleaned expired upload tasks on startup")
	}
}
