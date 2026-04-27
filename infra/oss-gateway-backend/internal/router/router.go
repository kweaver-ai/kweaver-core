package router

import (
	"oss-gateway/internal/handler"
	"oss-gateway/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type RouterConfig struct {
	Logger          *logrus.Entry
	HealthHandler   *handler.HealthHandler
	StorageHandler  *handler.StorageHandler
	UploadHandler   *handler.UploadHandler
	DownloadHandler *handler.DownloadHandler
	DeleteHandler   *handler.DeleteHandler
	HeadHandler     *handler.HeadHandler
}

func SetupRouter(config *RouterConfig) *gin.Engine {
	r := gin.New()

	r.Use(middleware.Recovery(config.Logger))
	r.Use(middleware.Logger(config.Logger))
	r.Use(middleware.CORS())

	r.GET("/health/ready", config.HealthHandler.Ready)
	r.GET("/health/alive", config.HealthHandler.Alive)

	// API routes with fixed prefix
	api := r.Group("/api/v1")
	{
		storages := api.Group("/storages")
		{
			storages.GET("", config.StorageHandler.List)
			storages.GET("/:id", config.StorageHandler.Get)
			storages.POST("", config.StorageHandler.Create)
			storages.PUT("/:id", config.StorageHandler.Update)
			storages.DELETE("/:id", config.StorageHandler.Delete)
			storages.POST("/:id/check", config.StorageHandler.CheckConnection)
		}

		api.GET("/head/:storageId/*key", config.HeadHandler.GetHeadURL)
		api.POST("/head/:storageId", config.HeadHandler.BatchGetHeadURL)

		api.GET("/upload/:storageId/*key", config.UploadHandler.GetUploadURL)

		api.GET("/initmultiupload/:storageId/*key", config.UploadHandler.InitMultipartUpload)
		api.POST("/uploadpart/:storageId/*key", config.UploadHandler.GetUploadPartURLs)
		api.POST("/completeupload/:storageId/*key", config.UploadHandler.CompleteMultipartUpload)

		api.GET("/download/:storageId/*key", config.DownloadHandler.GetDownloadURL)

		api.GET("/delete/:storageId/*key", config.DeleteHandler.GetDeleteURL)
	}

	return r
}
