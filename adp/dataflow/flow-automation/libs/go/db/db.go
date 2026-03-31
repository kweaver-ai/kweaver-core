package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	dm "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/dialect/dm"
	kdb "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/dialect/kdb"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/utils"
	_ "github.com/kweaver-ai/proton-rds-sdk-go/driver" // 注册数据库驱动
	mysqld "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

const (
	ProntonRdsDriver = "proton-rds"
	DM8Driver        = "DM8"
	KDBDBType        = "KDB"
)

// Config 数据库配置
type Config struct {
	Host            string
	Port            string
	Driver          string
	Name            string
	User            string
	Password        string
	Timezone        string
	MaxIdleConns    string
	MaxOpenConns    string
	ConnMaxLifetime string
}

// ParseDSN 解析dsn
func ParseDSN(config *Config) string {
	location, err := time.LoadLocation(config.Timezone)
	if err != nil {
		panic(err)
	}
	dsnConfig := mysql.NewConfig()
	host := config.Host
	if utils.IsIPv6(host) {
		config.Host = fmt.Sprintf("[%s]", host)
	}
	dsnConfig.Addr = config.Host + ":" + config.Port
	dsnConfig.User = config.User
	dsnConfig.Passwd = config.Password
	dsnConfig.DBName = config.Name
	dsnConfig.Net = "tcp"
	dsnConfig.Loc = location
	dsnConfig.Params = map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "true",
	}
	return dsnConfig.FormatDSN()
}

// InitGormDB return *gorm.DB
func InitGormDB(config *Config) error {
	var gormDB *gorm.DB
	var err error
	var dialector gorm.Dialector
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,         // Disable color
		},
	)
	dsn := ParseDSN(config)
	sqlDB, err := sql.Open(ProntonRdsDriver, dsn)
	if err != nil {
		return err
	}

	driver := config.Driver
	if strings.HasPrefix(driver, KDBDBType) {
		driver = KDBDBType
	}

	switch driver {
	case DM8Driver:
		dialector = dm.New(dm.Config{Conn: sqlDB})
	case KDBDBType:
		dialector = kdb.New(kdb.Config{Conn: sqlDB})
	default:
		dialector = mysqld.New(mysqld.Config{Conn: sqlDB})
	}
	if gormDB, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	}); err != nil {
		return err
	}
	maxIdleConns, err := strconv.Atoi(config.MaxIdleConns)
	if err != nil {
		panic(err)
	}

	maxOpenConns, err := strconv.Atoi(config.MaxOpenConns)
	if err != nil {
		panic(err)
	}

	connMaxLifetime, err := time.ParseDuration(config.ConnMaxLifetime)
	if err != nil {
		panic(err)
	}

	if sqlDB, err := gormDB.DB(); err != nil {
		return err
	} else {
		sqlDB.SetMaxIdleConns(maxIdleConns)
		sqlDB.SetMaxOpenConns(maxOpenConns)
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
	}
	db = gormDB
	return err
}

// InitTables init dabatases tables, param table is a slice which contain the table you want to add. pay attention to use &
func InitTables(gormDB *gorm.DB, tables []interface{}) {
	gormDB.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(tables...)
}

// NewDB 外部获取*gorm.DB实例的方式
func NewDB() *gorm.DB {
	return db
}
