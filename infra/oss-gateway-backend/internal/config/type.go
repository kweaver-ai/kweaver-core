package config

type AppConfig struct {
	CommonConfig   CommonConfig
	DatabaseConfig DatabaseConfig
	RedisConfig    RedisConfig
	LogConfig      LogConfig
	OSSConfig      OSSConfig
	CryptoConfig   CryptoConfig
}

type CommonConfig struct {
	Port string `envconfig:"PORT" default:"8080"`
	Name string `envconfig:"NAME" default:"oss-gateway"`
}

type DatabaseConfig struct {
	// 本地开发默认值
	Host     string `envconfig:"RDSHOST"     default:"localhost"`
	Port     string `envconfig:"RDSPORT"     default:"3306"`
	User     string `envconfig:"RDSUSER"     default:"root"`
	Password string `envconfig:"RDSPASS"     default:""`
	DBName   string `envconfig:"RDSDBNAME"   default:"kweaver"`
	TYPE     string `envconfig:"DB_TYPE"     default:"MYSQL"`
	SystemID string `envconfig:"DB_SYSTEMID" default:""`

	MaxIdleConns    int `envconfig:"DB_MAX_IDLE_CONNS"     default:"10"`
	MaxOpenConns    int `envconfig:"DB_MAX_OPEN_CONNS"     default:"100"`
	ConnMaxLifetime int `envconfig:"DB_CONN_MAX_LIFETIME"  default:"60"`
	ConnMaxIdleTime int `envconfig:"DB_CONN_MAX_IDLE_TIME" default:"30"`
}

type RedisConfig struct {
	// Redis cluster mode: standalone, master-slave, sentinel
	// 统一使用与 Python 项目一致的环境变量命名
	ClusterMode string `envconfig:"REDISCLUSTERMODE" default:"sentinel"`

	// Standalone mode (单机模式) - 本地开发默认值
	Host     string `envconfig:"REDISHOST" default:"localhost"`
	Port     string `envconfig:"REDISPORT" default:"6379"`
	User     string `envconfig:"REDISUSER" default:""`
	Password string `envconfig:"REDISPASS" default:""`
	DB       int    `envconfig:"REDIS_DB"  default:"2"`
	PoolSize int    `envconfig:"REDIS_POOL_SIZE" default:"100"`

	// Master-Slave mode (主从模式 - 读)
	ReadHost     string `envconfig:"REDISREADHOST" default:"localhost"`
	ReadPort     string `envconfig:"REDISREADPORT" default:"6379"`
	ReadUser     string `envconfig:"REDISREADUSER" default:""`
	ReadPassword string `envconfig:"REDISREADPASS" default:""`

	// Master-Slave mode (主从模式 - 写)
	WriteHost     string `envconfig:"REDISWRITEHOST" default:"localhost"`
	WritePort     string `envconfig:"REDISWRITEPORT" default:"6379"`
	WriteUser     string `envconfig:"REDISWRITEUSER" default:""`
	WritePassword string `envconfig:"REDISWRITEPASS" default:""`

	// Sentinel mode (哨兵模式)
	SentinelAddrs    []string `envconfig:"REDIS_SENTINEL_ADDRS" default:"192.168.40.104:26379"`
	SentinelMaster   string   `envconfig:"SENTINELMASTER"       default:"mymaster"`
	SentinelUser     string   `envconfig:"SENTINELUSER"         default:"root"`
	SentinelPassword string   `envconfig:"SENTINELPASS"         default:"dnPeNWubr0"`
}

type LogConfig struct {
	Level     string `envconfig:"LOG_LEVEL"       default:"info"`
	Format    string `envconfig:"LOG_FORMAT"      default:"json"`
	EnableSQL bool   `envconfig:"LOG_ENABLE_SQL"  default:"false"`
}

type OSSConfig struct {
	DefaultValidSeconds int64 `envconfig:"OSS_DEFAULT_VALID_SECONDS" default:"3600"`
	MaxPartSize         int64 `envconfig:"OSS_MAX_PART_SIZE"         default:"5368709120"`
	MinPartSize         int64 `envconfig:"OSS_MIN_PART_SIZE"         default:"5242880"`
	MaxParts            int   `envconfig:"OSS_MAX_PARTS"             default:"10000"`
}

type CryptoConfig struct {
	AESKey string `envconfig:"CRYPTO_AES_KEY" default:"k8WVs8pfQae0LhUgevDvPXiYPqYZ8HRM"`
}
