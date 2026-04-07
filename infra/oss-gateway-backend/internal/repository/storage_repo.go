package repository

import (
	"context"
	"oss-gateway/internal/model"

	"gorm.io/gorm"
)

type StorageRepository interface {
	Create(ctx context.Context, storage *model.StorageConfig) error
	Update(ctx context.Context, storage *model.StorageConfig) error
	Delete(ctx context.Context, storageID string) error
	GetByID(ctx context.Context, storageID string) (*model.StorageConfig, error)
	GetByIDInt64(ctx context.Context, id int64) (*model.StorageConfig, error)
	List(ctx context.Context, vendorType string, enabled *bool) ([]*model.StorageConfig, error)
	ListWithPagination(ctx context.Context, vendorType string, enabled *bool, isDefault *bool, name string, page int, size int, order string, rule string) ([]*model.StorageConfig, int, error)
	GetDefault(ctx context.Context) (*model.StorageConfig, error)
	HasDefaultStorage(ctx context.Context, excludeStorageID string) (*model.StorageConfig, error)
	// 唯一性校验方法
	ExistsByStorageName(ctx context.Context, storageName string) (bool, error)
	ExistsByBucketAndEndpoint(ctx context.Context, bucketName string, endpoint string) (bool, error)
	ExistsByBucketAndSiteID(ctx context.Context, bucketName string, siteID string) (bool, error)
}

type storageRepository struct {
	db *gorm.DB
}

func NewStorageRepository(db *gorm.DB) StorageRepository {
	return &storageRepository{db: db}
}

func (r *storageRepository) Create(ctx context.Context, storage *model.StorageConfig) error {
	return r.db.WithContext(ctx).Create(storage).Error
}

func (r *storageRepository) Update(ctx context.Context, storage *model.StorageConfig) error {
	return r.db.WithContext(ctx).Save(storage).Error
}

func (r *storageRepository) Delete(ctx context.Context, storageID string) error {
	return r.db.WithContext(ctx).Where("f_storage_id = ?", storageID).Delete(&model.StorageConfig{}).Error
}

func (r *storageRepository) GetByID(ctx context.Context, storageID string) (*model.StorageConfig, error) {
	var storage model.StorageConfig
	err := r.db.WithContext(ctx).Where("f_storage_id = ?", storageID).First(&storage).Error
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (r *storageRepository) GetByIDInt64(ctx context.Context, id int64) (*model.StorageConfig, error) {
	var storage model.StorageConfig
	err := r.db.WithContext(ctx).Where("f_id = ?", id).First(&storage).Error
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (r *storageRepository) List(ctx context.Context, vendorType string, enabled *bool) ([]*model.StorageConfig, error) {
	var storages []*model.StorageConfig
	query := r.db.WithContext(ctx)

	if vendorType != "" {
		query = query.Where("f_vendor_type = ?", vendorType)
	}

	if enabled != nil {
		query = query.Where("f_is_enabled = ?", *enabled)
	}

	err := query.Find(&storages).Error
	return storages, err
}

func (r *storageRepository) GetDefault(ctx context.Context) (*model.StorageConfig, error) {
	var storage model.StorageConfig
	err := r.db.WithContext(ctx).Where("f_is_default = ? AND f_is_enabled = ?", true, true).First(&storage).Error
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

func (r *storageRepository) HasDefaultStorage(ctx context.Context, excludeStorageID string) (*model.StorageConfig, error) {
	var storage model.StorageConfig
	query := r.db.WithContext(ctx).Where("f_is_default = ?", true)
	if excludeStorageID != "" {
		query = query.Where("f_storage_id != ?", excludeStorageID)
	}
	err := query.First(&storage).Error
	if err != nil {
		return nil, err
	}
	return &storage, nil
}

// ListWithPagination 分页查询存储列表，支持模糊搜索和排序
func (r *storageRepository) ListWithPagination(ctx context.Context, vendorType string, enabled *bool, isDefault *bool, name string, page int, size int, order string, rule string) ([]*model.StorageConfig, int, error) {
	var storages []*model.StorageConfig
	var total int64

	query := r.db.WithContext(ctx).Model(&model.StorageConfig{})

	// 过滤条件
	if vendorType != "" {
		query = query.Where("f_vendor_type = ?", vendorType)
	}

	if enabled != nil {
		query = query.Where("f_is_enabled = ?", *enabled)
	}

	if isDefault != nil {
		query = query.Where("f_is_default = ?", *isDefault)
	}

	// 模糊搜索存储名称
	if name != "" {
		query = query.Where("f_storage_name LIKE ?", "%"+name+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序字段映射
	orderField := "f_updated_at"
	switch rule {
	case "create_time":
		orderField = "f_created_at"
	case "update_time":
		orderField = "f_updated_at"
	case "storage_name":
		orderField = "f_storage_name"
	}

	// 排序方向
	orderDirection := "DESC"
	if order == "asc" {
		orderDirection = "ASC"
	}

	// 分页查询
	offset := (page - 1) * size
	err := query.Order(orderField + " " + orderDirection).
		Limit(size).
		Offset(offset).
		Find(&storages).Error

	return storages, int(total), err
}

// ExistsByStorageName 检查存储名称是否已存在（基于数据库）
func (r *storageRepository) ExistsByStorageName(ctx context.Context, storageName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.StorageConfig{}).
		Where("f_storage_name = ?", storageName).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByBucketAndEndpoint 检查 bucket + endpoint 是否已存在（基于数据库）
func (r *storageRepository) ExistsByBucketAndEndpoint(ctx context.Context, bucketName string, endpoint string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.StorageConfig{}).
		Where("f_bucket_name = ? AND f_endpoint = ?", bucketName, endpoint).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByBucketAndSiteID 检查 bucket + siteId 是否已存在（基于数据库）
// 注意：需要在 model.StorageConfig 中添加 SiteID 字段
func (r *storageRepository) ExistsByBucketAndSiteID(ctx context.Context, bucketName string, siteID string) (bool, error) {
	if siteID == "" {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&model.StorageConfig{}).
		Where("f_bucket_name = ? AND f_site_id = ?", bucketName, siteID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
