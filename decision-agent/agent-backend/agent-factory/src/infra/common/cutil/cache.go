package cutil

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// 设置缓存
func SetCache[T any](ctx context.Context, cache redis.Cmdable, key string, value T, expire time.Duration) error {
	// 序列化对象
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 存储到Redis
	err = cache.Set(ctx, key, data, expire).Err()
	if err != nil {
		return err
	}

	return nil
}

// 获取缓存
func GetCache[T any](ctx context.Context, cache redis.Cmdable, key string) (*T, error) {
	// 从Redis获取数据
	data, err := cache.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 键不存在
		}

		return nil, err
	}

	// 反序列化对象
	var obj T
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

// 删除缓存
func DelCache(ctx context.Context, cache redis.Cmdable, key string) error {
	err := cache.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}
