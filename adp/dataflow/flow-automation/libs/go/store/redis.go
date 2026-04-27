package store

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	redis_v8 "github.com/go-redis/redis/v8"
)

var redis *rdb

var ErrLockNotAcquired = errors.New("lock not acquired")

// RDB 缓存接口
type RDB interface {
	// Get 缓存获取key
	Get(key string) (interface{}, error)

	// Set 缓存设置key, val, ttl.
	Set(key string, value interface{}, ttl time.Duration) error

	// TryLock 获取缓存锁
	TryLock(key string, val interface{}, ttl time.Duration) error

	// Unlock cache lock unlock.
	Unlock(key, lockid string) (bool, error)

	// Renew 锁续租
	Renew(key, lockid string, ttl time.Duration) (bool, error)

	// GetExpireTime 获取锁剩余时间
	GetExpireTime(key string) (time.Duration, error)

	// GetClient
	GetClient() redis_v8.UniversalClient
}

// RedisConfiguration redis配置
type RedisConfiguration struct {
	Ctx              context.Context
	Host             string
	Port             int
	SlaveHost        string
	SlavePort        int
	UserName         string
	Password         string
	SentinelPassword string
	SentinelUsername string
	MasterGroupName  string
	// 是否是云模式
	ClusterMode string `yaml:"clusterMode,omitempty"`
}

type rdb struct {
	ctx context.Context
	rdb redis_v8.UniversalClient
}

// Get get key
func (r *rdb) Get(key string) (interface{}, error) {
	cmd := r.rdb.Get(r.ctx, key)
	if errors.Is(cmd.Err(), redis_v8.Nil) {
		return nil, nil
	}
	return cmd.Val(), nil
}

// Set set key val ttl.
func (r *rdb) Set(key string, val interface{}, ttl time.Duration) error {
	return r.rdb.Set(r.ctx, key, val, ttl).Err()
}

// TryLock 获取分布式锁
func (r *rdb) TryLock(key string, val interface{}, ttl time.Duration) error {
	lock, err := r.rdb.SetNX(r.ctx, key, val, ttl).Result()
	if err != nil {
		return err
	}

	if !lock {
		return ErrLockNotAcquired
	}

	return nil
}

// Unlock 解锁
func (r *rdb) Unlock(key, lockid string) (bool, error) {
	// atomic unlock
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1]
	then
		return redis.call("del", KEYS[1])
	else
		return 0
	end`

	result, err := redis_v8.NewScript(script).Run(r.ctx, r.rdb, []string{key}, lockid).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

// Renew 续租
func (r *rdb) Renew(key, lockid string, ttl time.Duration) (bool, error) {
	// atomic renew
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1]
	then
		return redis.call("pexpire", KEYS[1], ARGV[2])
	else
		return 0
	end`

	result, err := redis_v8.NewScript(script).Run(r.ctx, r.rdb, []string{key}, lockid, int64(ttl)).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}

// GetExpireTime 获取锁剩余时间
func (r *rdb) GetExpireTime(key string) (time.Duration, error) {
	res, err := r.rdb.PTTL(r.ctx, key).Result()
	return res, err
}

func (r *rdb) GetClient() redis_v8.UniversalClient {
	return r.rdb
}

// DB 获取go-redis client
func InitRedis(redisConfig *RedisConfiguration) {
	var cli redis_v8.UniversalClient
	var addr string

	if redisConfig.ClusterMode == "" {
		panic("connect type empty")
	}

	// 手动添加schema 保证parse不会出错 判断host中是否包含port
	result, err := url.Parse(fmt.Sprintf("http://%s", redisConfig.Host))
	if err != nil || result.Port() == "" {
		addr = fmt.Sprintf("%s:%v", redisConfig.Host, redisConfig.Port)
	} else {
		addr = redisConfig.Host
	}

	// 哨兵模式
	if redisConfig.ClusterMode == "sentinel" {
		cli = redis_v8.NewFailoverClient(&redis_v8.FailoverOptions{
			MasterName:       redisConfig.MasterGroupName,
			SentinelAddrs:    []string{addr},
			SentinelPassword: redisConfig.SentinelPassword,
			Username:         redisConfig.UserName,
			Password:         redisConfig.Password,
		})
	} else if redisConfig.ClusterMode == "master-slave" || redisConfig.ClusterMode == "standalone" {
		cli = redis_v8.NewClient(
			&redis_v8.Options{
				Addr:     addr,
				Username: redisConfig.UserName,
				Password: redisConfig.Password,
			})
	} else {
		hosts := strings.Split(redisConfig.Host, ",")
		addrs := make([]string, 0, len(hosts))
		for _, host := range hosts {
			if strings.Contains(host, ":") {
				addrs = append(addrs, host)
			} else {
				addrs = append(addrs, fmt.Sprintf("%s:%d", host, redisConfig.Port))
			}
		}
		cli = redis_v8.NewClusterClient(
			&redis_v8.ClusterOptions{
				Addrs:    addrs,
				Username: redisConfig.UserName,
				Password: redisConfig.Password,
			},
		)
	}
	s, err := cli.Ping(context.Background()).Result()
	if err != nil || s != "PONG" {
		panic(err)
	}
	redis = &rdb{
		ctx: redisConfig.Ctx,
		rdb: cli,
	}
}

func NewRedis() RDB {
	return redis
}
