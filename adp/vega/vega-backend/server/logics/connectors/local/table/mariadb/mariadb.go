// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mariadb provides MariaDB database connector implementation.
package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

type mariadbConfig struct {
	Host      string         `mapstructure:"host"`
	Port      int            `mapstructure:"port"`
	Username  string         `mapstructure:"username"`
	Password  string         `mapstructure:"password"`
	Databases []string       `mapstructure:"databases"`
	Options   map[string]any `mapstructure:"options"`
}

var (
	SYSTEM_DBS = []string{
		"information_schema",
		"mariadb",
		"mysql",
		"performance_schema",
		"sys",
	}
	SYSTEM_DBS_MAP = map[string]bool{
		"information_schema": true,
		"mariadb":            true,
		"mysql":              true,
		"performance_schema": true,
		"sys":                true,
	}
)

const (
	// DATABASE_NAME_MAX_LENGTH MariaDB 数据库名最大长度
	DATABASE_NAME_MAX_LENGTH = 64
	// PORT_MIN 有效端口最小值
	PORT_MIN = 1
	// PORT_MAX 有效端口最大值
	PORT_MAX = 65535
)

// MariaDBConnector implements TableConnector for MariaDB.
type MariaDBConnector struct {
	enabled bool

	config *mariadbConfig

	connected bool
	db        *sql.DB
}

// NewMariaDBConnector 创建 MariaDB connector 构建器
func NewMariaDBConnector() connectors.TableConnector {
	return &MariaDBConnector{}
}

// GetType returns the data source type.
func (c *MariaDBConnector) GetType() string {
	return "mariadb"
}

// GetName returns the connector name.
func (c *MariaDBConnector) GetName() string {
	return "mariadb"
}

// GetMode returns the connector mode.
func (c *MariaDBConnector) GetMode() string {
	return interfaces.ConnectorModeLocal
}

// GetCategory returns the connector category.
func (c *MariaDBConnector) GetCategory() string {
	return interfaces.ConnectorCategoryTable
}

// GetEnabled returns the enabled status.
func (c *MariaDBConnector) GetEnabled() bool {
	return c.enabled
}

// SetEnabled sets the enabled status.
func (c *MariaDBConnector) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// GetSensitiveFields returns the sensitive fields for MariaDB connector.
func (c *MariaDBConnector) GetSensitiveFields() []string {
	return []string{"password"}
}

// GetFieldConfig returns the field configuration for MariaDB connector.
func (c *MariaDBConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return map[string]interfaces.ConnectorFieldConfig{
		"host":      {Name: "主机地址", Type: "string", Description: "数据库服务器主机地址", Required: true, Encrypted: false},
		"port":      {Name: "端口号", Type: "integer", Description: "数据库服务器端口", Required: true, Encrypted: false},
		"username":  {Name: "用户名", Type: "string", Description: "数据库用户名", Required: true, Encrypted: false},
		"password":  {Name: "密码", Type: "string", Description: "数据库密码", Required: true, Encrypted: true},
		"databases": {Name: "数据库列表", Type: "array", Description: "数据库名称列表（可选，为空则连接实例级别）", Required: false, Encrypted: false},
		"options":   {Name: "连接参数", Type: "object", Description: "连接参数（如 charset, timeout 等）", Required: false, Encrypted: false},
	}
}

// New creates a new MariaDB connector.
// Databases 为可选字段，不指定时连接到实例级别。
func (c *MariaDBConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	var mCfg mariadbConfig
	if err := mapstructure.Decode(cfg, &mCfg); err != nil {
		return nil, fmt.Errorf("failed to decode mariadb config: %w", err)
	}

	if mCfg.Host == "" || mCfg.Port == 0 || mCfg.Username == "" || mCfg.Password == "" {
		return nil, fmt.Errorf("mariadb connector config is incomplete")
	}

	// 验证端口号范围
	if mCfg.Port < PORT_MIN || mCfg.Port > PORT_MAX {
		return nil, fmt.Errorf("port %d is out of valid range (%d-%d)", mCfg.Port, PORT_MIN, PORT_MAX)
	}

	// 验证 databases 名称长度（MariaDB 数据库名最大 64 字符）
	for _, db := range mCfg.Databases {
		if len(db) > DATABASE_NAME_MAX_LENGTH {
			return nil, fmt.Errorf("database name '%s' exceeds maximum length of %d characters", db, DATABASE_NAME_MAX_LENGTH)
		}
	}

	return &MariaDBConnector{
		config: &mCfg,
	}, nil
}

// Connect establishes connection to MariaDB database.
// 如果 Config.Database 为空，则连接到实例级别（不指定数据库）。
func (c *MariaDBConnector) Connect(ctx context.Context) error {
	if c.connected {
		return nil
	}

	// Build DSN
	values := url.Values{}
	values.Set("charset", "utf8mb4")
	values.Set("parseTime", "true")

	// Apply options
	for k, v := range c.config.Options {
		values.Set(k, fmt.Sprintf("%v", v))
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?%s",
		c.config.Username, c.config.Password, c.config.Host, c.config.Port,
		values.Encode())

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	c.db = db
	c.connected = true

	return nil
}

// Close closes the database connection.
func (c *MariaDBConnector) Close(ctx context.Context) error {
	if c.db != nil {
		err := c.db.Close()
		c.connected = false
		c.db = nil
		return err
	}
	return nil
}

// Ping checks the database connection.
func (c *MariaDBConnector) Ping(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	return c.db.Ping()
}

// TestConnection tests the connection to MariaDB database.
func (c *MariaDBConnector) TestConnection(ctx context.Context) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// 如果配置了 databases 列表，验证这些数据库是否存在
	if len(c.config.Databases) > 0 {
		if err := c.validateDatabases(ctx); err != nil {
			return err
		}
	}

	return nil
}

// validateDatabases 验证配置的数据库是否存在
func (c *MariaDBConnector) validateDatabases(ctx context.Context) error {
	// 获取所有数据库列表
	rows, err := c.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	existingDBs := make(map[string]bool)
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return fmt.Errorf("failed to scan database name: %w", err)
		}
		existingDBs[dbName] = true
	}

	// 检查配置的数据库是否都存在
	var notFoundDBs []string
	for _, db := range c.config.Databases {
		if !existingDBs[db] {
			notFoundDBs = append(notFoundDBs, db)
		}
	}

	if len(notFoundDBs) > 0 {
		return fmt.Errorf("databases not found: %v", notFoundDBs)
	}

	return nil
}
