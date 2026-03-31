// Package db 数据库连接池
// @file db.go
// @description 初始化连接池
package db

import (
	"database/sql"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/qustavo/sqlhooks/v2"

	// _ 注册proton-rds驱动
	protonRDS "github.com/kweaver-ai/proton-rds-sdk-go/driver"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

const (
	traceDriverName = "rds-trace"
)

var (
	dbOnce sync.Once
	dbPool *sqlx.DB
)

func initTraceHook() {
	hook := &telemetry.RDSHook{System: "rds"}
	sql.Register(traceDriverName, sqlhooks.Wrap(new(protonRDS.RDSDriver), hook))
}

// NewDBPool 获取数据库连接池
func NewDBPool() *sqlx.DB {
	dbOnce.Do(func() {
		initTraceHook()
		conf := config.NewConfigLoader()
		logger := conf.GetLogger()
		dbName := conf.GetDBName()
		connInfo := sqlx.DBConfig{
			User:         conf.DB.UserName,
			Password:     conf.DB.Password,
			Host:         conf.DB.Host,
			Port:         conf.DB.Port,
			HostRead:     conf.DB.Host,
			PortRead:     conf.DB.Port,
			Database:     dbName,
			Charset:      conf.DB.Charset,
			Timeout:      conf.DB.ConnTimeout,
			ReadTimeout:  conf.DB.ReadTimeout,
			WriteTimeout: conf.DB.WriteTimeout,
			MaxOpenConns: conf.DB.MaxOpenConns,
		}
		connInfo.CustomDriver = traceDriverName
		var err error
		dbPool, err = sqlx.NewDB(&connInfo)
		if err != nil {
			// 判断err里
			if err.Error() == "driver must implement driver.ConnBeginTx" {
				connInfo.CustomDriver = "proton-rds"
				dbPool, err = sqlx.NewDB(&connInfo)
			}
			if err != nil {
				logger.Errorf("new db operator failed; error:%s, connInfo:%+v, configLoader.DB:%+v",
					err.Error(), connInfo, conf.DB)
				panic(err)
			}
		}
	})
	return dbPool
}
