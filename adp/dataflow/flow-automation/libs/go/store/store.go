package store

import (
	"time"

	jsoniter "github.com/json-iterator/go"
	gocache "github.com/patrickmn/go-cache"
)

var localCacheSingleton *cache

var (
	localCache     LocalCache
	defaultTTL     = 10 * time.Second
	defaultCleanUp = 30 * time.Second
)

type cache struct {
	localCache *gocache.Cache
}

type Option struct {
	Expiration      time.Duration
	CleanUpInterval time.Duration
}

// InitSingletonLocalLocalCache 实例化
func InitSingletonLocalLocalCache(opt *Option) {
	if opt != nil {
		defaultTTL = opt.Expiration
		defaultCleanUp = opt.CleanUpInterval
	}
	localCacheSingleton = &cache{
		localCache: gocache.New(defaultTTL, defaultCleanUp),
	}
}

// Get 获取本地缓存
func (c *cache) Get(key string) ([]byte, bool) {
	res, exist := c.localCache.Get(key)
	if !exist {
		return nil, exist
	}
	resByte, _ := jsoniter.Marshal(res)
	return resByte, exist
}

// GetRaw 获取本地缓存
func (c *cache) GetRaw(key string) (interface{}, bool) {
	res, exist := c.localCache.Get(key)
	if !exist {
		return nil, exist
	}
	return res, exist
}

// Set 设置本地缓存
func (c *cache) Set(key string, value interface{}, ttl time.Duration) {
	c.localCache.Set(key, value, ttl)
}

// Delete 删除本地缓存
func (c *cache) Delete(key string) {
	c.localCache.Delete(key)
}

// GetDefaultExpiration 获取默认过期时间
func (c *cache) GetDefaultExpiration() time.Duration {
	return gocache.DefaultExpiration
}

// LocalCache 接口
type LocalCache interface {
	// Get 获取本地缓存
	Get(key string) ([]byte, bool)
	// GetRaw 获取本地缓存
	GetRaw(key string) (interface{}, bool)
	// Set 设置本地缓存
	Set(key string, value interface{}, ttl time.Duration)
	// Delete 删除本地缓存
	Delete(key string)
	// GetDefaultExpiration
	GetDefaultExpiration() time.Duration
}

// NewSingletonLocalCache 返回单例 LocalCache
func NewSingletonLocalCache() LocalCache {
	return localCacheSingleton
}

// NewLocalCache 实例化本地缓存
func NewLocalCache(opt *Option) LocalCache {
	var customTTL, customCleanUp = defaultTTL, defaultCleanUp
	if opt != nil {
		customTTL = opt.Expiration
		customCleanUp = opt.CleanUpInterval
	}
	return &cache{
		localCache: gocache.New(customTTL, customCleanUp),
	}
}
