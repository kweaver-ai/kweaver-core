package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

func NewConfig() *AppConfig {
	var cfg AppConfig

	if err := envconfig.Process("", &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Parse Sentinel addresses if in sentinel mode
	if cfg.RedisConfig.ClusterMode == "sentinel" {
		if addrs := os.Getenv("REDIS_SENTINEL_ADDRS"); addrs != "" {
			cfg.RedisConfig.SentinelAddrs = strings.Split(addrs, ",")
		}
	}

	return &cfg
}

func (d *DatabaseConfig) DSN() string {
	switch d.TYPE {
	case "DM8":
		return fmt.Sprintf("dm://%s:%s@%s:%s?schema=%s", d.User, d.Password, d.Host, d.Port, d.DBName)
	case "KDB9":
		return fmt.Sprintf("kingbase://%s:%s@%s:%s/%s?sslmode=disable", d.User, d.Password, d.Host, d.Port, d.DBName)
	default:
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", d.User, d.Password, d.Host, d.Port, d.DBName)
	}
}
