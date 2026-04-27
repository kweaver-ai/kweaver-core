package service

import (
	"context"
	"fmt"
	"oss-gateway/internal/config"
	"oss-gateway/internal/model"
	"oss-gateway/internal/repository"
	"oss-gateway/pkg/adapter"
	"oss-gateway/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
)

type URLService interface {
	GetHeadURL(ctx context.Context, storageID, objectKey string, validSeconds int64, useInternal bool) (*adapter.PresignedURL, error)
	BatchGetHeadURL(ctx context.Context, storageID string, keys []string, validSeconds int64, useInternal bool) (map[string]*adapter.PresignedURL, error)
	GetUploadURL(ctx context.Context, storageID, objectKey string, requestMethod string, validSeconds int64, useInternal bool) (*adapter.PresignedURL, error)
	InitMultipartUpload(ctx context.Context, storageID, objectKey string, fileSize int64) (*InitMultipartResponse, error)
	GetUploadPartURLs(ctx context.Context, storageID, objectKey, uploadID string, partNumbers []int, validSeconds int64, useInternal bool) (map[int]*adapter.PresignedURL, error)
	CompleteMultipartUpload(ctx context.Context, storageID, objectKey, uploadID string, parts []adapter.PartInfo) (*adapter.PresignedURL, error)
	GetDownloadURL(ctx context.Context, storageID, objectKey string, validSeconds int64, saveName string, useInternal bool) (*adapter.PresignedURL, error)
	GetDeleteURL(ctx context.Context, storageID, objectKey string, validSeconds int64, useInternal bool) (*adapter.PresignedURL, error)
	CleanExpiredTasks(ctx context.Context) error
}

type urlService struct {
	storageService StorageService
	taskRepo       repository.UploadTaskRepository
	config         *config.AppConfig
	log            *logrus.Entry
}

type InitMultipartResponse struct {
	UploadID string `json:"upload_id"`
	PartSize int64  `json:"part_size"`
	Key      string `json:"key"`
}

func NewURLService(storageService StorageService, taskRepo repository.UploadTaskRepository, config *config.AppConfig, log *logrus.Entry) URLService {
	return &urlService{
		storageService: storageService,
		taskRepo:       taskRepo,
		config:         config,
		log:            log,
	}
}

func (s *urlService) GetHeadURL(ctx context.Context, storageID, objectKey string, validSeconds int64, useInternal bool) (*adapter.PresignedURL, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	if validSeconds <= 0 {
		validSeconds = s.config.OSSConfig.DefaultValidSeconds
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, useInternal)
	if err != nil {
		return nil, err
	}

	return ossAdapter.GetHeadURL(ctx, objectKey, validSeconds)
}

func (s *urlService) BatchGetHeadURL(ctx context.Context, storageID string, keys []string, validSeconds int64, useInternal bool) (map[string]*adapter.PresignedURL, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("keys cannot be empty")
	}

	if len(keys) > 100 {
		return nil, fmt.Errorf("too many keys, maximum 100 allowed")
	}

	if validSeconds <= 0 {
		validSeconds = s.config.OSSConfig.DefaultValidSeconds
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, useInternal)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*adapter.PresignedURL)
	for _, key := range keys {
		if err := utils.ValidateObjectKey(key); err != nil {
			s.log.WithError(err).Warnf("skip invalid key: %s", key)
			continue
		}

		url, err := ossAdapter.GetHeadURL(ctx, key, validSeconds)
		if err != nil {
			s.log.WithError(err).Warnf("failed to get head url for key: %s", key)
			continue
		}

		result[key] = url
	}

	return result, nil
}

func (s *urlService) GetUploadURL(ctx context.Context, storageID, objectKey string, requestMethod string, validSeconds int64, useInternal bool) (*adapter.PresignedURL, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	if validSeconds <= 0 {
		validSeconds = s.config.OSSConfig.DefaultValidSeconds
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, useInternal)
	if err != nil {
		return nil, err
	}

	if requestMethod == "PUT" {
		return ossAdapter.GetPutUploadURL(ctx, objectKey, validSeconds)
	}

	return ossAdapter.GetUploadURL(ctx, objectKey, validSeconds)
}

