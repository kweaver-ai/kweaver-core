package cache

import (
	"context"
	"fmt"
	"oss-gateway/internal/config"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type RedisClient struct {
	client      redis.UniversalClient
	clusterMode string
	log         *logrus.Entry
}

// NewRedisClient creates a Redis client based on cluster mode
// Supports three modes: standalone, master-slave, sentinel
func NewRedisClient(cfg *config.AppConfig, log *logrus.Entry) (*RedisClient, error) {
	clusterMode := strings.ToLower(cfg.RedisConfig.ClusterMode)
	log.Infof("Initializing Redis client with mode: %s", clusterMode)

	var client redis.UniversalClient
	var err error

	switch clusterMode {
	case "sentinel":
		client, err = newSentinelClient(cfg, log)
	case "master-slave":
		client, err = newMasterSlaveClient(cfg, log)
	case "standalone":
		client, err = newStandaloneClient(cfg, log)
	default:
		return nil, fmt.Errorf("unsupported redis cluster mode: %s (must be: standalone, master-slave, sentinel)", clusterMode)
	}

	if err != nil {
		return nil, err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Infof("Redis connected successfully in %s mode", clusterMode)

	return &RedisClient{
		client:      client,
		clusterMode: clusterMode,
		log:         log,
	}, nil
}

// newStandaloneClient creates a standalone Redis client
func newStandaloneClient(cfg *config.AppConfig, log *logrus.Entry) (redis.UniversalClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.RedisConfig.Host, cfg.RedisConfig.Port)

	log.Infof("Connecting to Redis standalone: %s", addr)

	opts := &redis.Options{
		Addr:     addr,
		Password: cfg.RedisConfig.Password,
		DB:       cfg.RedisConfig.DB,
		PoolSize: cfg.RedisConfig.PoolSize,
	}

	if cfg.RedisConfig.User != "" {
		opts.Username = cfg.RedisConfig.User
	}

	return redis.NewClient(opts), nil
}

// newMasterSlaveClient creates a master-slave Redis client
// Write operations go to master, read operations go to slave
func newMasterSlaveClient(cfg *config.AppConfig, log *logrus.Entry) (redis.UniversalClient, error) {
	masterAddr := fmt.Sprintf("%s:%s", cfg.RedisConfig.WriteHost, cfg.RedisConfig.WritePort)
	slaveAddr := fmt.Sprintf("%s:%s", cfg.RedisConfig.ReadHost, cfg.RedisConfig.ReadPort)

	log.Infof("Connecting to Redis master-slave: master=%s, slave=%s", masterAddr, slaveAddr)

	// Use FailoverClient for master-slave setup
	opts := &redis.FailoverOptions{
		MasterName:    "master",
		SentinelAddrs: []string{masterAddr}, // In master-slave mode, we connect directly
		Password:      cfg.RedisConfig.WritePassword,
		DB:            cfg.RedisConfig.DB,
		PoolSize:      cfg.RedisConfig.PoolSize,

		// Read operations can go to slave
		RouteByLatency: true,
		RouteRandomly:  true,
	}

	if cfg.RedisConfig.WriteUser != "" {
		opts.Username = cfg.RedisConfig.WriteUser
	}

	// Note: For true master-slave without sentinel, we create a write client
	// and handle read/write separately in the cache layer
	writeOpts := &redis.Options{
		Addr:     masterAddr,
		Password: cfg.RedisConfig.WritePassword,
		DB:       cfg.RedisConfig.DB,
		PoolSize: cfg.RedisConfig.PoolSize,
	}

	if cfg.RedisConfig.WriteUser != "" {
		writeOpts.Username = cfg.RedisConfig.WriteUser
	}

	return redis.NewClient(writeOpts), nil
}

// newSentinelClient creates a Redis Sentinel client
func newSentinelClient(cfg *config.AppConfig, log *logrus.Entry) (redis.UniversalClient, error) {
	if len(cfg.RedisConfig.SentinelAddrs) == 0 {
		return nil, fmt.Errorf("sentinel mode requires at least one sentinel address")
	}

	if cfg.RedisConfig.SentinelMaster == "" {
		return nil, fmt.Errorf("sentinel mode requires master name")
	}

	log.Infof("Connecting to Redis sentinel: addrs=%v, master=%s",
		cfg.RedisConfig.SentinelAddrs, cfg.RedisConfig.SentinelMaster)

	opts := &redis.FailoverOptions{
		MasterName:       cfg.RedisConfig.SentinelMaster,
		SentinelAddrs:    cfg.RedisConfig.SentinelAddrs,
		SentinelPassword: cfg.RedisConfig.SentinelPassword,
		Password:         cfg.RedisConfig.Password,
		DB:               cfg.RedisConfig.DB,
		PoolSize:         cfg.RedisConfig.PoolSize,
	}

	if cfg.RedisConfig.SentinelUser != "" {
		opts.SentinelUsername = cfg.RedisConfig.SentinelUser
	}

	if cfg.RedisConfig.User != "" {
		opts.Username = cfg.RedisConfig.User
	}

	return redis.NewFailoverClient(opts), nil
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.client.Exists(ctx, keys...).Result()
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Client() redis.UniversalClient {
	return r.client
}

func (r *RedisClient) GetClusterMode() string {
	return r.clusterMode
}
