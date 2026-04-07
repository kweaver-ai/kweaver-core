package cconf

import (
	"testing"
)

func TestRedisConf_Fields(t *testing.T) {
	t.Run("redis config fields are accessible", func(t *testing.T) {
		redisconf := RedisConf{
			ConnectType: "standalone",
			UserName:    "redisuser",
			Password:    "redispass",
			Host:        "localhost",
			Port:        "6379",
			DB:          0,
			MaxRetries:  3,
			PoolSize:    10,
		}

		if redisconf.ConnectType != "standalone" {
			t.Errorf("Expected ConnectType to be 'standalone', got '%s'", redisconf.ConnectType)
		}

		if redisconf.DB != 0 {
			t.Errorf("Expected DB to be 0, got %d", redisconf.DB)
		}

		if redisconf.MaxRetries != 3 {
			t.Errorf("Expected MaxRetries to be 3, got %d", redisconf.MaxRetries)
		}
	})
}

func TestRedisConf_SentinelMode(t *testing.T) {
	t.Run("redis sentinel mode config", func(t *testing.T) {
		redisconf := RedisConf{
			ConnectType:     "sentinel",
			MasterGroupName: "mymaster",
			SentinelHost:    "sentinel-host",
			SentinelPort:    "26379",
		}

		if redisconf.ConnectType != "sentinel" {
			t.Errorf("Expected ConnectType to be 'sentinel', got '%s'", redisconf.ConnectType)
		}

		if redisconf.MasterGroupName != "mymaster" {
			t.Errorf("Expected MasterGroupName to be 'mymaster', got '%s'", redisconf.MasterGroupName)
		}
	})
}

func TestRedisConf_MasterSlaveMode(t *testing.T) {
	t.Run("redis master-slave mode config", func(t *testing.T) {
		redisconf := RedisConf{
			ConnectType: "master-slave",
			MasterHost:  "master-host",
			MasterPort:  "6379",
			SlaveHost:   "slave-host",
			SlavePort:   "6379",
		}

		if redisconf.ConnectType != "master-slave" {
			t.Errorf("Expected ConnectType to be 'master-slave', got '%s'", redisconf.ConnectType)
		}

		if redisconf.MasterHost != "master-host" {
			t.Errorf("Expected MasterHost to be 'master-host', got '%s'", redisconf.MasterHost)
		}
	})
}

func TestRedisConf_TimeoutSettings(t *testing.T) {
	t.Run("redis config timeout settings", func(t *testing.T) {
		redisconf := RedisConf{
			ReadTimeout:        3,
			WriteTimeout:       3,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
			MaxConnAge:         300,
			PoolTimeout:        8,
		}

		if redisconf.ReadTimeout != 3 {
			t.Errorf("Expected ReadTimeout to be 3, got %d", redisconf.ReadTimeout)
		}

		if redisconf.WriteTimeout != 3 {
			t.Errorf("Expected WriteTimeout to be 3, got %d", redisconf.WriteTimeout)
		}

		if redisconf.IdleTimeout != 300 {
			t.Errorf("Expected IdleTimeout to be 300, got %d", redisconf.IdleTimeout)
		}
	})
}
