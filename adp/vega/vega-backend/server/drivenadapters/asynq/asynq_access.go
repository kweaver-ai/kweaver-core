// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package drivenadapters provides Asynq task queue implementation.
package drivenadapters

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	aqAccessOnce sync.Once
	aqAccess     interfaces.AsynqAccess
)

type asynqAccess struct {
	appSetting *common.AppSetting
}

// NewAsynqAccess creates or returns the singleton AsynqAccess implementation.
func NewAsynqAccess(appSetting *common.AppSetting) interfaces.AsynqAccess {
	aqAccessOnce.Do(func() {
		aqAccess = &asynqAccess{
			appSetting: appSetting,
		}
	})

	return aqAccess
}

// CreateClient creates and returns the Asynq client for enqueueing tasks.
func (aqa *asynqAccess) CreateClient(ctx context.Context) *asynq.Client {
	redisOpt := aqa.getRedisClientOpt()
	return asynq.NewClient(redisOpt)
}

// CreateServer creates and returns the Asynq server for processing tasks.
func (aqa *asynqAccess) CreateServer(ctx context.Context) *asynq.Server {
	redisOpt := aqa.getRedisClientOpt()
	return asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			interfaces.HighQueue:    6,
			interfaces.DefaultQueue: 3,
			interfaces.LowQueue:     1,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Errorf("Task %s failed: %v", task.Type(), err)
		}),
	})
}

// getRedisClientOpt returns Redis client options based on ConnectType
func (aqa *asynqAccess) getRedisClientOpt() asynq.RedisConnOpt {
	redisSetting := aqa.appSetting.RedisSetting

	switch redisSetting.ConnectType {
	case "sentinel":
		// For sentinel mode, use the sentinel address
		return asynq.RedisFailoverClientOpt{
			SentinelAddrs:    []string{fmt.Sprintf("%s:%d", redisSetting.SentinelHost, redisSetting.SentinelPort)},
			Username:         redisSetting.Username,
			Password:         redisSetting.Password,
			SentinelUsername: redisSetting.SentinelUsername,
			SentinelPassword: redisSetting.SentinelPassword,
			MasterName:       redisSetting.MasterGroupName,
			DB:               0,
			DialTimeout:      5 * time.Second,  // 连接超时（默认 5s，调大）
			ReadTimeout:      60 * time.Second, // 读超时（默认 3s，调大）
			WriteTimeout:     60 * time.Second, // 写超时（默认 3s，调大）
			PoolSize:         20,               // 连接池大小（默认 10，根据并发调大）
		}
	case "master-slave":
		// For master-slave, use standalone mode with master address
		return asynq.RedisClientOpt{
			Addr:         fmt.Sprintf("%s:%d", redisSetting.MasterHost, redisSetting.MasterPort),
			Username:     redisSetting.Username,
			Password:     redisSetting.Password,
			DB:           0,
			DialTimeout:  5 * time.Second,  // 连接超时（默认 5s，调大）
			ReadTimeout:  60 * time.Second, // 读超时（默认 3s，调大）
			WriteTimeout: 60 * time.Second, // 写超时（默认 3s，调大）
			PoolSize:     20,               // 连接池大小（默认 10，根据并发调大）
		}
	case "cluster", "standalone":
		// For cluster and standalone mode, use the same configuration
		return asynq.RedisClientOpt{
			Addr:         fmt.Sprintf("%s:%d", redisSetting.Host, redisSetting.Port),
			Username:     redisSetting.Username,
			Password:     redisSetting.Password,
			DB:           0,
			DialTimeout:  5 * time.Second,  // 连接超时（默认 5s，调大）
			ReadTimeout:  60 * time.Second, // 读超时（默认 3s，调大）
			WriteTimeout: 60 * time.Second, // 写超时（默认 3s，调大）
			PoolSize:     20,               // 连接池大小（默认 10，根据并发调大）
		}
	default:
		// Fallback to standalone mode if ConnectType is unknown
		logger.Warnf("Unknown Redis ConnectType: %s, falling back to standalone mode", redisSetting.ConnectType)
		return asynq.RedisClientOpt{
			Addr:         fmt.Sprintf("%s:%d", redisSetting.Host, redisSetting.Port),
			Username:     redisSetting.Username,
			Password:     redisSetting.Password,
			DB:           0,
			DialTimeout:  5 * time.Second,  // 连接超时（默认 5s，调大）
			ReadTimeout:  60 * time.Second, // 读超时（默认 3s，调大）
			WriteTimeout: 60 * time.Second, // 写超时（默认 3s，调大）
			PoolSize:     20,               // 连接池大小（默认 10，根据并发调大）
		}
	}
}
