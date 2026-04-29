package service

import (
	"context"
	"fmt"
	"oss-gateway/internal/cache"
	"oss-gateway/internal/config"
	"oss-gateway/internal/model"
	"oss-gateway/internal/repository"
	"oss-gateway/pkg/adapter"
	"oss-gateway/pkg/crypto"
	"oss-gateway/pkg/errors"
	"oss-gateway/pkg/utils"
	"strings"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// StorageValidationError 存储校验错误
type StorageValidationError struct {
	Code        string
	Message     string
	Description string
}

func (e *StorageValidationError) Error() string {
	return e.Description
}

type StorageService interface {
	Create(ctx context.Context, req *CreateStorageRequest) (string, error)
	Update(ctx context.Context, storageID string, req *UpdateStorageRequest) error
	Delete(ctx context.Context, storageID string) error
	Get(ctx context.Context, storageID string) (*StorageResponse, error)
	List(ctx context.Context, req *ListStorageRequest) (*ListStorageResponse, error)
	CheckConnection(ctx context.Context, storageID string) error
	GetAdapter(ctx context.Context, storageID string, useInternal bool) (adapter.OSSAdapter, error)
}

// ListStorageRequest 列表查询请求，参考 Python FastAPI 的分页规范
type ListStorageRequest struct {
	Page       int    `form:"page"`        // 页码，从1开始，默认1
	Size       int    `form:"size"`        // 每页大小，默认10，最大1000
	Order      string `form:"order"`       // 排序方向: asc/desc，默认 desc
	Rule       string `form:"rule"`        // 排序字段: create_time/update_time/storage_name，默认 update_time
	Name       string `form:"name"`        // 模糊搜索存储名称
	VendorType string `form:"vendor_type"` // 供应商类型过滤
	Enabled    *bool  `form:"enabled"`     // 启用状态过滤
	IsDefault  *bool  `form:"is_default"`  // 默认存储过滤
}

// ListStorageResponse 列表响应
type ListStorageResponse struct {
	Count int                `json:"count"` // 总记录数
	Data  []*StorageResponse `json:"data"`  // 数据列表
}

type storageService struct {
	repo         repository.StorageRepository
	crypto       *crypto.AESCrypto
	storageCache *cache.StorageCache
	config       *config.AppConfig
	log          *logrus.Entry
}

type CreateStorageRequest struct {
	StorageName      string `json:"storage_name" binding:"required"`
	VendorType       string `json:"vendor_type" binding:"required"`
	Endpoint         string `json:"endpoint" binding:"required"`
	BucketName       string `json:"bucket_name" binding:"required"`
	AccessKeyID      string `json:"access_key_id" binding:"required"`
	AccessKeySecret  string `json:"access_key_secret" binding:"required"`
	Region           string `json:"region"` // OSS、OBS、TOS必填，ECEPH非必填
	IsDefault        bool   `json:"is_default"`
	InternalEndpoint string `json:"internal_endpoint"`
	SiteID           string `json:"site_id"` // 站点ID，用于校验 bucket_name + siteId 唯一性
}

type UpdateStorageRequest struct {
	StorageName      string `json:"storage_name"`
	Endpoint         string `json:"endpoint"`
	BucketName       string `json:"bucket_name"`
	AccessKeyID      string `json:"access_key_id"`
	AccessKeySecret  string `json:"access_key_secret"`
	Region           string `json:"region"`
	IsDefault        *bool  `json:"is_default"`
	IsEnabled        *bool  `json:"is_enabled"`
	InternalEndpoint string `json:"internal_endpoint"`
}

type StorageResponse struct {
	StorageID        string `json:"storage_id"`
	StorageName      string `json:"storage_name"`
	VendorType       string `json:"vendor_type"`
	Endpoint         string `json:"endpoint"`
	BucketName       string `json:"bucket_name"`
	Region           string `json:"region"`
	IsDefault        bool   `json:"is_default"`
	IsEnabled        bool   `json:"is_enabled"`
	InternalEndpoint string `json:"internal_endpoint"`
	SiteID           string `json:"site_id"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func NewStorageService(repo repository.StorageRepository, crypto *crypto.AESCrypto, storageCache *cache.StorageCache, config *config.AppConfig, log *logrus.Entry) StorageService {
	return &storageService{
		repo:         repo,
		crypto:       crypto,
		storageCache: storageCache,
		config:       config,
		log:          log,
	}
}

func (s *storageService) Create(ctx context.Context, req *CreateStorageRequest) (string, error) {
	if !s.isValidVendorType(req.VendorType) {
		return "", &StorageValidationError{
			Code:        errors.InvalidVendorType.Code,
			Message:     errors.InvalidVendorType.Message,
			Description: fmt.Sprintf(errors.InvalidVendorType.Description, req.VendorType),
		}
	}

	// Region 校验：OSS、OBS、TOS必填，ECEPH非必填
	if (req.VendorType == "OSS" || req.VendorType == "OBS" || req.VendorType == "TOS") && req.Region == "" {
		return "", fmt.Errorf("region is required for vendor type %s", req.VendorType)
	}

	if !strings.HasPrefix(req.Endpoint, "http://") && !strings.HasPrefix(req.Endpoint, "https://") {
		return "", fmt.Errorf("endpoint must start with http:// or https://")
	}

	// ========== 唯一性校验（基于数据库，Redis 仅作缓存加速） ==========

	// 唯一性校验1: storage_name 唯一性（先查数据库）
	nameExists, err := s.repo.ExistsByStorageName(ctx, req.StorageName)
	if err != nil {
		s.log.WithError(err).Error("failed to check storage name in database")
		return "", fmt.Errorf("failed to check storage name uniqueness")
	}
	if nameExists {
		return "", &StorageValidationError{
			Code:        errors.StorageNameExists.Code,
			Message:     errors.StorageNameExists.Message,
			Description: fmt.Sprintf(errors.StorageNameExists.Description, req.StorageName),
		}
	}

	// 唯一性校验2: bucket_name + endpoint 唯一性（基于数据库）
	bucketEndpointExists, err := s.repo.ExistsByBucketAndEndpoint(ctx, req.BucketName, req.Endpoint)
	if err != nil {
		s.log.WithError(err).Error("failed to check bucket+endpoint in database")
		return "", fmt.Errorf("failed to check bucket and endpoint uniqueness")
	}
	if bucketEndpointExists {
		return "", &StorageValidationError{
			Code:        errors.StorageExists.Code,
			Message:     errors.StorageExists.Message,
			Description: fmt.Sprintf("Bucket(%s) with endpoint(%s) already exists", req.BucketName, req.Endpoint),
		}
	}

	// 唯一性校验3: bucket_name + siteId 唯一性（基于数据库）
	if req.SiteID != "" {
		bucketSiteExists, err := s.repo.ExistsByBucketAndSiteID(ctx, req.BucketName, req.SiteID)
		if err != nil {
			s.log.WithError(err).Error("failed to check bucket+siteId in database")
			return "", fmt.Errorf("failed to check bucket and siteId uniqueness")
		}
		if bucketSiteExists {
			return "", &StorageValidationError{
				Code:        errors.StorageExists.Code,
				Message:     errors.StorageExists.Message,
				Description: fmt.Sprintf("Bucket(%s) with site already exists", req.BucketName),
			}
		}
	}

	encryptedKeyID, err := s.crypto.Encrypt(req.AccessKeyID)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt access_key_id: %w", err)
	}

	encryptedSecret, err := s.crypto.Encrypt(req.AccessKeySecret)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt access_key_secret: %w", err)
	}

	storageID := utils.GenerateStorageID()

	// 如果设置为默认存储，先检查是否已存在其他默认存储
	if req.IsDefault {
		existingDefault, err := s.repo.HasDefaultStorage(ctx, "")
		if err == nil && existingDefault != nil {
			// 已存在默认存储，拒绝创建
			return "", &StorageValidationError{
				Code:        errors.DefaultStorageExists.Code,
				Message:     errors.DefaultStorageExists.Message,
				Description: existingDefault.StorageName, // 只传存储名称
			}
		}
		// 如果 err == gorm.ErrRecordNotFound，说明没有默认存储，可以继续
	}

	storage := &model.StorageConfig{
		StorageID:        storageID,
		StorageName:      req.StorageName,
		VendorType:       req.VendorType,
		Endpoint:         req.Endpoint,
		BucketName:       req.BucketName,
		AccessKeyID:      encryptedKeyID,
		AccessKey:        encryptedSecret,
		Region:           req.Region,
		IsDefault:        req.IsDefault,
		IsEnabled:        true,
		InternalEndpoint: req.InternalEndpoint,
		SiteID:           req.SiteID, // 保存站点ID
	}

	if err := s.repo.Create(ctx, storage); err != nil {
		return "", fmt.Errorf("failed to create storage: %w", err)
	}

	// 创建成功后，设置缓存（仅用于加速查询）
	if err := s.storageCache.SetStorage(ctx, storage); err != nil {
		s.log.WithError(err).Warn("failed to cache storage config")
	}

	return storageID, nil
}

func (s *storageService) Update(ctx context.Context, storageID string, req *UpdateStorageRequest) error {
	storage, err := s.repo.GetByID(ctx, storageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("storage not found")
		}
		return err
	}

	if req.StorageName != "" {
		// 检查存储名称是否与其他存储重名
		if req.StorageName != storage.StorageName {
			nameExists, err := s.repo.ExistsByStorageName(ctx, req.StorageName)
			if err != nil {
				s.log.WithError(err).Error("failed to check storage name in database")
				return fmt.Errorf("failed to check storage name uniqueness")
			}
			if nameExists {
				return &StorageValidationError{
					Code:        errors.StorageNameExists.Code,
					Message:     errors.StorageNameExists.Message,
					Description: fmt.Sprintf(errors.StorageNameExists.Description, req.StorageName),
				}
			}
		}
		storage.StorageName = req.StorageName
	}
	if req.Endpoint != "" {
		if !strings.HasPrefix(req.Endpoint, "http://") && !strings.HasPrefix(req.Endpoint, "https://") {
			return fmt.Errorf("endpoint must start with http:// or https://")
		}
		storage.Endpoint = req.Endpoint
	}
	if req.BucketName != "" {
		storage.BucketName = req.BucketName
	}
	if req.AccessKeyID != "" {
		encrypted, err := s.crypto.Encrypt(req.AccessKeyID)
		if err != nil {
			return fmt.Errorf("failed to encrypt access_key_id: %w", err)
		}
		storage.AccessKeyID = encrypted
	}
	if req.AccessKeySecret != "" {
		encrypted, err := s.crypto.Encrypt(req.AccessKeySecret)
		if err != nil {
			return fmt.Errorf("failed to encrypt access_key_secret: %w", err)
		}
		storage.AccessKey = encrypted
	}
	if req.Region != "" {
		storage.Region = req.Region
	}
	if req.IsDefault != nil {
		// 如果要设置为默认存储，先检查是否已存在其他默认存储
		if *req.IsDefault {
			existingDefault, err := s.repo.HasDefaultStorage(ctx, storageID)
			if err == nil && existingDefault != nil {
				// 已存在其他默认存储，拒绝更新
				return &StorageValidationError{
					Code:        errors.DefaultStorageExists.Code,
					Message:     errors.DefaultStorageExists.Message,
					Description: existingDefault.StorageName, // 只传存储名称
				}
			}
			// 如果 err == gorm.ErrRecordNotFound，说明没有其他默认存储，可以继续
		}
		storage.IsDefault = *req.IsDefault
	}
	if req.IsEnabled != nil {
		storage.IsEnabled = *req.IsEnabled
	}
	if req.InternalEndpoint != "" {
		storage.InternalEndpoint = req.InternalEndpoint
	}

	if err := s.repo.Update(ctx, storage); err != nil {
		return err
	}

	if err := s.storageCache.InvalidateStorage(ctx, storageID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate storage cache")
	}

	return nil
}

func (s *storageService) Delete(ctx context.Context, storageID string) error {
	if err := s.repo.Delete(ctx, storageID); err != nil {
		return err
	}

	if err := s.storageCache.InvalidateStorage(ctx, storageID); err != nil {
		s.log.WithError(err).Warn("failed to invalidate storage cache")
	}

	return nil
}

func (s *storageService) Get(ctx context.Context, storageID string) (*StorageResponse, error) {
	storage, err := s.repo.GetByID(ctx, storageID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("storage not found")
		}
		return nil, err
	}

	return s.toResponse(storage), nil
}

func (s *storageService) List(ctx context.Context, req *ListStorageRequest) (*ListStorageResponse, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	if req.Size > 1000 {
		req.Size = 1000
	}
	if req.Order == "" {
		req.Order = "desc"
	}
	if req.Rule == "" {
		req.Rule = "update_time"
	}

	// 验证排序参数
	validOrders := map[string]bool{"asc": true, "desc": true}
	if !validOrders[req.Order] {
		return nil, fmt.Errorf("invalid order parameter: %s (must be asc or desc)", req.Order)
	}

	validRules := map[string]bool{"create_time": true, "update_time": true, "storage_name": true}
	if !validRules[req.Rule] {
		return nil, fmt.Errorf("invalid rule parameter: %s (must be create_time, update_time or storage_name)", req.Rule)
	}

	// 查询数据库
	storages, total, err := s.repo.ListWithPagination(ctx, req.VendorType, req.Enabled, req.IsDefault, req.Name, req.Page, req.Size, req.Order, req.Rule)
	if err != nil {
		return nil, err
	}

	responses := make([]*StorageResponse, 0, len(storages))
	for _, storage := range storages {
		responses = append(responses, s.toResponse(storage))
	}

	return &ListStorageResponse{
		Count: total,
		Data:  responses,
	}, nil
}

func (s *storageService) CheckConnection(ctx context.Context, storageID string) error {
	ossAdapter, err := s.GetAdapter(ctx, storageID, false)
	if err != nil {
		return err
	}

	return ossAdapter.CheckConnection(ctx)
}

func (s *storageService) GetAdapter(ctx context.Context, storageID string, useInternal bool) (adapter.OSSAdapter, error) {
	// Try to get from cache first
	storage, err := s.storageCache.GetStorage(ctx, storageID)
	if err != nil {
		s.log.WithError(err).Warn("failed to get storage from cache")
	}

	// If not in cache, get from data store
	if storage == nil {
		storage, err = s.repo.GetByID(ctx, storageID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("storage not found")
			}
			return nil, err
		}

		// Cache the storage config
		if err := s.storageCache.SetStorage(ctx, storage); err != nil {
			s.log.WithError(err).Warn("failed to cache storage config")
		}
	}

	if !storage.IsEnabled {
		return nil, fmt.Errorf("storage is disabled")
	}

	// 添加详细日志
	s.log.Infof("Decrypting storage %s", storageID)
	s.log.Infof("AccessKeyID length from DB: %d", len(storage.AccessKeyID))
	s.log.Infof("AccessKey length from DB: %d", len(storage.AccessKey))

	// 尝试解密，如果解密失败则认为是明文
	accessKeyID := storage.AccessKeyID
	if decrypted, err := s.crypto.Decrypt(storage.AccessKeyID); err == nil {
		accessKeyID = decrypted
		s.log.Infof("✅ AccessKeyID decrypted successfully, result length: %d", len(accessKeyID))
	} else {
		s.log.WithError(err).Warnf("⚠️ Failed to decrypt access_key_id, using as plaintext")
	}

	accessKeySecret := storage.AccessKey
	if decrypted, err := s.crypto.Decrypt(storage.AccessKey); err == nil {
		accessKeySecret = decrypted
		s.log.Infof("✅ AccessKey decrypted successfully, result length: %d", len(accessKeySecret))
	} else {
		s.log.WithError(err).Warnf("⚠️ Failed to decrypt access_key_secret, using as plaintext")
	}

	endpoint := storage.Endpoint
	if useInternal && storage.InternalEndpoint != "" {
		endpoint = storage.InternalEndpoint
	}

	endpointClean, useSSL := utils.ParseEndpoint(endpoint)

	adapterConfig := adapter.StorageConfig{
		StorageID:       storage.StorageID,
		VendorType:      adapter.VendorType(storage.VendorType),
		Endpoint:        endpointClean,
		BucketName:      storage.BucketName,
		AccessKeyID:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Region:          storage.Region,
		UseSSL:          useSSL,
	}

	// Note: Adapter instances are created on-demand, not cached
	// This ensures fresh connections and avoids connection pooling issues
	return adapter.NewAdapter(adapterConfig)
}

func (s *storageService) toResponse(storage *model.StorageConfig) *StorageResponse {
	return &StorageResponse{
		StorageID:        storage.StorageID,
		StorageName:      storage.StorageName,
		VendorType:       storage.VendorType,
		Endpoint:         storage.Endpoint,
		BucketName:       storage.BucketName,
		Region:           storage.Region,
		IsDefault:        storage.IsDefault,
		IsEnabled:        storage.IsEnabled,
		InternalEndpoint: storage.InternalEndpoint,
		SiteID:           storage.SiteID,
		CreatedAt:        storage.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        storage.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (s *storageService) isValidVendorType(vendorType string) bool {
	switch vendorType {
	case "OSS", "OBS", "ECEPH", "TOS":
		return true
	default:
		return false
	}
}