func (s *urlService) InitMultipartUpload(ctx context.Context, storageID, objectKey string, fileSize int64) (*InitMultipartResponse, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	if fileSize <= 0 {
		return nil, fmt.Errorf("file size must be greater than 0")
	}

	maxFileSize := int64(50) * 1024 * 1024 * 1024 * 1024
	if fileSize > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum limit (50TB)")
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, false)
	if err != nil {
		return nil, err
	}

	result, err := ossAdapter.InitMultipartUpload(ctx, objectKey, fileSize)
	if err != nil {
		return nil, err
	}

	totalParts := int((fileSize + result.PartSize - 1) / result.PartSize)

	// 生成雪花ID
	taskID := utils.GenerateStorageID()

	task := &model.MultipartUploadTask{
		ID:         taskID, // 使用雪花ID
		StorageID:  storageID,
		ObjectKey:  objectKey,
		UploadID:   result.UploadID,
		TotalSize:  fileSize,
		PartSize:   int(result.PartSize),
		TotalParts: totalParts,
		Status:     model.UploadStatusInProgress,
		ExpiresAt:  utils.CalculateExpiresAt(7 * 24 * 3600), // 7天过期，与预签名URL有效期保持一致
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		s.log.WithError(err).Warn("failed to save upload task")
	}

	return &InitMultipartResponse{
		UploadID: result.UploadID,
		PartSize: result.PartSize,
		Key:      objectKey,
	}, nil
}

func (s *urlService) GetUploadPartURLs(ctx context.Context, storageID, objectKey, uploadID string, partNumbers []int, validSeconds int64, useInternal bool) (map[int]*adapter.PresignedURL, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	if uploadID == "" {
		return nil, fmt.Errorf("upload_id cannot be empty")
	}

	if len(partNumbers) == 0 {
		return nil, fmt.Errorf("part_id cannot be empty")
	}

	if validSeconds <= 0 {
		validSeconds = s.config.OSSConfig.DefaultValidSeconds
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, useInternal)
	if err != nil {
		return nil, err
	}

	result := make(map[int]*adapter.PresignedURL)
	for _, partNumber := range partNumbers {
		if partNumber < 1 || partNumber > int(s.config.OSSConfig.MaxParts) {
			s.log.Warnf("invalid part number: %d", partNumber)
			continue
		}

		url, err := ossAdapter.GetUploadPartURL(ctx, objectKey, uploadID, partNumber, validSeconds)
		if err != nil {
			s.log.WithError(err).Warnf("failed to get upload part url for part: %d", partNumber)
			continue
		}

		result[partNumber] = url
	}

	return result, nil
}

func (s *urlService) CompleteMultipartUpload(ctx context.Context, storageID, objectKey, uploadID string, parts []adapter.PartInfo) (*adapter.PresignedURL, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	if uploadID == "" {
		return nil, fmt.Errorf("upload_id cannot be empty")
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("parts cannot be empty")
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, false)
	if err != nil {
		return nil, err
	}

	result, err := ossAdapter.GetCompleteMultipartUpload(ctx, objectKey, uploadID, parts)
	if err != nil {
		return nil, err
	}

	// 上传完成后立即删除任务记录
	task, err := s.taskRepo.GetByUploadID(ctx, storageID, objectKey, uploadID)
	if err == nil {
		if err := s.taskRepo.Delete(ctx, task.ID); err != nil {
			s.log.WithError(err).Warn("failed to delete completed upload task")
		} else {
			s.log.WithFields(logrus.Fields{
				"task_id":    task.ID,
				"storage_id": storageID,
				"object_key": objectKey,
			}).Info("deleted completed upload task")
		}
	}

	return result, nil
}

func (s *urlService) GetDownloadURL(ctx context.Context, storageID, objectKey string, validSeconds int64, saveName string, useInternal bool) (*adapter.PresignedURL, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	if validSeconds <= 0 {
		validSeconds = s.config.OSSConfig.DefaultValidSeconds
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, useInternal)
	if err != nil {
		return nil, err
	}

	return ossAdapter.GetDownloadURL(ctx, objectKey, validSeconds, saveName)
}

func (s *urlService) GetDeleteURL(ctx context.Context, storageID, objectKey string, validSeconds int64, useInternal bool) (*adapter.PresignedURL, error) {
	if err := utils.ValidateStorageID(storageID); err != nil {
		return nil, err
	}
	if err := utils.ValidateObjectKey(objectKey); err != nil {
		return nil, err
	}

	if validSeconds <= 0 {
		validSeconds = s.config.OSSConfig.DefaultValidSeconds
	}

	ossAdapter, err := s.storageService.GetAdapter(ctx, storageID, useInternal)
	if err != nil {
		return nil, err
	}

	return ossAdapter.GetDeleteURL(ctx, objectKey, validSeconds)
}

// CleanExpiredTasks 清理创建时间超过7天且未完成的上传任务
func (s *urlService) CleanExpiredTasks(ctx context.Context) error {
	// 计算7天前的时间
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	return s.taskRepo.DeleteExpired(ctx, sevenDaysAgo)
}
