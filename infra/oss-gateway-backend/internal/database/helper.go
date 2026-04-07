package database

import (
	"database/sql"
	"oss-gateway/internal/config"
	"time"

	_ "github.com/kweaver-ai/proton-rds-sdk-go/driver"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewGorm(app *config.AppConfig, log *logrus.Entry) *gorm.DB {
	operation, err := sql.Open("proton-rds", app.DatabaseConfig.DSN())
	if err != nil {
		log.WithError(err).Fatal("init gorm failed: open dns failed")
	}

	var dial gorm.Dialector
	switch app.DatabaseConfig.TYPE {
	case "DM8":
		log.Info("using DM8 database dialect")
		dial = mysql.New(mysql.Config{Conn: operation})
	case "KDB9":
		log.Info("using KDB9 database dialect")
		dial = mysql.New(mysql.Config{Conn: operation})
	default:
		log.Info("using MySQL/MariaDB database dialect")
		dial = mysql.New(mysql.Config{Conn: operation})
	}

	db, err := gorm.Open(dial, &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("init gorm failed")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.WithError(err).Fatal("failed to get sql.DB")
	}

	sqlDB.SetMaxIdleConns(app.DatabaseConfig.MaxIdleConns)
	sqlDB.SetMaxOpenConns(app.DatabaseConfig.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(app.DatabaseConfig.ConnMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(app.DatabaseConfig.ConnMaxIdleTime) * time.Minute)

	if app.LogConfig.EnableSQL {
		log.Info("enable sql log")
		db = db.Debug()
	}

	return db
}
