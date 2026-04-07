// Package common dbPool
package common

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"

	// _ 注册proton-rds驱动
	_ "github.com/kweaver-ai/proton-rds-sdk-go/driver"
)

var (
	dbOnce sync.Once
	dbPool *sqlx.DB = nil
)

// NewDBPool 获取数据库连接池
func NewDBPool() *sqlx.DB {
	dbOnce.Do(func() {
		dbLog := logger.GetLogger()
		connInfo := sqlx.DBConfig{
			User:             global.GConfig.DB.UserName,
			Password:         global.GConfig.DB.Password,
			Host:             global.GConfig.DB.DBHost,
			Port:             global.GConfig.DB.DBPort,
			HostRead:         global.GConfig.DB.DBHost,
			PortRead:         global.GConfig.DB.DBPort,
			Database:         global.GConfig.DB.DBName,
			Charset:          global.GConfig.DB.Charset,
			Timeout:          global.GConfig.DB.Timeout,
			ReadTimeout:      global.GConfig.DB.TimeoutRead,
			WriteTimeout:     global.GConfig.DB.TimeoutWrite,
			MaxOpenConns:     global.GConfig.DB.MaxOpenConns,
			MaxOpenReadConns: global.GConfig.DB.MaxOpenReadConns,
		}

		var err error

		dbPool, err = sqlx.NewDB(&connInfo)
		if err != nil {
			dbLog.Fatalf("new db operator failed: %v\n", err)
		}
	})

	return dbPool
}
