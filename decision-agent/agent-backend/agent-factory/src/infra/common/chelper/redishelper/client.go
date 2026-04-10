package redishelper

import (
	"context"
	"fmt"
	"sync"
	"time"

	redis "github.com/go-redis/redis/v8"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

const (
	MasterSlaveType string = "master-slave"
	StandaloneType  string = "standalone"
	SentinelType    string = "sentinel"
	ClusterType     string = "cluster"
)

var (
	redisOnce   sync.Once
	redisClient redis.UniversalClient
)

// ConnectRedis return a redis client. If not connected,
// it will automatically reconnect until connected.
func ConnectRedis(conf *cconf.RedisConf) redis.UniversalClient {
	redisOnce.Do(func() {
		ctx := context.Background()

		switch conf.ConnectType {
		case MasterSlaveType:
			for {
				redisClient = masterSlave(conf)
				if err := redisClient.Ping(ctx).Err(); err != nil {
					time.Sleep(time.Duration(3) * time.Second)
				} else {
					break
				}
			}
		case StandaloneType:
			for {
				redisClient = standalone(conf)
				if err := redisClient.Ping(ctx).Err(); err != nil {
					time.Sleep(time.Duration(3) * time.Second)
				} else {
					break
				}
			}
		case SentinelType:
			for {
				redisClient = sentinel(conf)
				if err := redisClient.Ping(ctx).Err(); err != nil {
					time.Sleep(time.Duration(3) * time.Second)
				} else {
					break
				}
			}
		case ClusterType:
			for {
				redisClient = cluster(conf)
				if err := redisClient.Ping(ctx).Err(); err != nil {
					time.Sleep(time.Duration(3) * time.Second)
				} else {
					break
				}
			}
		}

		logger.GetLogger().Infof("connect redis:[%s] success...", conf.ConnectType)
	})

	return redisClient
}

// masterSlave 主从模式
func masterSlave(conf *cconf.RedisConf) redis.UniversalClient {
	if conf.MasterHost == "" {
		conf.MasterHost = "proton-redis-proton-redis.resource.svc.cluster.local"
	}

	if conf.MasterPort == "" {
		conf.MasterPort = "6379"
	}

	opt := &redis.Options{
		Addr:               conf.MasterHost + ":" + conf.MasterPort,
		Password:           conf.Password,
		DB:                 conf.DB,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}

	return redis.NewClient(opt)
}

// standalone 标准模式客户端
func standalone(conf *cconf.RedisConf) redis.UniversalClient {
	if conf.Host == "" {
		conf.Host = "proton-redis-proton-redis.resource.svc.cluster.local"
	}

	if conf.Port == "" {
		conf.Port = "6379"
	}

	opt := &redis.Options{
		Addr:               conf.Host + ":" + conf.Port,
		Password:           conf.Password,
		DB:                 conf.DB,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}

	return redis.NewClient(opt)
}

// sentinel 哨兵模式客户端
func sentinel(conf *cconf.RedisConf) redis.UniversalClient {
	if conf.MasterGroupName == "" {
		conf.MasterGroupName = "mymaster"
	}

	if conf.SentinelPwd == "" {
		conf.SentinelPwd = "eisoo.com123"
	}

	if conf.SentinelHost == "" {
		conf.SentinelHost = "proton-redis-proton-redis-sentinel.resource.svc.cluster.local"
	}

	if conf.SentinelPort == "" {
		conf.SentinelPort = "26379"
	}

	opt := redis.FailoverOptions{
		MasterName:         conf.MasterGroupName,
		SentinelAddrs:      []string{fmt.Sprintf("%v:%v", conf.SentinelHost, conf.SentinelPort)},
		SentinelPassword:   conf.SentinelPwd,
		Username:           conf.UserName,
		Password:           conf.Password,
		DB:                 conf.DB,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}

	return redis.NewFailoverClient(&opt)
}

// cluster 集群模式客户端
func cluster(conf *cconf.RedisConf) redis.UniversalClient {
	if conf.ClusterPwd == "" {
		conf.ClusterPwd = "eisoo.com123"
	}

	opt := redis.ClusterOptions{
		Addrs:              conf.ClusterHosts,
		Password:           conf.ClusterPwd,
		MaxRetries:         conf.MaxRetries,
		PoolSize:           conf.PoolSize,
		ReadTimeout:        time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:       time.Duration(conf.WriteTimeout) * time.Second,
		IdleTimeout:        time.Duration(conf.IdleTimeout) * time.Second,
		IdleCheckFrequency: time.Duration(conf.IdleCheckFrequency) * time.Second,
		MaxConnAge:         time.Duration(conf.MaxConnAge) * time.Second,
		PoolTimeout:        time.Duration(conf.PoolTimeout) * time.Second,
	}

	return redis.NewClusterClient(&opt)
}

func RedisClient() redis.UniversalClient {
	if redisClient == nil {
		panic("[redishelper][RedisClient]: redis client not connected...")
	}

	return redisClient
}
