package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"oss-gateway/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	storageConfigPrefix     = "kweaver-core:oss-gateway-backend:storage:config:"
	storageNamePrefix       = "kweaver-core:oss-gateway-backend:storage:name:"        // storage_name 唯一性索引
	storageBucketHostPrefix = "kweaver-core:oss-gateway-backend:storage:bucket:host:" // bucket_name + host 唯一性索引
	storageBucketSitePrefix = "kweaver-core:oss-gateway-backend:storage:bucket:site:" // bucket_name + siteId 唯一性索引
	storageConfigTTL        = 1 * time.Hour
)

type StorageCache struct {
	redis *RedisClient
}

func NewStorageCache(redis *RedisClient) *StorageCache {
	return &StorageCache{redis: redis}
}

func (c *StorageCache) GetStorage(ctx context.Context, storageID string) (*model.StorageConfig, error) {
	key := storageConfigPrefix + storageID

	data, err := c.redis.Get(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get storage from cache: %w", err)
	}

	var storage model.StorageConfig
	if err := json.Unmarshal([]byte(data), &storage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage config: %w", err)
	}

	return &storage, nil
}

func (c *StorageCache) SetStorage(ctx context.Context, storage *model.StorageConfig) error {
	key := storageConfigPrefix + storage.StorageID

	data, err := json.Marshal(storage)
	if err != nil {
		return fmt.Errorf("failed to marshal storage: %w", err)
	}

	return c.redis.Set(ctx, key, data, storageConfigTTL)
}

func (c *StorageCache) DeleteStorage(ctx context.Context, storageID string) error {
	key := storageConfigPrefix + storageID
	return c.redis.Del(ctx, key)
}

func (c *StorageCache) InvalidateStorage(ctx context.Context, storageID string) error {
	return c.DeleteStorage(ctx, storageID)
}

// CheckStorageNameExists 检查 storage_name 是否已存在
func (c *StorageCache) CheckStorageNameExists(ctx context.Context, storageName string) (bool, error) {
	key := storageNamePrefix + storageName
	count, err := c.redis.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check storage name: %w", err)
	}
	return count > 0, nil
}

// SetStorageName 设置 storage_name 索引
func (c *StorageCache) SetStorageName(ctx context.Context, storageName string, storageID string) error {
	key := storageNamePrefix + storageName
	return c.redis.Set(ctx, key, storageID, storageConfigTTL)
}

// DeleteStorageName 删除 storage_name 索引
func (c *StorageCache) DeleteStorageName(ctx context.Context, storageName string) error {
	key := storageNamePrefix + storageName
	return c.redis.Del(ctx, key)
}

// CheckBucketHostExists 检查 bucket_name + host 是否已存在
func (c *StorageCache) CheckBucketHostExists(ctx context.Context, bucketName string, host string) (bool, error) {
	key := storageBucketHostPrefix + bucketName + ":" + host
	count, err := c.redis.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket and host: %w", err)
	}
	return count > 0, nil
}

// SetBucketHost 设置 bucket_name + host 索引
func (c *StorageCache) SetBucketHost(ctx context.Context, bucketName string, host string, storageID string) error {
	key := storageBucketHostPrefix + bucketName + ":" + host
	return c.redis.Set(ctx, key, storageID, storageConfigTTL)
}

// DeleteBucketHost 删除 bucket_name + host 索引
func (c *StorageCache) DeleteBucketHost(ctx context.Context, bucketName string, host string) error {
	key := storageBucketHostPrefix + bucketName + ":" + host
	return c.redis.Del(ctx, key)
}

// CheckBucketSiteExists 检查 bucket_name + siteId 是否已存在
func (c *StorageCache) CheckBucketSiteExists(ctx context.Context, bucketName string, siteID string) (bool, error) {
	key := storageBucketSitePrefix + bucketName + ":" + siteID
	count, err := c.redis.Exists(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket and site: %w", err)
	}
	return count > 0, nil
}

// SetBucketSite 设置 bucket_name + siteId 索引
func (c *StorageCache) SetBucketSite(ctx context.Context, bucketName string, siteID string, storageID string) error {
	key := storageBucketSitePrefix + bucketName + ":" + siteID
	return c.redis.Set(ctx, key, storageID, storageConfigTTL)
}

// DeleteBucketSite 删除 bucket_name + siteId 索引
func (c *StorageCache) DeleteBucketSite(ctx context.Context, bucketName string, siteID string) error {
	key := storageBucketSitePrefix + bucketName + ":" + siteID
	return c.redis.Del(ctx, key)
}
